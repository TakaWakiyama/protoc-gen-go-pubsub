package example

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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

func (is *innerHelloWorldSubscriberSubscriber) handle(
	ctx context.Context,
	event interface{},
	info SubscriberInfo,
	handler SubscriberHandler,
) error {
	if len(is.interceptors) == 0 {
		return handler(ctx, event)
	}
	f := is.chainInterceptors(is.interceptors)
	return f(ctx, event, info, handler)
}

func (is *innerHelloWorldSubscriberSubscriber) chainInterceptors(
	interceptors []SubscriberInterceptor,
) SubscriberInterceptor {
	return func(ctx context.Context, event interface{}, info SubscriberInfo, handler SubscriberHandler) error {
		return interceptors[0](ctx, event, info, is.getchainInterceptorHandler(interceptors, 0, ctx, event, info, handler))
	}
}

func (is *innerHelloWorldSubscriberSubscriber) getchainInterceptorHandler(
	interceptors []SubscriberInterceptor,
	curr int,
	ctx context.Context,
	event interface{},
	info SubscriberInfo,
	finalHandler SubscriberHandler,
) SubscriberHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}
	return func(ctx context.Context, event interface{}) error {
		return interceptors[curr+1](ctx, event, info, is.getchainInterceptorHandler(interceptors, curr+1, ctx, event, info, finalHandler))
	}
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
		fmt.Printf("\"starthandleing\": %v\n", "starthandleing")
		time.Sleep(6 * time.Second)
		info := newInnerSubscriberInfo(topicName, subscriptionName, sub, "HelloWorld", msg)
		var event HelloWorldRequest
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		if err := is.handle(ctx, &event, info, func(ctx context.Context, req interface{}) error {
			return is.service.HelloWorld(ctx, req.(*HelloWorldRequest))
		}); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	}

	subscribeGracefully(sub, ctx, callback)

	return nil
}

func subscribeGracefully(sub *pubsub.Subscription, ctx context.Context, callback func(ctx context.Context, msg *pubsub.Message)) error {
	stopped := false
	idToDoing := map[string]interface{}{}
	wrapper := func(ctx context.Context, msg *pubsub.Message) {
		if stopped {
			msg.Nack()
			return
		}
		idToDoing[msg.ID] = nil
		callback(ctx, msg)
		delete(idToDoing, msg.ID)
	}

	eChan := make(chan error, 1)
	go func() {
		err := sub.Receive(ctx, wrapper)
		if err != nil {
			eChan <- err
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Printf("\"stopping gracefully...\": %v\n", "stopping gracefully...")
	stopped = true
	retry.Do(func() error {
		if len(idToDoing) == 0 {
			cancel()
			return nil
		}
		return fmt.Errorf("not all messages are processed")
	}, retry.Attempts(15), retry.Delay(3*time.Second))

	return nil
}

// TODO: testコード書く
// Interceptor
// GracefulStop
// 1. メッセージの処理時間の延長
