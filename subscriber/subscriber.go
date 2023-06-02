package subscriber

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
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

	ctx2, cancel := context.WithCancel(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	if err := sub.Receive(ctx2, callback); err != nil {
		return err
	}

	return nil
}
