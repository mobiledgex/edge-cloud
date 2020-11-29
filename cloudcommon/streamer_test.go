package cloudcommon

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStreamer(t *testing.T) {
	var wg sync.WaitGroup

	streamer := NewStreamer()

	// Msgs to be published to all subscribers
	sendMsgs := []int{}
	for id := 0; id < 10; id++ {
		sendMsgs = append(sendMsgs, id)
	}

	// Subscriber
	subscriberFunc := func(wg *sync.WaitGroup, id int) {
		defer wg.Done()
		streamCh := streamer.Subscribe()
		rcvdMsgs := []int{}
		for streamMsg := range streamCh {
			rcvdMsgs = append(rcvdMsgs, streamMsg.(int))
		}
		require.Equal(t, sendMsgs, rcvdMsgs, fmt.Sprintf("Client %d: match received msgs", id))
	}

	// Create multiple subscribers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go subscriberFunc(&wg, i)
	}

	// Start publishing messages
	go func() {
		for _, msg := range sendMsgs {
			streamer.Publish(msg)
			time.Sleep(1 * time.Millisecond)
		}
		streamer.Stop()
	}()

	// Create some more subscribers started after a while
	for i := 3; i < 5; i++ {
		time.Sleep(2 * time.Millisecond)
		wg.Add(1)
		go subscriberFunc(&wg, i)
	}

	// Create another subscriber, who unsubscribes in between
	wg.Add(1)
	go func() {
		defer wg.Done()
		id := 5
		streamCh := streamer.Subscribe()
		rcvdMsgs := []int{}
		for streamMsg := range streamCh {
			rcvdMsg := streamMsg.(int)
			if rcvdMsg == 5 {
				streamer.Unsubscribe(streamCh)
				break
			}
			rcvdMsgs = append(rcvdMsgs, rcvdMsg)
		}
		require.NotEqual(t, sendMsgs, rcvdMsgs, fmt.Sprintf("Client %d: shoud not receive all the msgs", id))
		require.Equal(t, 5, len(rcvdMsgs), fmt.Sprintf("Client %d: match received msgs", id))
	}()

	// Wait for all subscribers to finish reading streamed messages
	wg.Wait()
}

type TestKey struct {
	id   int
	name string
}

var streamTest = &StreamObj{}

func TestStreamMaps(t *testing.T) {
	var wg sync.WaitGroup
	testKey1 := TestKey{id: 1, name: "testKey1"}
	testKey2 := TestKey{id: 2, name: "testKey2"}
	// Msgs to be published to all subscribers
	sendMsgs1 := []int{}
	for id := 0; id < 5; id++ {
		sendMsgs1 = append(sendMsgs1, id)
	}
	sendMsgs2 := []int{}
	for id := 5; id < 10; id++ {
		sendMsgs2 = append(sendMsgs2, id)
	}

	publishMsgs := func(key *TestKey, sendMsgs []int, streamAdded chan bool) {
		// Start publishing messages
		streamer := NewStreamer()
		streamTest.Add(*key, streamer)
		streamAdded <- true
		defer streamer.Stop()
		for _, msg := range sendMsgs {
			streamer.Publish(msg)
			time.Sleep(2 * time.Millisecond)
		}
		streamTest.Remove(*key, streamer)
	}

	// Start publishing messages
	streamAdded := make(chan bool)
	go publishMsgs(&testKey1, sendMsgs1, streamAdded)

	// Subscriber
	subscriberFunc := func(wg *sync.WaitGroup, key interface{}, id int, sendMsgs []int) {
		if wg != nil {
			defer wg.Done()
		}
		streamer := streamTest.Get(key)
		require.NotNil(t, streamer, fmt.Sprintf("Client %d: stream exists", id))
		streamCh := streamer.Subscribe()
		rcvdMsgs := []int{}
		for streamMsg := range streamCh {
			rcvdMsgs = append(rcvdMsgs, streamMsg.(int))
		}
		streamer.Unsubscribe(streamCh)
		require.Equal(t, sendMsgs, rcvdMsgs, fmt.Sprintf("Client %d: match received msgs", id))
	}

	// Wait for publisher to start streaming msgs
	if <-streamAdded {
		subscriberFunc(nil, testKey1, 0, sendMsgs1)
	}

	// Start another publisher, publishing new msgs on same key
	streamAdded = make(chan bool)
	go publishMsgs(&testKey1, sendMsgs2, streamAdded)

	// Wait for publisher to start streaming msgs
	if <-streamAdded {
		require.Equal(t, 1, len((*streamTest).streamMap), "One publisher exists")
		for i := 1; i < 3; i++ {
			time.Sleep(time.Duration(i) * time.Millisecond)
			wg.Add(1)
			go subscriberFunc(&wg, testKey1, i, sendMsgs2)
		}
	}

	// Start another publisher for different key
	streamAdded = make(chan bool)
	go publishMsgs(&testKey2, sendMsgs2, streamAdded)

	if <-streamAdded {
		require.Equal(t, 2, len((*streamTest).streamMap), "Two publisher exists")
		// Create multiple subscribers started with some time gap
		for i := 0; i < 3; i++ {
			time.Sleep(time.Duration(i) * time.Millisecond)
			wg.Add(1)
			go subscriberFunc(&wg, testKey2, i, sendMsgs2)
		}
	}

	// Wait for all subscribers to finish reading streamed messages
	wg.Wait()

	// Verify no streamObjs exists
	require.Equal(t, 0, len((*streamTest).streamMap), "No publishers exists")
}
