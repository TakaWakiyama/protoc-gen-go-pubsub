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

// SubscriberOption is the option for subscriber.
type SubscriberOption struct {
	// Interceptors is the slice of SubscriberInterceptor. call before and after subscriber method. default is empty.
	Interceptors []gosub.SubscriberInterceptor
	// SubscribeGracefully is the flag to stop subscribing gracefully. default is false.
	SubscribeGracefully bool
}

var defaultSubscriberOption = &SubscriberOption{
	Interceptors:        []gosub.SubscriberInterceptor{},
	SubscribeGracefully: true,
}

var retryOpts = []retry.Option{
	retry.Delay(1 * time.Second),
	retry.Attempts(3),
}

type ExampleSubscriber interface {
	// Hello world is super method
	HelloWorld(ctx context.Context, req *HelloWorldEvent) error
	OnHoge(ctx context.Context, req *HogeEvent) error
}

func RunExampleSubscriber(service ExampleSubscriber, client *pubsub.Client, option *SubscriberOption) error {
	if service == nil {
		return fmt.Errorf("service is nil")
	}
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	if option == nil {
		option = defaultSubscriberOption
	}
	is := newInnerExampleSubscriber(service, client, option)
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)

	go func() {
		if err := is.listenHelloWorld(ctx); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := is.listenOnHoge(ctx); err != nil {
			errChan <- err
		}
	}()
	err := <-errChan
	cancel()
	return err
}

type innerExampleSubscriber struct {
	service  ExampleSubscriber
	client   *pubsub.Client
	option   SubscriberOption
	accessor ExampleSubscriberPubSubAccessor
}

func newInnerExampleSubscriber(service ExampleSubscriber, client *pubsub.Client, option *SubscriberOption) *innerExampleSubscriber {
	if option == nil {
		option = defaultSubscriberOption
	}
	return &innerExampleSubscriber{
		service:  service,
		client:   client,
		option:   *option,
		accessor: NewExampleSubscriberPubSubAccessor(),
	}
}

func (is *innerExampleSubscriber) listenHelloWorld(ctx context.Context) error {
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
	}, retryOpts...); err != nil {
		return err
	}

	// Note: you can see ReceiveSettings https://github.com/googleapis/google-cloud-go/blob/pubsub/v1.31.0/pubsub/subscription.go#L720
	// Note: you can see default value  https://github.com/googleapis/google-cloud-go/blob/main/pubsub/subscription.go#L743
	sub.ReceiveSettings.Synchronous = false
	sub.ReceiveSettings.NumGoroutines = 10
	sub.ReceiveSettings.MaxOutstandingMessages = 1000

	callback := func(ctx context.Context, msg *pubsub.Message) {
		defer func() {
			if err := recover(); err != nil {
				msg.Nack()
			}
		}()
		info := gosub.NewSubscriberInfo(topicName, subscriptionName, sub, "HelloWorld", msg)
		var event HelloWorldEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		if err := gosub.Handle(is.option.Interceptors, ctx, &event, info, func(ctx context.Context, req interface{}) error {
			return is.service.HelloWorld(ctx, req.(*HelloWorldEvent))
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

func (is *innerExampleSubscriber) listenOnHoge(ctx context.Context) error {
	subscriptionName := "onHogeCreated"
	topicName := "hogeCreated"
	var sub *pubsub.Subscription
	if err := retry.Do(func() error {
		tmp, err := is.accessor.CreateOnHogeSubscriptionIFNotExists(ctx, is.client)
		if err != nil {
			return err
		}
		sub = tmp
		return nil
	}, retryOpts...); err != nil {
		return err
	}

	// Note: you can see ReceiveSettings https://github.com/googleapis/google-cloud-go/blob/pubsub/v1.31.0/pubsub/subscription.go#L720
	// Note: you can see default value  https://github.com/googleapis/google-cloud-go/blob/main/pubsub/subscription.go#L743
	sub.ReceiveSettings.Synchronous = false
	sub.ReceiveSettings.NumGoroutines = 10
	sub.ReceiveSettings.MaxOutstandingMessages = 1000

	callback := func(ctx context.Context, msg *pubsub.Message) {
		defer func() {
			if err := recover(); err != nil {
				msg.Nack()
			}
		}()
		info := gosub.NewSubscriberInfo(topicName, subscriptionName, sub, "OnHoge", msg)
		var event HogeEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		if err := gosub.Handle(is.option.Interceptors, ctx, &event, info, func(ctx context.Context, req interface{}) error {
			return is.service.OnHoge(ctx, req.(*HogeEvent))
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

// ExampleSubscriberPubSubAccessor: Interface for create topic and subscription
type ExampleSubscriberPubSubAccessor interface {
	CreateHelloWorldTopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error)
	CreateHelloWorldSubscriptionIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Subscription, error)

	CreateOnHogeTopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error)
	CreateOnHogeSubscriptionIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Subscription, error)
}

type innerExampleSubscriberPubSubAccessor struct{}

func NewExampleSubscriberPubSubAccessor() ExampleSubscriberPubSubAccessor {
	return &innerExampleSubscriberPubSubAccessor{}
}

func (c *innerExampleSubscriberPubSubAccessor) CreateHelloWorldTopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error) {
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

func (c *innerExampleSubscriberPubSubAccessor) CreateHelloWorldSubscriptionIFNotExists(
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
			Topic:                     t,
			AckDeadline:               60 * time.Second,
			EnableExactlyOnceDelivery: false,
		})
	}
	return sub, nil
}

