package generator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TakaWakiyama/protoc-gen-go-pubsub/option"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	GO_IMPORT_CONTEXT      = "context"
	GO_IMPORT_PUBSUB       = "cloud.google.com/go/pubsub"
	GO_IMPORT_PROTO        = "google.golang.org/protobuf/proto"
	GO_IMPORT_PROTOREFLECT = "google.golang.org/protobuf/reflect/protoreflect"
	GO_IMPORT_FMT          = "fmt"
	GO_IMPORT_TIME         = "time"
	GO_IMPORT_RETRY        = `retry "github.com/avast/retry-go"`
)

var allImportSubMods = []string{
	GO_IMPORT_CONTEXT,
	GO_IMPORT_PUBSUB,
	GO_IMPORT_PROTO,
	GO_IMPORT_FMT,
	GO_IMPORT_TIME,
	GO_IMPORT_RETRY,
}

var allImportPubMods = []string{
	GO_IMPORT_CONTEXT,
	GO_IMPORT_PUBSUB,
	GO_IMPORT_PROTO,
	GO_IMPORT_PROTOREFLECT,
}

type pubsubGenerator struct {
	gen         *protogen.Plugin
	file        *protogen.File
	g           *protogen.GeneratedFile
	isPublisher bool
}

func NewFileGenerator(gen *protogen.Plugin, file *protogen.File, isPublisher bool) *pubsubGenerator {
	file_suffix := "_subscriber.pb.go"
	if isPublisher {
		file_suffix = "_publisher.pb.go"
	}
	filename := file.GeneratedFilenamePrefix + file_suffix
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	return &pubsubGenerator{
		gen:         gen,
		file:        file,
		g:           g,
		isPublisher: isPublisher,
	}
}

func (pg *pubsubGenerator) Generate() {
	if pg.isPublisher {
		pg.generatePublisherFile()
	} else {
		pg.generateSubscriberFile()
	}
}

func (pg *pubsubGenerator) decrearePackageName() {
	g := pg.g
	g.P("// Code generated  by protoc-gen-go-event. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// source: ", pg.file.Proto.GetName())
	g.P()
	g.P("package ", pg.file.GoPackageName)
	g.P()
	g.P("import (")
	mods := allImportSubMods
	if pg.isPublisher {
		mods = allImportPubMods
	}

	for _, mod := range mods {
		if strings.Contains(mod, `"`) {
			g.P(mod)
		} else {
			g.P(`"`, mod, `"`)
		}
	}
	g.P(")")
}

func (pg *pubsubGenerator) generateSubscriberFile() *protogen.GeneratedFile {
	g := pg.g
	pg.decrearePackageName()
	if len(pg.file.Services) == 0 {
		return g
	}
	svc := pg.file.Services[0]

	svcName := svc.GoName
	// Create interface
	g.P("type ", svcName, " interface {")
	for _, m := range svc.Methods {
		// g.P("// ", m.Comments.Leading, " ", m.Comments.Trailing)
		g.P(m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") error ")
	}
	g.P("}")

	// Create Run Function
	template := `func Run(service {_svcName}, client *pubsub.Client, interceptors ...SubscriberInterceptor) error {
		if service == nil {
			return fmt.Errorf("service is nil")
		}
		if client == nil {
			return fmt.Errorf("client is nil")
		}
		ctx := context.Background()
		is := newInner{_svcName}Subscriber(service, client, interceptors...)
	`
	g.P(strings.Replace(template, "{_svcName}", svcName, -1))
	for _, m := range svc.Methods {
		template := `if err := is.listen{_methodName}(ctx); err != nil {
			return err
		}`
		g.P(strings.Replace(template, "{_methodName}", m.GoName, -1))
	}
	g.P("return nil")
	g.P("}")

	pg.genSubscriberInterceptor(svcName)
	// Create listen function
	for _, m := range svc.Methods {
		opt, _ := getSubOption(m)
		g.P("func (is *inner", svcName, "Subscriber) listen", m.GoName, "(ctx context.Context) error {")
		g.P("subscriptionName := ", `"`, opt.Subscription, `"`)
		g.P("topicName := ", `"`, opt.Topic, `"`)
		g.P(`var sub *pubsub.Subscription
		if err := retry.Do(func() error {
			tmp, err := is.createSubscriptionIFNotExists(ctx, topicName, subscriptionName)
			if err != nil {
				return err
			}
			sub = tmp
			return nil
		}); err != nil {
			return err
		}`)
		g.P("//", "TODO: メッセージの処理時間の延長を実装する必要がある")
		g.P(fmt.Sprintf(`info := newInnerSubscriberInfo(topicName, subscriptionName, sub, "%s")`, m.GoName))
		// https://christina04.hatenablog.com/entry/cloud-pubsub
		g.P("callback := func(ctx context.Context, msg *pubsub.Message) {")
		g.P("var event ", m.Input.GoIdent)
		g.P(`if err := proto.Unmarshal(msg.Data, &event); err != nil {
		msg.Nack()
		return
		}`)

		g.P(`
		err := is.chainInterceptors(info, msg)
		if err != nil {
			msg.Nack()
			return
		}`)

		g.P("if err := is.service.", m.GoName, "(ctx, &event); err != nil {")
		g.P(`msg.Nack()
		return
		}
		msg.Ack()
		}
		if err := sub.Receive(ctx, callback);err != nil {
			return err
		}
		return nil
		}`)
	}

	return g
}

func (pg *pubsubGenerator) generatePublisherFile() *protogen.GeneratedFile {
	g := pg.g
	pg.decrearePackageName()
	if len(pg.file.Services) == 0 {
		return g
	}
	svc := pg.file.Services[0]

	svcName := svc.GoName
	genClientCode(svcName, svc.Methods, g)
	// Create Publish Function
	genCreateTopicFunction(g)

	return g
}

