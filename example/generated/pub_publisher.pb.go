// Code generated  by protoc-gen-go-event. DO NOT EDIT.
// versions:
// - protoc-gen-go-pubsub v1.0.0
// - protoc             v3.21.12
// source: pub.proto

package example

import (
	"context"
	"time"
	"cloud.google.com/go/pubsub"
	gopub "github.com/TakaWakiyama/protoc-gen-go-pubsub/publisher"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type BatchPublishResult struct {
	ID    string
	Error error
}

// ClientOption is the option for publisher client
type ClientOption struct {
	// Gracefully is the flag to stop publishing gracefully
	Gracefully bool
	// MaxAttempts is the max attempts when wait for publishing gracefully
	MaxAttempts int
	// Delay is the delay time when wait for publishing gracefully
	Delay time.Duration
}

var defaultClientOption = &ClientOption{
	Gracefully:  false,
	MaxAttempts: 3,
	Delay:       1 * time.Second,
}

type ExamplePublisherClient interface {
	PublishHelloWorld(ctx context.Context, req *HelloWorldEvent) (string, error)
	PublishHogeCreated(ctx context.Context, req *HogeEvent) (string, error)
	PublishHogesCreated(ctx context.Context, req *HogeEvent) (string, error)
}

type innerExamplePublisherClient struct {
	client          *pubsub.Client
	nameToTopic     map[string]*pubsub.Topic
	nameToPublisher map[string]gopub.Publisher
	option          ClientOption
}

func NewExamplePublisherClient(client *pubsub.Client, option *ClientOption) *innerExamplePublisherClient {
	if option == nil {
		option = defaultClientOption
	}
	return &innerExamplePublisherClient{
		client:          client,
		nameToTopic:     make(map[string]*pubsub.Topic),
		nameToPublisher: make(map[string]gopub.Publisher),
		option:          *option,
	}
}

func (c *innerExamplePublisherClient) getTopic(topicName string) (*pubsub.Topic, error) {
	if t, ok := c.nameToTopic[topicName]; ok {
		return t, nil
	}
	t, err := gopub.GetOrCreateTopicIfNotExists(c.client, topicName)
	if err != nil {
		return nil, err
	}
	c.nameToTopic[topicName] = t
	return t, nil
}

func (c *innerExamplePublisherClient) getPublisher(topicName string) (gopub.Publisher, error) {
	if p, ok := c.nameToPublisher[topicName]; ok {
		return p, nil
	}
	p := gopub.NewPublisher(c.client, &gopub.PublisherOption{
		Gracefully:  c.option.Gracefully,
		MaxAttempts: uint(c.option.MaxAttempts),
		Delay:       c.option.Delay,
	})
	c.nameToPublisher[topicName] = p
	return p, nil
}

func (c *innerExamplePublisherClient) publish(topic string, event protoreflect.ProtoMessage) (string, error) {
	ctx := context.Background()

	t, err := c.getTopic(topic)
	if err != nil {
		return "", err
	}
	p, err := c.getPublisher(topic)
	if err != nil {
		return "", err
	}

	ev, err := proto.Marshal(event)
	if err != nil {
		return "", err
	}
	return p.Publish(ctx, t, &pubsub.Message{
		Data: ev,
	})
}

func (c *innerExamplePublisherClient) batchPublish(topic string, events []protoreflect.ProtoMessage) ([]BatchPublishResult, error) {
	ctx := context.Background()

	t, err := c.getTopic(topic)
	if err != nil {
		return nil, err
	}
	p, err := c.getPublisher(topic)
	if err != nil {
		return nil, err
	}

	var msgs []*pubsub.Message
	for _, e := range events {
		ev, err := proto.Marshal(e)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &pubsub.Message{
			Data: ev,
		})
	}
	res, err := p.BatchPublish(ctx, t, msgs)
	if err != nil {
		return nil, err
	}
	var results []BatchPublishResult
	for _, r := range res {
		results = append(results, BatchPublishResult{
			ID:    r.ID,
			Error: r.Error,
		})
	}
	return results, nil
}

func (c *innerExamplePublisherClient) PublishHelloWorld(ctx context.Context, req *HelloWorldEvent) (string, error) {
	return c.publish("helloworldtopic", req)
}
func (c *innerExamplePublisherClient) PublishHogeCreated(ctx context.Context, req *HogeEvent) (string, error) {
	return c.publish("hogeCreated", req)
}
func (c *innerExamplePublisherClient) BatchPublishHogesCreated(ctx context.Context, req []*HogeEvent) ([]BatchPublishResult, error) {
	events := make([]protoreflect.ProtoMessage, len(req))
	for i, r := range req {
		events[i] = r
	}
	return c.batchPublish("hogeCreated", events)
}
