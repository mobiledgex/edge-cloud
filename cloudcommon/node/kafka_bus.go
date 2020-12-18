package node

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var topic_prefix_operator = "operator"
var topic_prefix_developer = "developer"

// keep a list of producers and the addresses they point to so that we can still access them on a delete cloudlet
var producers = make(map[edgeproto.CloudletKey]*sarama.SyncProducer)
var kafkaEndpoints = make(map[edgeproto.CloudletKey]string)

func (s *NodeMgr) kafkaSend(ctx context.Context, event EventData, keyTags map[string]string, keysAndValues ...string) {
	fmt.Printf("asdf made it to kafkasend: %+v\n", event)
	orgName, ok := keyTags["cloudletorg"]
	if !ok {
		if orgName, ok = keyTags["org"]; !ok {
			return
		}
	}
	cloudletName, ok := keyTags["cloudlet"]
	if !ok {
		return
	}
	fmt.Printf("1\n")
	cloudletKey := edgeproto.CloudletKey{
		Organization: orgName,
		Name:         cloudletName,
	}
	cloudlet := edgeproto.Cloudlet{}
	var producer *sarama.SyncProducer
	if !s.CloudletLookup.Get(&cloudletKey, &cloudlet) {
		if event.Name == "cloudlet deleted" { // TODO: figure out the actual cloudlet deleted event name
			producer, ok = producers[cloudletKey]
			if !ok { // the cloudlet didnt have a kafka endpoint specified in the first place
				return
			}
		} else {
			log.SpanLog(ctx, log.DebugLevelInfo, "Cloudlet Details specified but no cloudlet found", "cloudletkey", cloudletKey)
			////////////////////////////////////////////////////////////////////////////////////////////////////
			// TODO: remove this later, this is hardcoded for demo purposes so mc wont complain
			fmt.Printf("NO CLOUDLET FOUND, GOING TO HARDCODED ADDRESS\n")
			kafkaClusterEndpoint := "localhost:9092"
			endpoint := kafkaEndpoints[cloudletKey]
			if kafkaClusterEndpoint != endpoint {
				//either this is the first time we're seeing this cloudlet or the endpoint got changed in an UpdateCloudlet
				config := sarama.NewConfig()
				config.Producer.Return.Successes = true
				config.Producer.Return.Errors = true
				newProducer, err := sarama.NewSyncProducer([]string{kafkaClusterEndpoint}, config)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfo, "Error creating producer", "cloudlet", cloudletKey, "kafkaEndpoint", kafkaClusterEndpoint, "err", err)
					return
				}
				producer = &newProducer
				// close the current producer before putting in this one
				removeProducer(&cloudletKey)
				producers[cloudletKey] = producer
				kafkaEndpoints[cloudletKey] = kafkaClusterEndpoint
			} else {
				producer = producers[cloudletKey]
			}
			//END OF REMOVE BLOCK
			////////////////////////////////////////////////////////////////////////////////////////////////////////
			// return
		}
		fmt.Printf("2\n")
	} else { // found cloudlet detail
		kafkaClusterEndpoint := cloudlet.KafkaCluster
		// if the operator didnt specify a kafka cluster to push events to
		if kafkaClusterEndpoint == "" {
			removeProducer(&cloudletKey)
			return
		}
		fmt.Printf("3\n")
		endpoint := kafkaEndpoints[cloudletKey]
		if kafkaClusterEndpoint != endpoint {
			//either this is the first time we're seeing this cloudlet or the endpoint got changed in an UpdateCloudlet
			config := sarama.NewConfig()
			config.Producer.Return.Successes = true
			config.Producer.Return.Errors = true
			newProducer, err := sarama.NewSyncProducer([]string{kafkaClusterEndpoint}, config)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Error creating producer", "cloudlet", cloudletKey, "kafkaEndpoint", kafkaClusterEndpoint, "err", err)
				return
			}
			producer = &newProducer
			// close the current producer before putting in this one
			removeProducer(&cloudletKey)
			producers[cloudletKey] = producer
			kafkaEndpoints[cloudletKey] = kafkaClusterEndpoint
		} else {
			producer = producers[cloudletKey]
		}
	}
	fmt.Printf("4\n")

	// TODO: CHANGE THIS WHEN JON'S CHANGES FOR PRETAGGING EVENTS COME IN (EC-3465)
	// make sure the operator has permission to view this event
	// if event.Org != orgName {
	// 	fmt.Printf("org names mismatch: event.Org: %s, orgName: %s\n", event.Org, orgName)
	// 	return
	// }

	//split the events into two main topics, "operator events" and "developer events"
	//TODO: turns this into a "contains"
	// if there are other orgs tagged besides the operator org and "mobiledgex", its a dev event
	// TODO: add a third prefix for events that we do?
	topic := topic_prefix_operator + "-" + cloudletName
	if event.Org != orgName && event.Org != cloudcommon.OrganizationMobiledgeX {
		topic = topic_prefix_developer + "-" + cloudletName
	}

	message := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(buildMessageBody(event)),
		Timestamp: event.Timestamp,
	}
	fmt.Printf("sending message: %+v\n", message)

	_, _, err := (*producer).SendMessage(message)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to send event to operator kafka cluster", "cloudletKey", cloudletKey, "message", message)
	}
	fmt.Printf("done sending\n")
}

func removeProducer(key *edgeproto.CloudletKey) {
	producer, ok := producers[*key]
	if ok {
		(*producer).Close()
	}
	delete(producers, *key)
	delete(kafkaEndpoints, *key)
}

func buildMessageBody(event EventData) string {
	message := event.Name
	for _, tag := range event.Tags {
		message = message + "\n"
		message = message + tag.Key + ":" + tag.Value
	}
	message = message + "\n\n"
	return message
}
