package publisher

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	retry "github.com/avast/retry-go"
	"github.com/google/uuid"
)

type BatchPublishResult struct {
	ID    string
	Error error
}

type PublisherOption struct {
	Gracefully  bool
	MaxAttempts uint
	Delay       time.Duration
}

var defaultPublisherOption = PublisherOption{
	Gracefully:  false,
	MaxAttempts: 3,
	Delay:       1 * time.Second,
}

type Publisher interface {
	Publish(ctx context.Context, topic *pubsub.Topic, msg *pubsub.Message) (string, error)
	BatchPublish(ctx context.Context, topic *pubsub.Topic, msgs []*pubsub.Message) ([]BatchPublishResult, error)
}

type innerPublisher struct {
	client  *pubsub.Client
	option  PublisherOption
	pids    map[string]struct{}
	running bool
}

func NewPublisher(client *pubsub.Client, option *PublisherOption) Publisher {
	if option == nil {
		option = &defaultPublisherOption
	}
	return &innerPublisher{
		client:  client,
		option:  *option,
		pids:    make(map[string]struct{}),
		running: true,
	}
}

func (p *innerPublisher) Publish(ctx context.Context, topic *pubsub.Topic, msg *pubsub.Message) (string, error) {
	if !p.running {
		return "", nil
	}
	p.gracefullyStopIFSet()
	pid := p.start()
	defer p.stop(pid)
	result := topic.Publish(ctx, msg)
	id, err := result.Get(ctx)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (p *innerPublisher) BatchPublish(ctx context.Context, topic *pubsub.Topic, msgs []*pubsub.Message) ([]BatchPublishResult, error) {
	if len(msgs) == 0 {
		return nil, nil
	}
	if !p.running {
		return nil, errors.New("publisher is not running")
	}
	p.gracefullyStopIFSet()
	pid := p.start()
	defer p.stop(pid)
	var results = make([]*pubsub.PublishResult, len(msgs))
	for _, msg := range msgs {
		result := topic.Publish(ctx, msg)
		results = append(results, result)
	}
	out := make([]BatchPublishResult, len(msgs))
	for i, result := range results {
		// Block until the result is returned and a server-generated
		id, err := result.Get(ctx)
		out[i] = BatchPublishResult{
			ID:    id,
			Error: err,
		}
	}

	return out, nil
}

func (p *innerPublisher) start() string {
	if !p.option.Gracefully {
		return ""
	}
	u := uuid.New().String()
	p.pids[u] = struct{}{}
	return u
}

func (p *innerPublisher) stop(pid string) {
	if !p.option.Gracefully {
		return
	}
	delete(p.pids, pid)
}

func (p *innerPublisher) gracefullyStopIFSet() {
	if !p.option.Gracefully {
		return
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sig
		p.running = false
		retry.Do(func() error {
			if len(p.pids) == 0 {
				return nil
			}
			return errors.New("publisher is still running")
		},
			retry.Attempts(p.option.MaxAttempts),
			retry.Delay(p.option.Delay),
		)
	}()
}

func GetOrCreateTopicIfNotExists(client *pubsub.Client, topicName string) (*pubsub.Topic, error) {
	topic := client.Topic(topicName)
	exists, err := topic.Exists(context.Background())
	if err != nil {
		return nil, err
	}
	if !exists {
		topic, err = client.CreateTopic(context.Background(), topicName)
		if err != nil {
			return nil, err
		}
	}
	return topic, nil
}
