package main

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	event "github.com/TakaWakiyama/protoc-gen-go-pubsub/example/generated"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"

	"cloud.google.com/go/pubsub"
)

// const defaultTimeout = 10 * time.Second

type onCalledFunc = func(ctx context.Context, req proto.Message) error

var doNotghin = func(ctx context.Context, req proto.Message) error {
	return nil
}

type sub struct {
	onCalledHello onCalledFunc
	onCalledHoge  onCalledFunc
}

func (s sub) HelloWorld(ctx context.Context, req *event.HelloWorldEvent) error {
	return s.onCalledHello(ctx, req)
}

func (s sub) OnHoge(ctx context.Context, req *event.HogeEvent) error {
	return s.onCalledHoge(ctx, req)
}

func newsub(cbh onCalledFunc, cnh onCalledFunc) sub {
	return sub{
		onCalledHello: cbh,
		onCalledHoge:  cnh,
	}
}

func setup() {
	ctx := context.Background()
	proj := os.Getenv("PROJECT_ID")
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	// errorになっているメッセージを全てflushする
	go func() {
		event.Run(service{}, client, nil)
	}()
	time.Sleep(300 * time.Millisecond)
}

// go test -timeout=30s

// should settimeout
func Test(t *testing.T) {
	ctx := context.Background()
	proj := os.Getenv("PROJECT_ID")
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	defer client.Close()
	c := event.NewHelloWorldServiceClient(client, nil)
	acc := event.NewPubSubAccessor()
	acc.CreateHelloWorldTopicIFNotExists(ctx, client)

	want := &event.HelloWorldEvent{
		Name:          "a",
		EventID:       "1234",
		UnixTimeStamp: time.Now().Unix(),
	}
	c.PublishHelloWorld(ctx, want)

	resultChan := make(chan *event.HelloWorldEvent)
	helloCalled := func(ctx context.Context, e proto.Message) error {
		resultChan <- e.(*event.HelloWorldEvent)
		return nil
	}
	// action
	s := newsub(helloCalled, doNotghin)
	go func() {
		event.Run(s, client, nil)
	}()
	// assert
	actual := <-resultChan
	opt := cmpopts.IgnoreUnexported(event.HelloWorldEvent{})
	if diff := cmp.Diff(want, actual, opt); diff != "" {
		t.Errorf("X value is mismatch (-num1 +num2):%s\n", diff)
	}
}

// Publish後にSubscriberMethodが呼ばれていること。データが正しく引数にわたっていること。
// err if topic does not exist errになること
// BatchPublish後にSubscriberMethodが呼ばれていること。データが正しく引数にわたっていること。
// SubscriberMethodがpanicした場合、panicが発生しないこと。
// errが発生した場合Nackが呼ばれること。
// SubscriberMethodが完了した場合Ackが呼ばれること。
// Interceptorが呼ばれていること。1つ
// Interceptorが呼ばれていること。複数
// Interceptorがpanicした場合、panicが発生しないこと。
// Gracefullyがtrueの場合、Publishが完了するまで待つこと。
// Gracefullyがtrueの場合、SubscriberMethodが完了するまで待つこと。
