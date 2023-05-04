package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	event "github.com/TakaWakiyama/protoc-gen-go-pubsub/example/generated"
	gopub "github.com/TakaWakiyama/protoc-gen-go-pubsub/publisher"
	gosub "github.com/TakaWakiyama/protoc-gen-go-pubsub/subscriber"
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
	t, err := gopub.GetOrCreateTopicIfNotExists(client, "helloworld")
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

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

	interceptor := func(ctx context.Context, msg interface{}, info gosub.SubscriberInfo, handler gosub.SubscriberHandler) error {
		fmt.Printf("start interceptor1 \ninfo: %+v\n", info)
		err := handler(ctx, msg)
		fmt.Println("end interceptor1")
		return err
	}
	interceptor2 := func(ctx context.Context, msg interface{}, _ gosub.SubscriberInfo, handler gosub.SubscriberHandler) error {
		fmt.Println("start interceptor2")
		start := time.Now()
		err := handler(ctx, msg)
		fmt.Printf("end interceptor2: %v\n", time.Since(start))
		return err
	}

	interceptor3 := func(ctx context.Context, msg interface{}, e gosub.SubscriberInfo, handler gosub.SubscriberHandler) error {
		fmt.Printf("start interceptor3 \n messageID %v \n", e.GetMessage().ID)
		err := handler(ctx, msg)
		fmt.Printf("end interceptor3 \n messageID %v \n", e.GetMessage().ID)
		return err
	}
	option := &event.SubscriberOption{
		Interceptors: []gosub.SubscriberInterceptor{
			interceptor,
			interceptor2,
			interceptor3,
		},
	}

	event.Run(s, client, option)
}

func publish(ctx context.Context, client *pubsub.Client) {
	c := event.NewHelloWorldServiceClient(client, nil)
	msg := uuid.New().String()
	eid, err := c.PublishHelloWorld(ctx, &event.HelloWorldRequest{
		Name:          "Taka",
		EventID:       msg,
		UnixTimeStamp: time.Now().Unix(),
	})
	fmt.Printf("eid: %v\n", eid)
	fmt.Printf("err: %v\n", err)
}
