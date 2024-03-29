package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	option "github.com/TakaWakiyama/protoc-gen-go-pubsub/option"
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
	pg.generatePublisherOption()
	for _, svc := range pg.file.Services {
		pg.genClientCode(svc.GoName, svc.Methods, g)
	}
	return g
}

func (pg *pubsubGenerator) generatePublisherOption() {
	optionDec := `
	type BatchPublishResult struct {
		ID    string
		Error error
	}

	// ClientOption is the option for publisher client
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
	pg.g.P(optionDec)

}

// Client Code生成
func (pg *pubsubGenerator) genClientCode(svcName string, methods []*protogen.Method, g *protogen.GeneratedFile) {

	g.P("type ", svcName, "Client interface {")
	for _, m := range methods {
		opt, _ := getPubOption(methods[0])
		var batchPublish = false
		if opt.BatchPublish != nil {
			batchPublish = *opt.BatchPublish
		}
		if batchPublish {
			g.P("BatchPublish", m.GoName, "(ctx context.Context, req []*", g.QualifiedGoIdent(m.Input.GoIdent), ") ([]*BatchPublishResult, error)")
		} else {
			g.P("Publish", m.GoName, "(ctx context.Context, req *", g.QualifiedGoIdent(m.Input.GoIdent), ") (string, error)")
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

	func (c *inner{{.Name}}Client) getTopic(topicName string) (*pubsub.Topic, error) {
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

	func (c *inner{{.Name}}Client) getPublisher(topicName string) (gopub.Publisher, error) {
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

	func (c *inner{{.Name}}Client) batchPublish(topic string, events []protoreflect.ProtoMessage) ([]BatchPublishResult, error) {
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
			g.P("func (c *inner", svcName, "Client) BatchPublish", m.GoName, "(ctx context.Context, req []*", g.QualifiedGoIdent(m.Input.GoIdent), ") ([]BatchPublishResult, error) {")
			g.P(`events := make([]protoreflect.ProtoMessage, len(req))
			for i, r := range req {
				events[i] = r
			}`)
			g.P("return c.batchPublish(", `"`, opt.Topic, `"`, ", events)")
			g.P("}")
		} else {
			g.P("func (c *inner", svcName, "Client) Publish", m.GoName, "(ctx context.Context, req *", g.QualifiedGoIdent(m.Input.GoIdent), ") (string, error) {")
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
	for _, svc := range pg.file.Services {
		pg.generateSubscriberInterface(svc)
		pg.generateEntryPoint(svc)
		pg.generateInnerSubscriber(svc)
		pg.generateEachSubscribeFunction(svc)
		pg.generatePubSubAccessorInterface(svc)
		pg.generatePubSubAccessorImpl(svc)
	}

	return g
}

func (pg *pubsubGenerator) generateSubscriberInterface(svc *protogen.Service) {
	pg.g.P("type ", svc.GoName, " interface {")
	for _, m := range svc.Methods {
		pg.g.P(m.Comments.Leading,
			m.GoName, "(ctx context.Context, req *", pg.g.QualifiedGoIdent(m.Input.GoIdent), ") error ")
	}
	pg.g.P("}")
}

func (pg *pubsubGenerator) generateSubscriberOption() {
	pg.g.P(`// SubscriberOption is the option for subscriber.
	type SubscriberOption struct {
		// Interceptors is the slice of SubscriberInterceptor. call before and after subscriber method. default is empty.
		Interceptors []gosub.SubscriberInterceptor
		// SubscribeGracefully is the flag to stop subscribing gracefully. default is false.
		SubscribeGracefully bool
	}

	var defaultSubscriberOption = &SubscriberOption{
		Interceptors:        []gosub.SubscriberInterceptor{},
		SubscribeGracefully: true,
	}

	var retryOpts = []retry.Option{
		retry.Delay(1 * time.Second),
		retry.Attempts(3),
	}
	`)
}

func (pg *pubsubGenerator) generateEntryPoint(svc *protogen.Service) {
	template := `func Run{_svcName}(service {_svcName}, client *pubsub.Client, option *SubscriberOption) error {
		if service == nil {
			return fmt.Errorf("service is nil")
		}
		if client == nil {
			return fmt.Errorf("client is nil")
		}
		if option == nil {
			option = defaultSubscriberOption
		}
		is := newInner{_svcName}(service, client, option)
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error)
		%s
		err := <-errChan
		cancel()
		return err
	}`
	template = strings.Replace(template, "{_svcName}", svc.GoName, -1)
	fs := make([]string, 0, len(svc.Methods))
	for _, m := range svc.Methods {
		l := `
		go func() {
			if err := is.listen{_methodName}(ctx); err != nil {
				errChan <- err
			}
		}()`
		l = strings.Replace(l, "{_methodName}", m.GoName, -1)
		fs = append(fs, l)
	}
	out := fmt.Sprintf(template, strings.Join(fs, "\n"))
	pg.g.P(out)
}

func (pg *pubsubGenerator) generateInnerSubscriber(svc *protogen.Service) {
	template := `
	type inner{_svc.Name} struct {
		service {_svc.Name}
		client  *pubsub.Client
		option  SubscriberOption
		accessor {_svc.Name}PubSubAccessor
	}

	func newInner{_svc.Name}(service {_svc.Name}, client *pubsub.Client, option *SubscriberOption) *inner{_svc.Name} {
		if option == nil {
			option = defaultSubscriberOption
		}
		return &inner{_svc.Name}{
			service: service,
			client:  client,
			option:  *option,
			accessor: New{_svc.Name}PubSubAccessor(),
		}
	}
	`
	template = strings.Replace(template, "{_svc.Name}", svc.GoName, -1)
	pg.g.P(template)
}

func (pg *pubsubGenerator) generateEachSubscribeFunction(svc *protogen.Service) {

	template := `func (is *inner{_svcName}) listen{_m.GoName}(ctx context.Context) error {
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
	}, retryOpts...); err != nil {
		return err
	}

	// Note: you can see ReceiveSettings https://github.com/googleapis/google-cloud-go/blob/pubsub/v1.31.0/pubsub/subscription.go#L720
	// Note: you can see default value  https://github.com/googleapis/google-cloud-go/blob/main/pubsub/subscription.go#L743
	sub.ReceiveSettings.Synchronous = {_opt.Synchronous}
	sub.ReceiveSettings.NumGoroutines = {_opt.NumGoroutines}
	sub.ReceiveSettings.MaxOutstandingMessages = {_opt.MaxOutstandingMessages}

	callback := func(ctx context.Context, msg *pubsub.Message) {
		defer func() {
			if err := recover(); err != nil {
				msg.Nack()
			}
		}()
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

	for _, m := range svc.Methods {
		opt, _ := getSubOption(m)
		template := strings.Replace(template, "{_svcName}", svc.GoName, -1)
		template = strings.Replace(template, "{_m.GoName}", m.GoName, -1)
		template = strings.Replace(template, "{_m.Input}", pg.g.QualifiedGoIdent(m.Input.GoIdent), -1)
		template = strings.Replace(template, "{_opt.Topic}", opt.Topic, -1)
		template = strings.Replace(template, "{_opt.Subscription}", opt.Subscription, -1)
		synchronous := false
		if opt.Synchronous != nil {
			synchronous = *opt.Synchronous
		}
		var numGoroutines int32 = 10
		if opt.NumGoroutines != nil {
			numGoroutines = *opt.NumGoroutines
		}
		var maxOutstandingMessages int32 = 1000
		if opt.MaxOutstandingMessages != nil {
			maxOutstandingMessages = *opt.MaxOutstandingMessages
		}
		template = strings.Replace(template, "{_opt.Synchronous}", strconv.FormatBool(synchronous), -1)
		template = strings.Replace(template, "{_opt.NumGoroutines}", strconv.FormatInt(int64(numGoroutines), 10), -1)
		template = strings.Replace(template, "{_opt.MaxOutstandingMessages}", strconv.FormatInt(int64(maxOutstandingMessages), 10), -1)

		pg.g.P(template)
	}
}

