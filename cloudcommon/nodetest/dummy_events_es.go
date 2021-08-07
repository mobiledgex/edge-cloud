package nodetest

import (
	"encoding/json"
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
)

// This captures events meant to be sent to elastic search.
// This allows verification of events/audit logs in unit-tests,
// without needing to run elastic search.
// This does not capture jaeger logs.
type DummyEventsES struct {
	Events []*node.EventData
	Mux    sync.Mutex
}

// This assumes httpmock has already been initialized via:
// httpmock.Activate()
// defer httpmock.DeactiveAndReset()
func (s *DummyEventsES) InitHttpMock(addr string) {
	s.Events = make([]*node.EventData, 0)

	matchAll := "=~" + addr + `/.*\z`
	// regexp match POST events
	httpmock.RegisterResponder("POST", matchAll, s.Handle)
	// ignore searches
	httpmock.RegisterResponder("GET", matchAll, s.HandleIgnore)
	// ignore PUT index template
	httpmock.RegisterResponder("PUT", matchAll, s.HandleIgnore)
}

func (s *DummyEventsES) Handle(req *http.Request) (*http.Response, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
		return httpmock.NewStringResponse(400, "failed to read body"), nil
	}
	// data is {{index{}}}\ndata\n
	data := strings.Split(string(body), "\n")
	for ii := 0; ii < len(data); ii++ {
		// ignore first line, it's {index:{}}
		ii++
		if ii >= len(data) {
			continue
		}
		event := node.EventData{}
		err = json.Unmarshal([]byte(data[ii]), &event)
		if err != nil {
			fmt.Printf("failed to unmarshal data %s: %v\n", data[ii], err)
			continue
		}
		fmt.Printf("Received event %s type %s\n", event.Name, event.Type)
		s.Mux.Lock()
		s.Events = append(s.Events, &event)
		s.Mux.Unlock()
	}
	return httpmock.NewStringResponse(200, ""), nil
}

func (s *DummyEventsES) HandleIgnore(req *http.Request) (*http.Response, error) {
	fmt.Printf("DummyEventsES ignoring request %s\n", req.URL.String())
	return httpmock.NewStringResponse(200, `{"hits":{"total":{"value":0}}}`), nil
}

func (s *DummyEventsES) WaitLastEventMatches(matchFunc func(e *node.EventData) bool) bool {
	matches := false
	for ii := 0; ii < 20; ii++ {
		s.Mux.Lock()
		if len(s.Events) == 0 {
			s.Mux.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		matches = matchFunc(s.Events[len(s.Events)-1])
		s.Mux.Unlock()
		if matches {
			return matches
		}
		time.Sleep(100 * time.Millisecond)
	}
	return matches
}
