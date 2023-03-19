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
func listenHelloWorld(ctx context.Context, service HelloWorldSubscriber, client *pubsub.Client) error {
	subscriptionName := "helloworldsubscription"
	topicName := "helloworldtopic"
	t := client.Topic(topicName)
	if exsits, err := t.Exists(ctx); !exsits {
		return fmt.Errorf("topic does not exsit: %w", err)
	}
	sub := client.Subscription(subscriptionName)
	if exsits, _ := sub.Exists(ctx); !exsits {
		client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic: t,
			// TODO: デフォルトの時間はoptionで設定できるようにした方が良い
			AckDeadline: 60 * time.Second,
		})
	}
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
	err := pullMsgs(ctx, client, subscriptionName, topicName, callback)
	if err != nil {
		return err
	}
	return nil
}

func pullMsgs(ctx context.Context, client *pubsub.Client, subScriptionName, topicName string, callback func(context.Context, *pubsub.Message)) error {
	sub := client.Subscription(subScriptionName)
	err := sub.Receive(ctx, callback)
	if err != nil {
		return err
	}
	return nil
}
