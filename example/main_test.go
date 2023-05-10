package main

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	event "github.com/TakaWakiyama/protoc-gen-go-pubsub/example/generated"
	gosub "github.com/TakaWakiyama/protoc-gen-go-pubsub/subscriber"
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

func clear() {
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

func fakeHelloWorldEvent() *event.HelloWorldEvent {
	return &event.HelloWorldEvent{
		Name:          uuid.NewString(),
		EventID:       uuid.NewString(),
		UnixTimeStamp: time.Now().Unix(),
	}
}

func fakeHogeCreatedEvent() *event.HogeEvent {
	return &event.HogeEvent{
		Message:       uuid.NewString(),
		EventID:       uuid.NewString(),
		UnixTimeStamp: time.Now().Unix(),
	}
}

func setTestTimeout(t *testing.T, d time.Duration) {
	t.Helper()
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(context.TODO(), d)
	t.Cleanup(func() {
		cancel()
	})
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			t.Error("test timeout")
		}
	}()
}

func TestMain(m *testing.M) {
	clear()
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
	want := fakeHelloWorldEvent()
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
				fakeHogeCreatedEvent(),
			},
			batchPublish: false,
		},
		{
			name: "subscriber method is to be called after two events published for each. each Data are to be passed to subscriber method.",
			want: []*event.HogeEvent{
				fakeHogeCreatedEvent(),
				fakeHogeCreatedEvent(),
			},
			batchPublish: false,
		},
		{
			name: "subscriber method is to be called after two events published at once. Data are to be passed to subscriber method.",
			want: []*event.HogeEvent{
				fakeHogeCreatedEvent(),
				fakeHogeCreatedEvent(),
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
			clear()
		})
	}
}

func TestSubscriberInterceptor(t *testing.T) {
	ctx := context.Background()
	proj := os.Getenv("PROJECT_ID")
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	defer client.Close()
	c := event.NewHelloWorldServiceClient(client, nil)
	ac := event.NewPubSubAccessor()
	if _, err := ac.CreateHelloWorldTopicIFNotExists(ctx, client); err != nil {
		t.Fatalf("CreateOnHelloWorldSubscriptionIFNotExists error: %v", err)
	}

	o := []int{}
	reset := func() {
		o = []int{}
	}
	add := func(c int) {
		o = append(o, c)
	}

	testCases := []struct {
		name         string
		want         *event.HelloWorldEvent
		wanto        []int
		interceptors []gosub.SubscriberInterceptor
	}{
		{
			name:  "subscriber method works when interceptor is nil.",
			wanto: []int{},
			want:  fakeHelloWorldEvent(),
		},
		{
			name:  "an interceptor is to be called.",
			want:  fakeHelloWorldEvent(),
			wanto: []int{1, 2},
			interceptors: []gosub.SubscriberInterceptor{
				func(ctx context.Context, msg interface{}, info gosub.SubscriberInfo, handler gosub.SubscriberHandler) error {
					add(1)
					res := handler(ctx, msg)
					add(2)
					return res
				},
			},
		},
		{
			name: "interceptors are to be called in order.",
			want: &event.HelloWorldEvent{
				Name:          uuid.NewString(),
				EventID:       uuid.NewString(),
				UnixTimeStamp: time.Now().Unix(),
			},
			wanto: []int{1, 2, 3, 4},
			interceptors: []gosub.SubscriberInterceptor{
				func(ctx context.Context, msg interface{}, info gosub.SubscriberInfo, handler gosub.SubscriberHandler) error {
					add(1)
					res := handler(ctx, msg)
					add(4)
					return res
				},
				func(ctx context.Context, msg interface{}, info gosub.SubscriberInfo, handler gosub.SubscriberHandler) error {
					add(2)
					res := handler(ctx, msg)
					add(3)
					return res
				},
			},
		},
	}
	t.Cleanup(func() {
		reset()
		clear()
	})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultChan := make(chan *event.HelloWorldEvent)
			onHelloWorldCreated := func(ctx context.Context, e proto.Message) error {
				resultChan <- e.(*event.HelloWorldEvent)
				return nil
			}
			// action
			s := newsub(onHelloWorldCreated, doNotghing)
			go func() {
				opt := &event.SubscriberOption{
					Interceptors: tc.interceptors,
				}
				event.Run(s, client, opt)
			}()
			time.Sleep(1 * time.Second)
			if _, err := c.PublishHelloWorld(ctx, tc.want); err != nil {
				t.Fatalf("PublishHelloWorld error: %v", err)
			}
			actual := <-resultChan
			opt := cmpopts.IgnoreUnexported(event.HelloWorldEvent{})
			if diff := cmp.Diff(tc.want, actual, opt); diff != "" {
				t.Errorf("HelloWorldEvent value is mismatch (-actual +opt):%s\n", diff)
			}
			if diff := cmp.Diff(tc.wanto, o); diff != "" {
				t.Errorf("X value is mismatch (-num1 +num2):%s\n", diff)
			}
		})
	}
}