func (c *innerExampleSubscriberPubSubAccessor) CreateOnHogeTopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error) {
	topicName := "hogeCreated"
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

func (c *innerExampleSubscriberPubSubAccessor) CreateOnHogeSubscriptionIFNotExists(
	ctx context.Context,
	client *pubsub.Client,
) (*pubsub.Subscription, error) {
	subscriptionName := "onHogeCreated"
	topicName := "hogeCreated"
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
			Topic:                     t,
			AckDeadline:               30 * time.Second,
			EnableExactlyOnceDelivery: false,
		})
	}
	return sub, nil
}

type Example2Subscriber interface {
	// Hello world is super method
	HelloWorld2(ctx context.Context, req *HelloWorldEvent) error
	OnHoge2(ctx context.Context, req *HogeEvent) error
}

func RunExample2Subscriber(service Example2Subscriber, client *pubsub.Client, option *SubscriberOption) error {
	if service == nil {
		return fmt.Errorf("service is nil")
	}
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	if option == nil {
		option = defaultSubscriberOption
	}
	is := newInnerExample2Subscriber(service, client, option)
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)

	go func() {
		if err := is.listenHelloWorld2(ctx); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := is.listenOnHoge2(ctx); err != nil {
			errChan <- err
		}
	}()
	err := <-errChan
	cancel()
	return err
}

type innerExample2Subscriber struct {
	service  Example2Subscriber
	client   *pubsub.Client
	option   SubscriberOption
	accessor Example2SubscriberPubSubAccessor
}

func newInnerExample2Subscriber(service Example2Subscriber, client *pubsub.Client, option *SubscriberOption) *innerExample2Subscriber {
	if option == nil {
		option = defaultSubscriberOption
	}
	return &innerExample2Subscriber{
		service:  service,
		client:   client,
		option:   *option,
		accessor: NewExample2SubscriberPubSubAccessor(),
	}
}

