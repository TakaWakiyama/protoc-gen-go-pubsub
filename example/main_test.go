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
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"cloud.google.com/go/pubsub"
)

// const defaultTimeout = 10 * time.Second

// Note: mockgenを使ってテストコードを書くとcallbackの呼び出しが行われない現象に遭遇したので、自前で評価する構造体を作成した。
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

type onCalledFunc = func(ctx context.Context, req proto.Message) error

var doNotghing = func(ctx context.Context, req proto.Message) error {
	return nil
}

func setup() {
	ctx := context.Background()
	proj := os.Getenv("PROJECT_ID")
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	defer client.Close()
	// errorになっているメッセージを全てflushする
	ht := client.Topic("helloworldtopic")
	hct := client.Topic("hogeCreated")
	reset := func(t *pubsub.Topic) {
		if t == nil {
			return
		}
		itr := t.Subscriptions(ctx)
		for {
			sub, err := itr.Next()
			if err != nil {
				break
			}
			sub.Delete(ctx)
			t.Delete(ctx)
		}
	}
	reset(ht)
	reset(hct)
	acc := event.NewPubSubAccessor()
	if _, err := acc.CreateHelloWorldTopicIFNotExists(ctx, client); err != nil {
		log.Fatalf("CreateHelloWorldTopicIFNotExists error: %v", err)
	}
	if _, err := acc.CreateOnHogeTopicIFNotExists(ctx, client); err != nil {
		log.Fatalf("CreateHogeCreatedTopicIFNotExists error: %v", err)
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

// TestPusSubHelloWorld: Subscriber method is to be called after publish. Data is to be passed to subscriber method.
func TestPusSubHelloWorld(t *testing.T) {
	ctx := context.Background()
	proj := os.Getenv("PROJECT_ID")
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	defer client.Close()
	c := event.NewHelloWorldServiceClient(client, nil)
	resultChan := make(chan *event.HelloWorldEvent)
	helloCalled := func(ctx context.Context, e proto.Message) error {
		resultChan <- e.(*event.HelloWorldEvent)
		return nil
	}
	// action
	s := newsub(helloCalled, doNotghing)
	go func() {
		if err := event.Run(s, client, nil); err != nil {
			panic(err)
		}
	}()
	time.Sleep(1 * time.Second)
	want := &event.HelloWorldEvent{
		Name:          uuid.NewString(),
		EventID:       uuid.NewString(),
		UnixTimeStamp: time.Now().Unix(),
	}
	if _, err := c.PublishHelloWorld(ctx, want); err != nil {
		t.Fatalf("PublishHelloWorld error: %v", err)
	}
	// assert
	actual := <-resultChan
	opt := cmpopts.IgnoreUnexported(event.HelloWorldEvent{})
	if diff := cmp.Diff(want, actual, opt); diff != "" {
		t.Errorf("X value is mismatch (-num1 +num2):%s\n", diff)
	}
}

// TestPusSubOnHoge: Publish後にSubscriberMethodが呼ばれていること。データが正しく引数にわたっていること。
func TestPusSubOnHoge(t *testing.T) {
	ctx := context.Background()
	proj := os.Getenv("PROJECT_ID")
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	defer client.Close()
	c := event.NewHelloWorldServiceClient(client, nil)
	testCases := []struct {
		name         string
		want         []*event.HogeEvent
		batchPublish bool
	}{
		{
			name: "subscriber method is to be called after an event published. Data is to be passed to subscriber method.",
			want: []*event.HogeEvent{
				{
					Message:       uuid.NewString(),
					EventID:       uuid.NewString(),
					UnixTimeStamp: time.Now().Unix(),
				},
			},
			batchPublish: false,
		},
		{
			name: "subscriber method is to be called after two events published for each. each Data are to be passed to subscriber method.",
			want: []*event.HogeEvent{
				{
					Message:       uuid.NewString(),
					EventID:       uuid.NewString(),
					UnixTimeStamp: time.Now().Unix(),
				},
				{
					Message:       uuid.NewString(),
					EventID:       uuid.NewString(),
					UnixTimeStamp: time.Now().Unix(),
				},
			},
			batchPublish: false,
		},
		{
			name: "subscriber method is to be called after two events published at once. Data are to be passed to subscriber method.",
			want: []*event.HogeEvent{
				{
					Message:       uuid.NewString(),
					EventID:       uuid.NewString(),
					UnixTimeStamp: time.Now().Unix(),
				},
				{
					Message:       uuid.NewString(),
					EventID:       uuid.NewString(),
					UnixTimeStamp: time.Now().Unix(),
				},
			},
			batchPublish: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ac := event.NewPubSubAccessor()
			if _, err := ac.CreateOnHogeSubscriptionIFNotExists(ctx, client); err != nil {
				t.Fatalf("CreateOnHogeSubscriptionIFNotExists error: %v", err)
			}
			resultChan := make(chan []*event.HogeEvent)
			result := []*event.HogeEvent{}
			onHogeCreated := func(ctx context.Context, e proto.Message) error {
				result = append(result, e.(*event.HogeEvent))
				if len(result) == len(tc.want) {
					resultChan <- result
				}
				return nil
			}
			// action
			s := newsub(doNotghing, onHogeCreated)
			go func() {
				event.Run(s, client, nil)
			}()
			time.Sleep(1 * time.Second)
			if tc.batchPublish {
				if _, err := c.BatchPublishHogesCreated(ctx, tc.want); err != nil {
					t.Fatalf("BatchPublishHogeCreated error: %v", err)
				}
			} else {
				for _, w := range tc.want {
					if _, err := c.PublishHogeCreated(ctx, w); err != nil {
						t.Fatalf("PublishOnHoge error: %v", err)
					}
				}
			}
			actual := <-resultChan
			// Note: ignore order of slice elements due to asychronous processing
			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(event.HogeEvent{}),
				cmpopts.SortSlices(func(i, j *event.HogeEvent) bool {
					return i.EventID < j.EventID
				}),
			}
			if diff := cmp.Diff(tc.want, actual, opts...); diff != "" {
				t.Errorf("X value is mismatch (-num1 +num2):%s\n", diff)
			}
			setup()
		})
	}
}

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