func TestRunErrorIFTopicDoesnotExsits(t *testing.T) {
	ctx := context.Background()
	proj := os.Getenv("PROJECT_ID")
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	defer client.Close()
	testCases := []struct {
		name      string
		topicName string
	}{
		{
			name:      "helloworldtopic topic does not exist",
			topicName: "helloworldtopic",
		},
		{
			name:      "hogeCreated topic does not exist",
			topicName: "hogeCreated",
		},
	}
	setTestTimeout(t, 10*time.Second)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := client.Topic(tc.topicName).Delete(ctx); err != nil {
				t.Fatalf("Delete error: %v", err)
			}
			s := newsub(doNotghing, doNotghing)
			err = event.Run(s, client, nil)
			if err == nil {
				t.Errorf("Run error is nil")
			}
		})
	}
}

// TestRunErrorIFSubscriberMethodPanic: recover and nack when subscriber method panics.
// TODO: pubsub messageの nackを完全に検知するために ack or nackする構造体のIFを作成する必要がある。
// Note: 2回呼ばれることでnackされることを確認している。
func TestRunErrorIFSubscriberMethodPanic(t *testing.T) {
	ctx := context.Background()
	proj := os.Getenv("PROJECT_ID")
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	defer client.Close()
	setTestTimeout(t, 30*time.Second)
	ac := event.NewPubSubAccessor()
	ac.CreateHelloWorldTopicIFNotExists(ctx, client)
	ac.CreateHelloWorldSubscriptionIFNotExists(ctx, client)
	ac.CreateOnHogeTopicIFNotExists(ctx, client)
	ac.CreateOnHogeSubscriptionIFNotExists(ctx, client)
	ch := make(chan struct{})
	panicedOnce := false
	t.Cleanup(func() {
		clear()
		panicedOnce = false
	})
	testCases := []struct {
		name         string
		interceptors []gosub.SubscriberInterceptor
		callback     onCalledFunc
	}{
		{
			name:         "subscriber method panic",
			interceptors: nil,
			callback: func(ctx context.Context, e proto.Message) error {
				if panicedOnce {
					ch <- struct{}{}
					return nil
				}
				panicedOnce = true
				panic("panic")
			},
		},
		{
			name: "subscriber method panic with interceptor",
			interceptors: []gosub.SubscriberInterceptor{
				func(ctx context.Context, msg interface{}, info gosub.SubscriberInfo, handler gosub.SubscriberHandler) error {
					ch <- struct{}{}
					panic("panic")
				},
			},
			callback: func(ctx context.Context, e proto.Message) error {
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := event.NewHelloWorldServiceClient(client, nil)
			s := newsub(doNotghing, tc.callback)
			opt := &event.SubscriberOption{
				Interceptors: tc.interceptors,
			}
			go func() {
				if e := event.Run(s, client, opt); e != nil {
					t.Errorf("Run error: %v", e)
				}
			}()
			if _, err := c.PublishHogeCreated(ctx, fakeHogeCreatedEvent()); err != nil {
				t.Fatalf("PublishHogeCreated error: %v", err)
			}
			<-ch
			if !panicedOnce {
				t.Errorf("panic did not occur")
			}
		})
	}
}