func (is *innerExample2Subscriber) listenHelloWorld2(ctx context.Context) error {
	subscriptionName := "helloworldsubscription"
	topicName := "helloworldtopic"
	var sub *pubsub.Subscription
	if err := retry.Do(func() error {
		tmp, err := is.accessor.CreateHelloWorld2SubscriptionIFNotExists(ctx, is.client)
		if err != nil {
			return err
		}
		sub = tmp
		return nil
	}, retryOpts...); err != nil {
		return err
	}

	// Note: you can see ReceiveSettings https://github.com/googleapis/google-cloud-go/blob/pubsub/v1.31.0/pubsub/subscription.go#L720
	// Note: you can see default value  https://github.com/googleapis/google-cloud-go/blob/main/pubsub/subscription.go#L743
	sub.ReceiveSettings.Synchronous = true
	sub.ReceiveSettings.NumGoroutines = 1
	sub.ReceiveSettings.MaxOutstandingMessages = -1

	callback := func(ctx context.Context, msg *pubsub.Message) {
		defer func() {
			if err := recover(); err != nil {
				msg.Nack()
			}
		}()
		info := gosub.NewSubscriberInfo(topicName, subscriptionName, sub, "HelloWorld2", msg)
		var event HelloWorldEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		if err := gosub.Handle(is.option.Interceptors, ctx, &event, info, func(ctx context.Context, req interface{}) error {
			return is.service.HelloWorld2(ctx, req.(*HelloWorldEvent))
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

func (is *innerExample2Subscriber) listenOnHoge2(ctx context.Context) error {
	subscriptionName := "onHogeCreated"
	topicName := "hogeCreated"
	var sub *pubsub.Subscription
	if err := retry.Do(func() error {
		tmp, err := is.accessor.CreateOnHoge2SubscriptionIFNotExists(ctx, is.client)
		if err != nil {
			return err
		}
		sub = tmp
		return nil
	}, retryOpts...); err != nil {
		return err
	}

	// Note: you can see ReceiveSettings https://github.com/googleapis/google-cloud-go/blob/pubsub/v1.31.0/pubsub/subscription.go#L720
	// Note: you can see default value  https://github.com/googleapis/google-cloud-go/blob/main/pubsub/subscription.go#L743
	sub.ReceiveSettings.Synchronous = false
	sub.ReceiveSettings.NumGoroutines = 10
	sub.ReceiveSettings.MaxOutstandingMessages = 1000

	callback := func(ctx context.Context, msg *pubsub.Message) {
		defer func() {
			if err := recover(); err != nil {
				msg.Nack()
			}
		}()
		info := gosub.NewSubscriberInfo(topicName, subscriptionName, sub, "OnHoge2", msg)
		var event HogeEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		if err := gosub.Handle(is.option.Interceptors, ctx, &event, info, func(ctx context.Context, req interface{}) error {
			return is.service.OnHoge2(ctx, req.(*HogeEvent))
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

// Example2SubscriberPubSubAccessor: Interface for create topic and subscription
type Example2SubscriberPubSubAccessor interface {
	CreateHelloWorld2TopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error)
	CreateHelloWorld2SubscriptionIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Subscription, error)

	CreateOnHoge2TopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error)
	CreateOnHoge2SubscriptionIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Subscription, error)
}

type innerExample2SubscriberPubSubAccessor struct{}

func NewExample2SubscriberPubSubAccessor() Example2SubscriberPubSubAccessor {
	return &innerExample2SubscriberPubSubAccessor{}
}

func (c *innerExample2SubscriberPubSubAccessor) CreateHelloWorld2TopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error) {
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

func (c *innerExample2SubscriberPubSubAccessor) CreateHelloWorld2SubscriptionIFNotExists(
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
			Topic:                     t,
			AckDeadline:               60 * time.Second,
			EnableExactlyOnceDelivery: false,
		})
	}
	return sub, nil
}

func (c *innerExample2SubscriberPubSubAccessor) CreateOnHoge2TopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error) {
	topicName := "hogeCreated"
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

func (c *innerExample2SubscriberPubSubAccessor) CreateOnHoge2SubscriptionIFNotExists(
	ctx context.Context,
	client *pubsub.Client,
) (*pubsub.Subscription, error) {
	subscriptionName := "onHogeCreated"
	topicName := "hogeCreated"
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
			Topic:                     t,
			AckDeadline:               30 * time.Second,
			EnableExactlyOnceDelivery: false,
		})
	}
	return sub, nil
}