func (pg *pubsubGenerator) generatePubSubAccessorInterface(svc *protogen.Service) {
	template := `
	// {svc.name}PubSubAccessor: Interface for create topic and subscription
	type {svc.name}PubSubAccessor interface {
	%s
	}
	`
	template = strings.Replace(template, "{svc.name}", svc.GoName, -1)
	fs := make([]string, 0, len(svc.Methods))
	for _, m := range svc.Methods {
		f := `
		Create{_m.GoName}TopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error)
		Create{_m.GoName}SubscriptionIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Subscription, error)`
		f = strings.Replace(f, "{_m.GoName}", m.GoName, -1)
		fs = append(fs, f)
	}

	out := fmt.Sprintf(template, strings.Join(fs, "\n"))
	pg.g.P(out)
}

func (pg *pubsubGenerator) generatePubSubAccessorImpl(svc *protogen.Service) {
	impl := `
	type inner{_svc.Name}PubSubAccessor struct{}

	func New{_svc.Name}PubSubAccessor() {_svc.Name}PubSubAccessor {
		return &inner{_svc.Name}PubSubAccessor{}
	}`
	impl = strings.Replace(impl, "{_svc.Name}", svc.GoName, -1)
	pg.g.P(impl)

	for _, m := range svc.Methods {
		opt, _ := getSubOption(m)
		template := `
		func (c *inner{_svc.Name}PubSubAccessor) Create{_m.GoName}TopicIFNotExists(ctx context.Context, client *pubsub.Client) (*pubsub.Topic, error) {
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

		func (c *inner{_svc.Name}PubSubAccessor) Create{_m.GoName}SubscriptionIFNotExists(
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
					EnableExactlyOnceDelivery: {_opt.EnableExactlyOnceDelivery},
				})
			}
			return sub, nil
		}`

		template = strings.Replace(template, "{_svc.Name}", svc.GoName, -1)
		template = strings.Replace(template, "{_opt.Topic}", opt.Topic, -1)
		template = strings.Replace(template, "{_opt.Subscription}", opt.Subscription, -1)
		template = strings.Replace(template, "{_m.GoName}", m.GoName, -1)
		var ackDeadlineSeconds uint32 = 60
		if opt.AckDeadlineSeconds != nil {
			ackDeadlineSeconds = *opt.AckDeadlineSeconds
		}
		template = strings.Replace(template, "{_opt.AckDeadlineSeconds}", fmt.Sprintf("%d", ackDeadlineSeconds), -1)
		var enableExactlyOnceDelivery bool
		if opt.EnableExactlyOnceDelivery != nil {
			enableExactlyOnceDelivery = *opt.EnableExactlyOnceDelivery
		}
		template = strings.Replace(template, "{_opt.EnableExactlyOnceDelivery}", fmt.Sprintf("%t", enableExactlyOnceDelivery), -1)
		pg.g.P(template)
	}
}
