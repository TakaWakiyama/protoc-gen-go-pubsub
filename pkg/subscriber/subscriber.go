package subscriber

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/avast/retry-go"
)

type GracefulStopOption struct {
	Timeout     time.Duration
	MaxAttempts uint
	Delay       time.Duration
}

var defaultGracefulStopOption = GracefulStopOption{
	Timeout:     60 * time.Second,
	MaxAttempts: 10,
	Delay:       1 * time.Second,
}

func SubscribeGracefully(
	sub *pubsub.Subscription,
	ctx context.Context,
	callback func(ctx context.Context, msg *pubsub.Message),
	opt *GracefulStopOption,
) error {
	if opt == nil {
		opt = &defaultGracefulStopOption
	}
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

	ctx, cancel := context.WithTimeout(ctx, opt.Timeout*time.Second)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	stopped = true
	retry.Do(func() error {
		if len(idToDoing) == 0 {
			cancel()
			return nil
		}
		return fmt.Errorf("not all messages are processed")
	}, retry.Attempts(opt.MaxAttempts), retry.Delay(opt.Delay))

	return nil
}
