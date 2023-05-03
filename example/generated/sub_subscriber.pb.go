package example

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	retry "github.com/avast/retry-go"
	"google.golang.org/protobuf/proto"
)

type HelloWorldSubscriber interface {
	HelloWorld(ctx context.Context, req *HelloWorldRequest) error
}

func Run(service HelloWorldSubscriber, client *pubsub.Client, interceptors ...SubscriberInterceptor) error {
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

type SubscriberInfo interface {
	GetTopicName() string
	GetSubscriptionName() string
	GetSubscription() pubsub.Subscription
	GetMethod() string
	GetMessage() pubsub.Message
}

type innerSubscriberInfo struct {
	topicName        string
	subscriptionName string
	subscription     *pubsub.Subscription
	method           string
	message          *pubsub.Message
}

func newInnerSubscriberInfo(topicName string, subscriptionName string, subscription *pubsub.Subscription, method string, message *pubsub.Message) *innerSubscriberInfo {
	return &innerSubscriberInfo{
		topicName:        topicName,
		subscriptionName: subscriptionName,
		subscription:     subscription,
		method:           method,
		message:          message,
	}
}

func (i *innerSubscriberInfo) GetTopicName() string {
	return i.topicName
}
func (i *innerSubscriberInfo) GetSubscriptionName() string {
	return i.subscriptionName
}
func (i *innerSubscriberInfo) GetSubscription() pubsub.Subscription {
	return *i.subscription
}
func (i *innerSubscriberInfo) GetMethod() string {
	return i.method
}

func (i *innerSubscriberInfo) GetMessage() pubsub.Message {
	return *i.message
}

// https://github.com/grpc/grpc-go/blob/v1.53.0/interceptor.go#L87
// type UnaryServerInterceptor func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)
type SubscriberHandler func(ctx context.Context, req interface{}) error
type SubscriberInterceptor func(ctx context.Context, msg interface{}, info SubscriberInfo, handler SubscriberHandler) error

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

func (is *innerHelloWorldSubscriberSubscriber) chainInterceptors(ctx context.Context, event interface{}, info *innerSubscriberInfo, handler SubscriberHandler) SubscriberHandler {
	if len(is.interceptors) == 0 {
		return handler
	}
	var out SubscriberHandler = handler
	for i := 0; i < len(is.interceptors); i++ {
		f := func(ctx context.Context, req interface{}) error {
			return is.interceptors[i](ctx, req, info, out)
		}
		out = f
	}
	return out
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
	callback := func(ctx context.Context, msg *pubsub.Message) {
		info := newInnerSubscriberInfo(topicName, subscriptionName, sub, "HelloWorld", msg)
		var event HelloWorldRequest
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		handler := is.chainInterceptors(ctx, &event, info, func(ctx context.Context, req interface{}) error {
			return is.service.HelloWorld(ctx, req.(*HelloWorldRequest))
		})
		if err := handler(ctx, &event); err != nil {
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
