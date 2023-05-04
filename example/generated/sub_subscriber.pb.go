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

type HelloWorldSubscriber interface {
	HelloWorld(ctx context.Context, req *HelloWorldRequest) error
}

func Run(service HelloWorldSubscriber, client *pubsub.Client, interceptors ...gosub.SubscriberInterceptor) error {
	if service == nil {
		return fmt.Errorf("service is nil")
	}
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	ctx := context.Background()
	is := newInnerHelloWorldSubscriberSubscriber(service, client, interceptors...)

	if err := is.listenHelloWorld(ctx); err != nil {
		return err
	}
	return nil
}

type innerHelloWorldSubscriberSubscriber struct {
	service      HelloWorldSubscriber
	client       *pubsub.Client
	interceptors []gosub.SubscriberInterceptor
}

func newInnerHelloWorldSubscriberSubscriber(service HelloWorldSubscriber, client *pubsub.Client, interceptors ...gosub.SubscriberInterceptor) *innerHelloWorldSubscriberSubscriber {
	return &innerHelloWorldSubscriberSubscriber{
		service:      service,
		client:       client,
		interceptors: interceptors,
	}
}

func (is *innerHelloWorldSubscriberSubscriber) createSubscriptionIFNotExists(ctx context.Context, topicName string, subscriptionName string) (*pubsub.Subscription, error) {
	t := is.client.Topic(topicName)
	if exsits, err := t.Exists(ctx); !exsits {
		return nil, fmt.Errorf("topic does not exsit: %w", err)
	}
	sub := is.client.Subscription(subscriptionName)
	if exsits, _ := sub.Exists(ctx); !exsits {
		is.client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic:       t,
			AckDeadline: 60 * time.Second,
		})
	}
	return sub, nil
}

func (is *innerHelloWorldSubscriberSubscriber) listenHelloWorld(ctx context.Context) error {
	subscriptionName := "helloworldsubscription"
	topicName := "helloworldtopic"
	var sub *pubsub.Subscription
	if err := retry.Do(func() error {
		tmp, err := is.createSubscriptionIFNotExists(ctx, topicName, subscriptionName)
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
		if err := gosub.Handle(is.interceptors, ctx, &event, info, func(ctx context.Context, req interface{}) error {
			return is.service.HelloWorld(ctx, req.(*HelloWorldRequest))
		}); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	}
	gosub.SubscribeGracefully(sub, ctx, callback, nil)

	return nil
}

// TODO: testコード書く
// Batch処理
