// Code generated  by protoc-gen-go-event. DO NOT EDIT.
// versions:
// - protoc-gen-go-pubsub v1.0.0
// - protoc             v3.21.12
// source: sub.proto

package example

import (
	"context"
	"fmt"
	"time"
	"cloud.google.com/go/pubsub"
	gosub "github.com/TakaWakiyama/protoc-gen-go-pubsub/subscriber"
	retry "github.com/avast/retry-go"
	"google.golang.org/protobuf/proto"
)

// SubscriberOption is the option for HelloWorldSubscriber
type SubscriberOption struct {
	// Interceptors is the slice of SubscriberInterceptor. call before and after HelloWorldSubscriber method. default is empty.
	Interceptors []gosub.SubscriberInterceptor
	// SubscribeGracefully is the flag to stop subscribing gracefully. default is false.
	SubscribeGracefully bool
}

var defaultSubscriberOption = &SubscriberOption{
	Interceptors:        []gosub.SubscriberInterceptor{},
	SubscribeGracefully: false,
}

type HelloWorldSubscriber interface {
	// Hello world is super method
	HelloWorld(ctx context.Context, req *HelloWorldRequest) error
}

func Run(service HelloWorldSubscriber, client *pubsub.Client, option *SubscriberOption) error {
	if service == nil {
		return fmt.Errorf("service is nil")
	}
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	if option == nil {
		option = defaultSubscriberOption
	}
	ctx := context.Background()
	is := newInnerHelloWorldSubscriberSubscriber(service, client, option)
	if err := is.listenHelloWorld(ctx); err != nil {
		return err
	}
	return nil
}

type innerHelloWorldSubscriberSubscriber struct {
	service  HelloWorldSubscriber
	client   *pubsub.Client
	option   SubscriberOption
	accessor PubSubAccessor
}

func newInnerHelloWorldSubscriberSubscriber(service HelloWorldSubscriber, client *pubsub.Client, option *SubscriberOption) *innerHelloWorldSubscriberSubscriber {
	if option == nil {
		option = defaultSubscriberOption
	}
	return &innerHelloWorldSubscriberSubscriber{
		service:  service,
		client:   client,
		option:   *option,
		accessor: NewPubSubAccessor(),
	}
}

func (is *innerHelloWorldSubscriberSubscriber) listenHelloWorld(ctx context.Context) error {
	subscriptionName := "helloworldsubscription"
	topicName := "helloworldtopic"
	var sub *pubsub.Subscription
	if err := retry.Do(func() error {
		tmp, err := is.accessor.CreateHelloWorldSubscriptionIFNotExists(ctx, is.client)
		if err != nil {
			return err
		}
		sub = tmp
		return nil
	}); err != nil {
		return err
	}
	callback := func(ctx context.Context, msg *pubsub.Message) {
		info := gosub.NewSubscriberInfo(topicName, subscriptionName, sub, "HelloWorld", msg)
		var event HelloWorldRequest
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		if err := gosub.Handle(is.option.Interceptors, ctx, &event, info, func(ctx context.Context, req interface{}) error {
			return is.service.HelloWorld(ctx, req.(*HelloWorldRequest))
		}); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	}

	if is.option.SubscribeGracefully {
		gosub.SubscribeGracefully(sub, ctx, callback, nil)
	} else {
		sub.Receive(ctx, callback)
	}

	return nil
}

// PubSubAccessor: accessor for HelloWorldPubSub
type PubSubAccessor interface {
	CreateHelloWorldTopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error)
	CreateHelloWorldSubscriptionIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Subscription, error)
}

type pubSubAccessorImpl struct{}

func NewPubSubAccessor() PubSubAccessor {
	return &pubSubAccessorImpl{}
}

func (c *pubSubAccessorImpl) CreateHelloWorldTopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error) {
	topicName := "helloworldtopic"
	t := client.Topic(topicName)
	exsits, err := t.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check topic exsits: %w", err)
	}
	if !exsits {
		return client.CreateTopic(ctx, topicName)
	}
	return t, nil
}

func (c *pubSubAccessorImpl) CreateHelloWorldSubscriptionIFNotExists(
	ctx context.Context,
	client *pubsub.Client,
) (*pubsub.Subscription, error) {
	subscriptionName := "helloworldsubscription"
	topicName := "helloworldtopic"
	t := client.Topic(topicName)
	if exsits, err := t.Exists(ctx); !exsits {
		return nil, fmt.Errorf("topic does not exsit: %w", err)
	}
	sub := client.Subscription(subscriptionName)
	exsits, err := sub.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check subscription exsits: %w", err)
	}
	if !exsits {
		return client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic:       t,
			AckDeadline: 60 * time.Second,
		})
	}
	return sub, nil
}
