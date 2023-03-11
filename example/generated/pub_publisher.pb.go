// Code generated  by protoc-gen-go-event. DO NOT EDIT.
// versions:
// source: pub.proto

package example

import (
	"context"
	"cloud.google.com/go/pubsub"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type HelloWorldServiceClient interface {
	PublishHelloWorld(ctx context.Context, req *HelloWorldRequest) (string, error)
}

type innerHelloWorldServiceClient struct {
	client *pubsub.Client
}

func NewHelloWorldServiceClient(client *pubsub.Client) *innerHelloWorldServiceClient {
	return &innerHelloWorldServiceClient{
		client: client,
	}
}

var topicCache = map[string]*pubsub.Topic{}

func (c *innerHelloWorldServiceClient) getTopic(topic string) (*pubsub.Topic, error) {
	if t, ok := topicCache[topic]; ok {
		return t, nil
	}
	t, err := GetOrCreateTopicIfNotExists(c.client, topic)
	if err != nil {
		return nil, err
	}
	topicCache[topic] = t
	return t, nil
}

func (c *innerHelloWorldServiceClient) publish(topic string, event protoreflect.ProtoMessage) (string, error) {
	ctx := context.Background()

	t, err := c.getTopic(topic)
	if err != nil {
		return "", err
	}

	ev, err := proto.Marshal(event)
	if err != nil {
		return "", err
	}

	result := t.Publish(ctx, &pubsub.Message{
		Data: ev,
	})
	id, err := result.Get(ctx)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (c *innerHelloWorldServiceClient) PublishHelloWorld(ctx context.Context, req *HelloWorldRequest) (string, error) {
	return c.publish("helloworldtopic", req)
}

// GetOrCreateTopicIfNotExists: topicが存在しない場合は作成する
func GetOrCreateTopicIfNotExists(c *pubsub.Client, topic string) (*pubsub.Topic, error) {
	ctx := context.Background()
	t := c.Topic(topic)
	ok, err := t.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		return t, nil
	}
	t, err = c.CreateTopic(ctx, topic)
	if err != nil {
		return nil, err
	}
	return t, nil
}
