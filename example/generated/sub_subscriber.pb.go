// Code generated  by protoc-gen-go-event. DO NOT EDIT.
// versions:
// source: sub.proto

package example

import (
	"context"
	"cloud.google.com/go/pubsub"
	"google.golang.org/protobuf/proto"
	"fmt"
	"time"
	retry "github.com/avast/retry-go"
)

type HelloWorldSubscriber interface {
	HelloWorld(ctx context.Context, req *HelloWorldRequest) error
}

func Run(service HelloWorldSubscriber, client *pubsub.Client, interceptors ...SubscriberInterceptor) error {
	if service == nil {
		return fmt.Errorf("service is nil")
	}
	if client == nil {
		return fmt.Errorf("service is nil")
	}
	ctx := context.Background()
	is := newInnerHelloWorldSubscriberSubscriber(service, client, interceptors...)

	if err := is.listenHelloWorld(ctx); err != nil {
		return err
	}
	return nil
}

type SubscriberInfo interface {
	GetTopicName() string
	GetSubscriptionName() string
	GetSubscription() *pubsub.Subscription
	GetMethod() string
}

type innerSubscriberInfo struct {
	topicName        string
	subscriptionName string
	subscription     *pubsub.Subscription
	method           string
}

func newInnerSubscriberInfo(topicName string, subscriptionName string, subscription *pubsub.Subscription, method string) *innerSubscriberInfo {
	return &innerSubscriberInfo{
		topicName:        topicName,
		subscriptionName: subscriptionName,
		subscription:     subscription,
		method:           method,
	}
}

func (i *innerSubscriberInfo) GetTopicName() string {
	return i.topicName
}
func (i *innerSubscriberInfo) GetSubscriptionName() string {
	return i.subscriptionName
}
func (i *innerSubscriberInfo) GetSubscription() *pubsub.Subscription {
	return i.subscription
}
func (i *innerSubscriberInfo) GetMethod() string {
	return i.method
}

type SubscriberInterceptor func(ctx context.Context, msg interface{}, info SubscriberInfo) error

type innerHelloWorldSubscriberSubscriber struct {
	service      HelloWorldSubscriber
	client       *pubsub.Client
	interceptors []SubscriberInterceptor
}

func newInnerHelloWorldSubscriberSubscriber(service HelloWorldSubscriber, client *pubsub.Client, interceptors ...SubscriberInterceptor) *innerHelloWorldSubscriberSubscriber {
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

func (is *innerHelloWorldSubscriberSubscriber) chainInterceptors(info SubscriberInfo, msg interface{}) error {
	errors := []error{}
	hasError := false
	for i := len(is.interceptors) - 1; i >= 0; i-- {
		chain := is.interceptors[i]
		err := chain(context.Background(), msg, info)
		errors = append(errors, err)
		if err != nil {
			hasError = true
		}
	}
	if hasError {
		return fmt.Errorf("error occured in interceptors: %v", errors)
	}
	return nil
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
	// TODO: メッセージの処理時間の延長を実装する必要がある
	info := newInnerSubscriberInfo(topicName, subscriptionName, sub, "HelloWorld")
	callback := func(ctx context.Context, msg *pubsub.Message) {
		var event HelloWorldRequest
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}

		err := is.chainInterceptors(info, msg)
		if err != nil {
			msg.Nack()
			return
		}
		if err := is.service.HelloWorld(ctx, &event); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	}
	if err := sub.Receive(ctx, callback); err != nil {
		return err
	}
	return nil
}
