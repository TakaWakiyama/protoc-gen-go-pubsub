package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	event "github.com/TakaWakiyama/protoc-gen-go-pubsub/example/generated"
	"github.com/google/uuid"
)

type service struct{}

func (s service) HelloWorld(ctx context.Context, req *event.HelloWorldRequest) error {
	fmt.Printf("HelloWorld Event req: %+v\n", req)
	return nil
}

func main() {
	ctx := context.Background()
	proj := os.Getenv("PROJECT_ID")
	// Read GOOGLE_APPLICATION_CREDENTIALS or PUBSUB_EMULATOR_HOST
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	defer client.Close()
	// utils
	t, _ := event.GetOrCreateTopicIfNotExists(client, "helloworldtopic")
	if _, err := client.CreateSubscription(ctx, "helloworldsubscription", pubsub.SubscriptionConfig{
		Topic:       t,
		AckDeadline: 60 * time.Second,
	}); err != nil {
		fmt.Printf("err: %v\n", err)
	}
	fun := os.Getenv("PFUNC")
	if fun == "" {
		subscribe(ctx, proj)
	} else {
		publish(ctx, client)
	}
}

func subscribe(ctx context.Context, proj string) {
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	fmt.Println("Service Start")
	s := service{}

	interceptor := func(ctx context.Context, msg interface{}, info event.SubscriberInfo, handler event.SubscriberHandler) error {
		fmt.Printf("interceptor1 \ninfo: %+v\n", info)
		err := handler(ctx, msg)
		return err
	}
	interceptor2 := func(ctx context.Context, msg interface{}, _ event.SubscriberInfo, handler event.SubscriberHandler) error {
		start := time.Now()
		err := handler(ctx, msg)
		fmt.Printf("interceptor2: %v\n", time.Since(start))
		return err
	}

	event.Run(s, client, interceptor, interceptor2)
}

func publish(ctx context.Context, client *pubsub.Client) {
	c := event.NewHelloWorldServiceClient(client)
	msg := uuid.New().String()
	eid, err := c.PublishHelloWorld(ctx, &event.HelloWorldRequest{
		Name:          "Taka",
		EventID:       msg,
		UnixTimeStamp: time.Now().Unix(),
	})
	fmt.Printf("eid: %v\n", eid)
	fmt.Printf("err: %v\n", err)
}
