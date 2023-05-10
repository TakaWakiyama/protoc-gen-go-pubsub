package subscriber

import (
	"context"

	"cloud.google.com/go/pubsub"
)

// SubscriberInfo: access to subscriber information in interceptor
type SubscriberInfo interface {
	// GetTopicName: get pubsub topic name
	GetTopicName() string
	// GetSubscriptionName: get pubsub subscription name
	GetSubscriptionName() string
	// GetSubscription: get pubsub subscription
	GetSubscription() *pubsub.Subscription
	// GetMethod: get method name
	GetMethod() string
	// GetMessage: get pubsub row message
	GetMessage() pubsub.Message
}

type innerSubscriberInfo struct {
	topicName        string
	subscriptionName string
	subscription     *pubsub.Subscription
	method           string
	message          *pubsub.Message
}

func NewSubscriberInfo(topicName string, subscriptionName string, subscription *pubsub.Subscription, method string, message *pubsub.Message) SubscriberInfo {
	return &innerSubscriberInfo{
		topicName:        topicName,
		subscriptionName: subscriptionName,
		subscription:     subscription,
		method:           method,
		message:          message,
	}
}

func (i *innerSubscriberInfo) GetTopicName() string {
	return i.topicName
}
func (i *innerSubscriberInfo) GetSubscriptionName() string {
	return i.subscriptionName
}
func (i *innerSubscriberInfo) GetSubscription() *pubsub.Subscription {
	return i.subscription
}
func (i *innerSubscriberInfo) GetMethod() string {
	return i.method
}

func (i *innerSubscriberInfo) GetMessage() pubsub.Message {
	return *i.message
}

// SubscriberHandler: subscriber handler
type SubscriberHandler func(ctx context.Context, req interface{}) error

// SubscriberInterceptor: subscriber interceptor
type SubscriberInterceptor func(ctx context.Context, msg interface{}, info SubscriberInfo, handler SubscriberHandler) error

func Handle(
	interceptors []SubscriberInterceptor,
	ctx context.Context,
	event interface{},
	info SubscriberInfo,
	handler SubscriberHandler,
) error {
	if len(interceptors) == 0 {
		return handler(ctx, event)
	}
	f := chainInterceptors(interceptors)
	return f(ctx, event, info, handler)
}

func chainInterceptors(
	interceptors []SubscriberInterceptor,
) SubscriberInterceptor {
	return func(ctx context.Context, event interface{}, info SubscriberInfo, handler SubscriberHandler) error {
		return interceptors[0](ctx, event, info, getchainInterceptorHandler(interceptors, 0, info, handler))
	}
}

func getchainInterceptorHandler(
	interceptors []SubscriberInterceptor,
	curr int,
	info SubscriberInfo,
	finalHandler SubscriberHandler,
) SubscriberHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}
	return func(ctx context.Context, event interface{}) error {
		return interceptors[curr+1](ctx, event, info, getchainInterceptorHandler(interceptors, curr+1, info, finalHandler))
	}
}
