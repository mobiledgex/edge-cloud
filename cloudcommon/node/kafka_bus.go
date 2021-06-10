package node

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

// Kafka credentials, put here to avoid import cyclee between node and accessapi
type KafkaCreds struct {
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type producer struct {
	producer *sarama.AsyncProducer
	address  string
}

var kafka_access_api_type = "get-kafka-creds"
var topic_prefix_operator = "operator"
var topic_prefix_developer = "developer"

// keep a list of producers and the addresses they point to so that we can still access them on a delete cloudlet
var producerLock sync.Mutex
var producers = make(map[edgeproto.CloudletKey]producer)

func (s *NodeMgr) kafkaSend(ctx context.Context, event EventData, keyTags map[string]string, keysAndValues ...string) {
	var err error
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
	region := event.Region
	// mc isnt region specific so its nodeMgr doesnt have a region, try to fill it
	if region == "" {
		region, ok = keyTags["region"]
		if !ok {
			return
		}
	}
	cloudletKey := edgeproto.CloudletKey{
		Organization: orgName,
		Name:         cloudletName,
	}
	producerLock.Lock()
	producer, ok := producers[cloudletKey]
	producerLock.Unlock()
	// check to make sure this cloudlet even has an associated kafka cluster:
	cloudlet := edgeproto.Cloudlet{}
	cloudletFound := s.CloudletLookup.GetCloudlet(region, &cloudletKey, &cloudlet)
	if !ok {
		if !cloudletFound {
			return
		} else if cloudlet.KafkaCluster == "" {
			return
		}
		producer, err = s.newProducer(ctx, region, &cloudletKey)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to create new producer", "cloudlet", cloudletKey, "error", err)
			return
		}
	} else if cloudletFound {
		// kafka cluster endpoint got changed in an updateCloudlet
		if cloudlet.KafkaCluster == "" {
			removeProducer(&cloudletKey)
			return
		}
		if producer.address != cloudlet.KafkaCluster {
			producer, err = s.newProducer(ctx, region, &cloudletKey)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Failed to create new producer", "cloudlet", cloudletKey, "error", err)
				return
			}
		}
	}
	//split the events into two main topics, "operator events" and "developer events"
	// if there are other orgs tagged besides the operator org and "mobiledgex", its a dev event
	// TODO: add a third prefix (mobiledgex) for events that we do?
	allowed := false
	topic := topic_prefix_operator + "-" + cloudletName
	for _, eventorg := range event.Org {
		if eventorg == orgName {
			allowed = true
		}
		if eventorg != orgName && eventorg != cloudcommon.OrganizationMobiledgeX {
			topic = topic_prefix_developer + "-" + cloudletName
		}
	}
	// make sure the operator has permission to view this event
	if !allowed {
		return
	}

	message := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(buildMessageBody(event)),
		Timestamp: event.Timestamp,
	}
	go s.sendMessage(ctx, producer.producer, message, &cloudletKey, region)
}

func (s *NodeMgr) sendMessage(ctx context.Context, producer *sarama.AsyncProducer, message *sarama.ProducerMessage, cloudletKey *edgeproto.CloudletKey, region string) {
	(*producer).Input() <- message
	for {
		select {
		case <-(*producer).Errors():
			// repull the credentials from vault and try it again in case the creds got changed in an UpdateCloudlet
			newProducer, err := s.newProducer(ctx, region, cloudletKey)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Failed to create new producer", "cloudlet", cloudletKey, "error", err)
				return
			}
			(*newProducer.producer).Input() <- message
			for {
				select {
				case err := <-(*newProducer.producer).Errors():
					log.SpanLog(ctx, log.DebugLevelInfo, "Failed to send event to operator kafka cluster", "cloudletKey", cloudletKey, "message", message, "error", err)
				case <-(*newProducer.producer).Successes():
					return
				}
			}
		case <-(*producer).Successes():
			return
		}
	}
}

func (s *NodeMgr) newProducer(ctx context.Context, region string, key *edgeproto.CloudletKey) (producer, error) {
	kafkaCreds := KafkaCreds{}
	// if youre connected to controller, go through controller
	if s.MyNode.Key.Type == NodeTypeCRM || s.AccessKeyClient.enabled {
		req := &edgeproto.AccessDataRequest{
			Type: kafka_access_api_type,
		}
		reply, err := s.AccessApiClient.GetAccessData(ctx, req)
		if err != nil {
			return producer{}, err
		}
		err = json.Unmarshal(reply.Data, &kafkaCreds)
		if err != nil {
			return producer{}, err
		}
	} else {
		path := GetKafkaVaultPath(region, key.Name, key.Organization)
		err := vault.GetData(s.VaultConfig, path, 0, &kafkaCreds)
		if err != nil {
			return producer{}, fmt.Errorf("Error pulling kafka credentials from vault: %v", err)
		}
	}
	config := sarama.NewConfig()
	config.ClientID = s.MyNode.Key.Type
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	// always use SSL encryption
	config.Net.TLS.Enable = true
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return producer{}, fmt.Errorf("Unable to get system certs")
	}
	newConfig := tls.Config{RootCAs: rootCAs}
	if s.unitTestMode {
		newConfig.InsecureSkipVerify = true
	}
	config.Net.TLS.Config = &newConfig
	// parameters for SASL/plain authentification with kafka clusters
	if kafkaCreds.Username != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = kafkaCreds.Username
		config.Net.SASL.Password = kafkaCreds.Password
	}
	newProducer, err := sarama.NewAsyncProducer([]string{kafkaCreds.Endpoint}, config)
	if err != nil {
		return producer{}, fmt.Errorf("Error creating producer: %v", err)
	}
	producer := producer{
		producer: &newProducer,
		address:  kafkaCreds.Endpoint,
	}
	// close the current producer (if there is one) before putting in this one
	producerLock.Lock()
	oldProducer, ok := producers[*key]
	if ok {
		(*oldProducer.producer).AsyncClose()
	}
	producers[*key] = producer
	producerLock.Unlock()
	return producer, nil
}

func removeProducer(key *edgeproto.CloudletKey) {
	producerLock.Lock()
	defer producerLock.Unlock()
	producer, ok := producers[*key]
	if ok {
		(*producer.producer).AsyncClose()
	}
	delete(producers, *key)
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

func GetKafkaVaultPath(region, cloudletName, org string) string {
	return fmt.Sprintf("secret/data/kafka/%s/%s/%s", region, org, cloudletName)
}
