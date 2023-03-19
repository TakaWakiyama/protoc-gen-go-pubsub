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

func Run(service HelloWorldSubscriber, client *pubsub.Client) {
	ctx := context.Background()
	if err := listenHelloWorld(ctx, service, client); err != nil {
		panic(err)
	}
}
func createSubscriptionIFNotExists(ctx context.Context, client *pubsub.Client, topicName string, subscriptionName string) (*pubsub.Subscription, error) {
	t := client.Topic(topicName)
	if exsits, err := t.Exists(ctx); !exsits {
		return nil, fmt.Errorf("topic does not exsit: %w", err)
	}
	sub := client.Subscription(subscriptionName)
	if exsits, _ := sub.Exists(ctx); !exsits {
		client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic:       t,
			AckDeadline: 60 * time.Second,
		})
	}
	return sub, nil
}
func listenHelloWorld(ctx context.Context, service HelloWorldSubscriber, client *pubsub.Client) error {
	subscriptionName := "helloworldsubscription"
	topicName := "helloworldtopic"
	var sub *pubsub.Subscription
	err := retry.Do(func() error {
		tmp, err := createSubscriptionIFNotExists(ctx, client, topicName, subscriptionName)
		if err != nil {
			return err
		}
		sub = tmp
		return nil
	})
	// TODO: メッセージの処理時間の延長を実装する必要がある
	callback := func(ctx context.Context, msg *pubsub.Message) {
		var event HelloWorldRequest
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		if err := service.HelloWorld(ctx, &event); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	}
	if err = sub.Receive(ctx, callback); err != nil {
		return err
	}
	return nil
}