func getSubOption(m *protogen.Method) (*option.SubOption, error) {
	options := m.Desc.Options().(*descriptorpb.MethodOptions)
	ext := proto.GetExtension(options, option.E_SubOption)
	opt, ok := ext.(*option.SubOption)
	if !ok {
		return nil, errors.New("no pubsub option")
	}
	return opt, nil
}

func getPubOption(m *protogen.Method) (*option.PubOption, error) {
	options := m.Desc.Options().(*descriptorpb.MethodOptions)
	ext := proto.GetExtension(options, option.E_PubOption)
	opt, ok := ext.(*option.PubOption)
	if !ok {
		return nil, errors.New("no pubsub option")
	}
	return opt, nil
}

// Client Code生成
func genClientCode(svcName string, methods []*protogen.Method, g *protogen.GeneratedFile) {

	g.P("type ", svcName, "Client interface {")
	for _, m := range methods {
		g.P("Publish", m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") (string, error)")
	}
	g.P("}")

	template := `
	type inner{{.Name}}Client struct {
		client *pubsub.Client
	}

	func New{{.Name}}Client(client *pubsub.Client) *inner{{.Name}}Client {
		return &inner{{.Name}}Client{
			client: client,
		}
	}

	var topicCache = map[string]*pubsub.Topic{}

	func (c *innerHelloWorldServiceClient) getTopic(topic string) (*pubsub.Topic, error) {
		if t, ok := topicCache[topic]; ok {
			return t, nil
		}
		t, err := GetOrCreateTopicIfNotExists(c.client, topic)
		if err != nil {
			return nil, err
		}
		topicCache[topic] = t
		return t, nil
	}


	func (c *inner{{.Name}}Client) publish(topic string, event protoreflect.ProtoMessage) (string, error) {
		ctx := context.Background()

		t, err := c.getTopic(topic)
		if err != nil {
			return "", err
		}

		ev, err := proto.Marshal(event)
		if err != nil {
			return "", err
		}

		result := t.Publish(ctx, &pubsub.Message{
			Data: ev,
		})
		id, err := result.Get(ctx)
		if err != nil {
			return "", err
		}
		return id, nil
	}
	`
	constructor := strings.ReplaceAll(template, "{{.Name}}", svcName)
	g.P(constructor)

	for _, m := range methods {
		opt, _ := getPubOption(m)
		g.P("func (c *inner", svcName, "Client) Publish", m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") (string, error) {")
		g.P("return c.publish(", `"`, opt.Topic, `"`, ", req)")
		g.P("}")
	}
}

func (pg *pubsubGenerator) genSubscriberInterceptor(svcName string) {
	template := `type SubscriberInfo interface {
		GetTopicName() string
		GetSubscriptionName() string
		GetSubscription() *pubsub.Subscription
		GetMethod() string
	}

	type innerSubscriberInfo struct {
		topicName string
		subscriptionName string
		subscription *pubsub.Subscription
		method string
	}

	func newInnerSubscriberInfo(topicName string, subscriptionName string, subscription *pubsub.Subscription, method string) *innerSubscriberInfo {
		return &innerSubscriberInfo{
			topicName: topicName,
			subscriptionName: subscriptionName,
			subscription: subscription,
			method: method,
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

	type SubscriberInterceptor func(ctx context.Context, msg interface{}, info SubscriberInfo) error

	type inner{_svcName}Subscriber struct {
		service HelloWorldSubscriber
		client *pubsub.Client
		interceptors []SubscriberInterceptor
	}

	func newInner{_svcName}Subscriber(service HelloWorldSubscriber, client *pubsub.Client, interceptors ...SubscriberInterceptor) *inner{_svcName}Subscriber {
		return &inner{_svcName}Subscriber{
			service: service,
			client: client,
			interceptors: interceptors,
		}
	}

	func (is *inner{_svcName}Subscriber) createSubscriptionIFNotExists(ctx context.Context, topicName string, subscriptionName string) (*pubsub.Subscription, error) {
		t := is.client.Topic(topicName)
		if exsits, err := t.Exists(ctx); !exsits {
			return nil, fmt.Errorf("topic does not exsit: %w", err)
		}
		sub := is.client.Subscription(subscriptionName)
		if exsits, _ := sub.Exists(ctx); !exsits {
			is.client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
				Topic:       t,
				AckDeadline: 60 * time.Second,
			})
		}
		return sub, nil
	}

	func (is *inner{_svcName}Subscriber)chainInterceptors(info SubscriberInfo, msg interface{}) error {
		errors := []error{}
		hasError := false
		for i := len(is.interceptors) - 1; i >= 0; i-- {
			chain := is.interceptors[i]
			err := chain(context.Background(), msg, info)
			errors = append(errors, err)
			if err != nil {
				hasError = true
			}
		}
		if hasError {
			return fmt.Errorf("error occured in interceptors: %v", errors)
		}
		return nil
	}`
	pg.g.P(strings.ReplaceAll(template, "{_svcName}", svcName))
}

func genCreateTopicFunction(g *protogen.GeneratedFile) {
	funcString := `
	// GetOrCreateTopicIfNotExists: topicが存在しない場合は作成する
	func GetOrCreateTopicIfNotExists(c *pubsub.Client, topic string) (*pubsub.Topic, error) {
		ctx := context.Background()
		t := c.Topic(topic)
		ok, err := t.Exists(ctx)
		if err != nil {
			return nil, err
		}
		if ok {
			return t, nil
		}
		t, err = c.CreateTopic(ctx, topic)
		if err != nil {
			return nil, err
		}
		return t, nil
	}`
	g.P(funcString)
}
