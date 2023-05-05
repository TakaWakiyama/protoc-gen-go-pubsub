package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TakaWakiyama/protoc-gen-go-pubsub/option"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	GO_IMPORT_GOSUB        = `gosub "github.com/TakaWakiyama/protoc-gen-go-pubsub/subscriber"`
	GO_IMPORT_GOPUB        = `gopub "github.com/TakaWakiyama/protoc-gen-go-pubsub/publisher"`
)

const (
	// FileDescriptorProto.package field number
	fileDescriptorProtoPackageFieldNumber = 2
	// FileDescriptorProto.syntax field number
	fileDescriptorProtoSyntaxFieldNumber = 12
)

var allImportSubMods = []string{
	GO_IMPORT_CONTEXT,
	GO_IMPORT_FMT,
	GO_IMPORT_TIME,
	GO_IMPORT_PUBSUB,
	GO_IMPORT_GOSUB,
	GO_IMPORT_RETRY,
	GO_IMPORT_PROTO,
}

var allImportPubMods = []string{
	GO_IMPORT_CONTEXT,
	GO_IMPORT_TIME,
	GO_IMPORT_PUBSUB,
	GO_IMPORT_GOPUB,
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

/* Common Parts*/
func (pg *pubsubGenerator) decrearePackageName() {
	g := pg.g
	g.P("// Code generated  by protoc-gen-go-event. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-go-pubsub v", version)
	g.P("// - protoc             ", protocVersion(pg.gen))
	g.P("// source: ", pg.file.Proto.GetName())
	g.P()
	genLeadingComments(g, pg.file.Desc.SourceLocations().ByPath(protoreflect.SourcePath{fileDescriptorProtoPackageFieldNumber}))
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

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

func genLeadingComments(g *protogen.GeneratedFile, loc protoreflect.SourceLocation) {
	for _, s := range loc.LeadingDetachedComments {
		g.P(protogen.Comments(s))
		g.P()
	}
	if s := loc.LeadingComments; s != "" {
		g.P(protogen.Comments(s))
		g.P()
	}
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

/* Publisher Parts */
func (pg *pubsubGenerator) generatePublisherFile() *protogen.GeneratedFile {
	if len(pg.file.Services) == 0 {
		return pg.g
	}
	g := pg.g
	pg.decrearePackageName()
	//
	svc := pg.file.Services[0]

	svcName := svc.GoName
	pg.genClientCode(svcName, svc.Methods, g)
	return g
}

// Client Code生成
func (pg *pubsubGenerator) genClientCode(svcName string, methods []*protogen.Method, g *protogen.GeneratedFile) {

	optionDec := `
	type BatchPublishResult struct {
		ID    string
		Error error
	}

	// ClientOption is the option for HelloWorldServiceClient
	type ClientOption struct {
		// Gracefully is the flag to stop publishing gracefully
		Gracefully bool
		// MaxAttempts is the max attempts when wait for publishing gracefully
		MaxAttempts int
		// Delay is the delay time when wait for publishing gracefully
		Delay time.Duration
	}

	var defaultClientOption = &ClientOption{
		Gracefully:  false,
		MaxAttempts: 3,
		Delay:       1 * time.Second,
	}
	`
	g.P(optionDec)

	g.P("type ", svcName, "Client interface {")
	for _, m := range methods {
		opt, _ := getPubOption(methods[0])
		var batchPublish = false
		if opt.BatchPublish != nil {
			batchPublish = *opt.BatchPublish
		}
		if batchPublish {
			g.P("BatchPublish", m.GoName, "(ctx context.Context, req []*", m.Input.GoIdent, ") ([]*BatchPublishResult, error)")
		} else {
			g.P("Publish", m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") (string, error)")
		}
	}
	g.P("}")

	template := `
	type inner{{.Name}}Client struct {
		client          *pubsub.Client
		nameToTopic     map[string]*pubsub.Topic
		nameToPublisher map[string]gopub.Publisher
		option          ClientOption
	}

	func New{{.Name}}Client(client *pubsub.Client, option *ClientOption) *inner{{.Name}}Client {
		if option == nil {
			option = defaultClientOption
		}
		return &inner{{.Name}}Client{
			client:          client,
			nameToTopic:     make(map[string]*pubsub.Topic),
			nameToPublisher: make(map[string]gopub.Publisher),
			option:          *option,
		}
	}

	func (c *innerHelloWorldServiceClient) getTopic(topicName string) (*pubsub.Topic, error) {
		if t, ok := c.nameToTopic[topicName]; ok {
			return t, nil
		}
		t, err := gopub.GetOrCreateTopicIfNotExists(c.client, topicName)
		if err != nil {
			return nil, err
		}
		c.nameToTopic[topicName] = t
		return t, nil
	}

	func (c *innerHelloWorldServiceClient) getPublisher(topicName string) (gopub.Publisher, error) {
		if p, ok := c.nameToPublisher[topicName]; ok {
			return p, nil
		}
		p := gopub.NewPublisher(c.client, &gopub.PublisherOption{
			Gracefully:  c.option.Gracefully,
			MaxAttempts: uint(c.option.MaxAttempts),
			Delay:       c.option.Delay,
		})
		c.nameToPublisher[topicName] = p
		return p, nil
	}


	func (c *inner{{.Name}}Client) publish(topic string, event protoreflect.ProtoMessage) (string, error) {
		ctx := context.Background()

		t, err := c.getTopic(topic)
		if err != nil {
			return "", err
		}
		p, err := c.getPublisher(topic)
		if err != nil {
			return "", err
		}

		ev, err := proto.Marshal(event)
		if err != nil {
			return "", err
		}
		return p.Publish(ctx, t, &pubsub.Message{
			Data: ev,
		})
	}

	func (c *innerHelloWorldServiceClient) batchPublish(topic string, events []protoreflect.ProtoMessage) ([]BatchPublishResult, error) {
		ctx := context.Background()

		t, err := c.getTopic(topic)
		if err != nil {
			return nil, err
		}
		p, err := c.getPublisher(topic)
		if err != nil {
			return nil, err
		}

		var msgs []*pubsub.Message
		for _, e := range events {
			ev, err := proto.Marshal(e)
			if err != nil {
				return nil, err
			}
			msgs = append(msgs, &pubsub.Message{
				Data: ev,
			})
		}
		res, err := p.BatchPublish(ctx, t, msgs)
		if err != nil {
			return nil, err
		}
		var results []BatchPublishResult
		for _, r := range res {
			results = append(results, BatchPublishResult{
				ID:    r.ID,
				Error: r.Error,
			})
		}
		return results, nil
	}
	`
	constructor := strings.ReplaceAll(template, "{{.Name}}", svcName)
	g.P(constructor)

	for _, m := range methods {
		opt, _ := getPubOption(m)
		var batchPublish = false
		if opt.BatchPublish != nil {
			batchPublish = *opt.BatchPublish
		}
		if batchPublish {
			g.P("func (c *inner", svcName, "Client) BatchPublish", m.GoName, "(ctx context.Context, req []*", m.Input.GoIdent, ") ([]*BatchPublishResult, error) {")
			g.P("return c.batchPublish(", `"`, opt.Topic, `"`, ", req)")
			g.P("}")
		} else {
			g.P("func (c *inner", svcName, "Client) Publish", m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") (string, error) {")
			g.P("return c.publish(", `"`, opt.Topic, `"`, ", req)")
			g.P("}")
		}
	}
}

/* Subscriber */
func (pg *pubsubGenerator) generateSubscriberFile() *protogen.GeneratedFile {
	if len(pg.file.Services) == 0 {
		return pg.g
	}

	g := pg.g
	pg.decrearePackageName()
	genLeadingComments(g, pg.file.Desc.SourceLocations().ByPath(protoreflect.SourcePath{fileDescriptorProtoSyntaxFieldNumber}))
	pg.generateSubscriberOption()
	pg.generateSubscriberInterface()
	// Note: 複数サービスのユースケースを確認する
	svc := pg.file.Services[0]
	pg.generateEntryPoint(svc)
	pg.generateInnerSubscriber(svc)
	pg.generateEachSubscribeFunction()
	ms := pg.file.Services[0].Methods
	pg.generatePubSubAccessorInterface(ms)
	pg.generatePubSubAccessorImpl(ms)

	return g
}

func (pg *pubsubGenerator) generateSubscriberInterface() {
	svc := pg.file.Services[0]
	pg.g.P("type ", svc.GoName, " interface {")
	for _, m := range svc.Methods {
		pg.g.P(m.Comments.Leading,
			m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") error ")
	}
	pg.g.P("}")
}

func (pg *pubsubGenerator) generateSubscriberOption() {
	pg.g.P(`// SubscriberOption is the option for HelloWorldSubscriber
	type SubscriberOption struct {
		// Interceptors is the slice of SubscriberInterceptor. call before and after HelloWorldSubscriber method. default is empty.
		Interceptors []gosub.SubscriberInterceptor
		// SubscribeGracefully is the flag to stop subscribing gracefully. default is false.
		SubscribeGracefully bool
	}

	var defaultSubscriberOption = &SubscriberOption{
		Interceptors:        []gosub.SubscriberInterceptor{},
		SubscribeGracefully: false,
	}`)
}

func (pg *pubsubGenerator) generateEntryPoint(svc *protogen.Service) {
	template := `func Run(service {_svcName}, client *pubsub.Client, option *SubscriberOption) error {
		if service == nil {
			return fmt.Errorf("service is nil")
		}
		if client == nil {
			return fmt.Errorf("client is nil")
		}
		if option == nil {
			option = defaultSubscriberOption
		}
		ctx := context.Background()
		is := newInner{_svcName}Subscriber(service, client, option)
		%s
		return nil
	}
	`
	template = strings.Replace(template, "{_svcName}", svc.GoName, -1)
	fs := make([]string, 0, len(svc.Methods))
	for _, m := range svc.Methods {
		l := `if err := is.listen{_methodName}(ctx); err != nil {
			return err
		}`
		l = strings.Replace(l, "{_methodName}", m.GoName, -1)
		fs = append(fs, l)
	}
	out := fmt.Sprintf(template, strings.Join(fs, "\n"))
	pg.g.P(out)
}

func (pg *pubsubGenerator) generateInnerSubscriber(svc *protogen.Service) {
	template := `
	type inner{_svc.Name}Subscriber struct {
		service HelloWorldSubscriber
		client  *pubsub.Client
		option  SubscriberOption
		accessor PubSubAccessor
	}

	func newInner{_svc.Name}Subscriber(service HelloWorldSubscriber, client *pubsub.Client, option *SubscriberOption) *innerHelloWorldSubscriberSubscriber {
		if option == nil {
			option = defaultSubscriberOption
		}
		return &inner{_svc.Name}Subscriber{
			service: service,
			client:  client,
			option:  *option,
			accessor: NewPubSubAccessor(),
		}
	}
	`
	template = strings.Replace(template, "{_svc.Name}", svc.GoName, -1)
	pg.g.P(template)
}

func (pg *pubsubGenerator) generateEachSubscribeFunction() {

	template := `func (is *inner{_svcName}Subscriber) listen{_m.GoName}(ctx context.Context) error {
	subscriptionName := "{_opt.Subscription}"
	topicName := "{_opt.Topic}"
	var sub *pubsub.Subscription
	if err := retry.Do(func() error {
		tmp, err := is.accessor.Create{_m.GoName}SubscriptionIFNotExists(ctx, is.client)
		if err != nil {
			return err
		}
		sub = tmp
		return nil
	}); err != nil {
		return err
	}
	callback := func(ctx context.Context, msg *pubsub.Message) {
		info := gosub.NewSubscriberInfo(topicName, subscriptionName, sub, "{_m.GoName}", msg)
		var event {_m.Input}
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		if err := gosub.Handle(is.option.Interceptors, ctx, &event, info, func(ctx context.Context, req interface{}) error {
			return is.service.{_m.GoName}(ctx, req.(*{_m.Input}))
		}); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	}

	if is.option.SubscribeGracefully {
		gosub.SubscribeGracefully(sub, ctx, callback, nil)
	} else {
		sub.Receive(ctx, callback)
	}

	return nil
}
	`

	for _, m := range pg.file.Services[0].Methods {
		opt, _ := getSubOption(m)
		template := strings.Replace(template, "{_svcName}", pg.file.Services[0].GoName, -1)
		template = strings.Replace(template, "{_m.GoName}", m.GoName, -1)
		template = strings.Replace(template, "{_m.Input}", m.Input.GoIdent.GoName, -1)
		template = strings.Replace(template, "{_opt.Topic}", opt.Topic, -1)
		template = strings.Replace(template, "{_opt.Subscription}", opt.Subscription, -1)
		pg.g.P(template)
	}
}

func (pg *pubsubGenerator) generatePubSubAccessorInterface(ms []*protogen.Method) {
	template := `
	// PubSubAccessor: accessor for HelloWorldPubSub
	type PubSubAccessor interface {
	%s
	}
	`
	fs := make([]string, 0, len(ms))
	for _, m := range ms {
		f := `
		Create{_m.GoName}TopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error)
		Create{_m.GoName}SubscriptionIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Subscription, error)`
		f = strings.Replace(f, "{_m.GoName}", m.GoName, -1)
		fs = append(fs, f)
	}

	out := fmt.Sprintf(template, strings.Join(fs, "\n"))
	pg.g.P(out)
}

func (pg *pubsubGenerator) generatePubSubAccessorImpl(ms []*protogen.Method) {
	fixed := `
	type pubSubAccessorImpl struct{}

	func NewPubSubAccessor() PubSubAccessor {
		return &pubSubAccessorImpl{}
	}`
	pg.g.P(fixed)

	for _, m := range ms {
		opt, _ := getSubOption(m)
		template := `
		func (c *pubSubAccessorImpl) Create{_m.GoName}TopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error) {
			topicName := "{_opt.Topic}"
			t := client.Topic(topicName)
			exsits, err := t.Exists(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to check topic exsits: %w", err)
			}
			if !exsits {
				return client.CreateTopic(ctx, topicName)
			}
			return t, nil
		}

		func (c *pubSubAccessorImpl) Create{_m.GoName}SubscriptionIFNotExists(
			ctx context.Context,
			client *pubsub.Client,
		) (*pubsub.Subscription, error) {
			subscriptionName := "{_opt.Subscription}"
			topicName := "{_opt.Topic}"
			t := client.Topic(topicName)
			if exsits, err := t.Exists(ctx); !exsits {
				return nil, fmt.Errorf("topic does not exsit: %w", err)
			}
			sub := client.Subscription(subscriptionName)
			exsits, err := sub.Exists(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to check subscription exsits: %w", err)
			}
			if !exsits {
				return client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
					Topic:       t,
					AckDeadline: {_opt.AckDeadlineSeconds} * time.Second,
				})
			}
			return sub, nil
		}`
		template = strings.Replace(template, "{_opt.Topic}", opt.Topic, -1)
		template = strings.Replace(template, "{_opt.Subscription}", opt.Subscription, -1)
		template = strings.Replace(template, "{_m.GoName}", m.GoName, -1)
		var ackDeadlineSeconds uint32 = 60
		if opt.AckDeadlineSeconds != nil {
			ackDeadlineSeconds = *opt.AckDeadlineSeconds
		}
		template = strings.Replace(template, "{_opt.AckDeadlineSeconds}", fmt.Sprintf("%d", ackDeadlineSeconds), -1)
		pg.g.P(template)
	}
}
