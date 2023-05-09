package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	event "github.com/TakaWakiyama/protoc-gen-go-pubsub/example/generated"
	gosub "github.com/TakaWakiyama/protoc-gen-go-pubsub/subscriber"
	"github.com/google/uuid"
)

type service struct{}

func (s service) HelloWorld(ctx context.Context, req *event.HelloWorldEvent) error {
	fmt.Printf("HelloWorld Event req: %+v\n", req)
	return nil
}

func (s service) OnHoge(ctx context.Context, req *event.HogeEvent) error {
	panic("implement me")
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

	/*
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
	*/
	option := &event.SubscriberOption{
		Interceptors: []gosub.SubscriberInterceptor{
			// interceptor,
			// interceptor2,
			// interceptor3,
		},
	}

	if err := event.Run(s, client, option); err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		fmt.Println("Service End")
	}
}

func publish(ctx context.Context, client *pubsub.Client) {
	c := event.NewHelloWorldServiceClient(client, nil)
	msg := uuid.New().String()
	c.PublishHelloWorld(ctx, &event.HelloWorldEvent{
		Name:          "Taka",
		EventID:       msg,
		UnixTimeStamp: time.Now().Unix(),
	})
	c.PublishHogeCreated(ctx, &event.HogeEvent{
		Message:       "Taka",
		EventID:       msg,
		UnixTimeStamp: time.Now().Unix(),
	})
	if true {
		return
	}
	c.BatchPublishHogesCreated(ctx, []*event.HogeEvent{
		{
			Message:       "Taka",
			EventID:       msg,
			UnixTimeStamp: time.Now().Unix(),
		},
		{
			Message:       "Taka",
			EventID:       msg,
			UnixTimeStamp: time.Now().Unix(),
		},
	})
}
