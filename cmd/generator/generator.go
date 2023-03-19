package generator

import (
	"errors"
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
)

var allImportSubMods = []string{
	GO_IMPORT_CONTEXT,
	GO_IMPORT_PUBSUB,
	GO_IMPORT_PROTO,
	GO_IMPORT_FMT,
	GO_IMPORT_TIME,
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
		g.P(`"`, mod, `"`)
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
	g.P("func Run(service ", svcName, ", client *pubsub.Client) {")
	g.P("ctx := context.Background()")
	for _, m := range svc.Methods {
		g.P("if err := listen", m.GoName, "(ctx, service, client); err != nil {")
		g.P("panic(err)")
		g.P("}")
	}
	g.P("}")
	// Create listen function
	for _, m := range svc.Methods {
		opt, _ := getSubOption(m)
		g.P("func listen", m.GoName, "(ctx context.Context, service ", svcName, ", client *pubsub.Client) error {")
		g.P("subscriptionName := ", `"`, opt.Subscription, `"`)
		g.P("topicName := ", `"`, opt.Topic, `"`)
		g.P(`t := client.Topic(topicName)
		if exsits, err := t.Exists(ctx); !exsits {
			return fmt.Errorf("topic does not exsit: %w", err)
		}
		sub := client.Subscription(subscriptionName)
		if exsits, _ := sub.Exists(ctx); !exsits {
			client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
				Topic:       t,
				// TODO: デフォルトの時間はoptionで設定できるようにした方が良い
				AckDeadline: 60 * time.Second,
			})
		}`)
		g.P("//", "TODO: メッセージの処理時間の延長を実装する必要がある")
		// https://christina04.hatenablog.com/entry/cloud-pubsub
		g.P("callback := func(ctx context.Context, msg *pubsub.Message) {")
		g.P("var event ", m.Input.GoIdent)
		g.P("if err := proto.Unmarshal(msg.Data, &event); err != nil {")
		g.P(`msg.Nack()
		return`)
		// error 処理
		g.P("}")
		g.P("if err := service.", m.GoName, "(ctx, &event); err != nil {")
		// 再送信させる
		g.P(`msg.Nack()
		return
		}
		msg.Ack()`)
		g.P("}")
		g.P("err := pullMsgs(ctx, client, subscriptionName, topicName, callback)")
		g.P("if err != nil {")
		g.P("return err")
		g.P("}")
		g.P("return nil")
		g.P("}")
	}

	// Create pullMsgs function
	g.P(`
	func pullMsgs(ctx context.Context, client *pubsub.Client, subScriptionName, topicName string, callback func(context.Context, *pubsub.Message)) error {
		sub := client.Subscription(subScriptionName)
		err := sub.Receive(ctx, callback)
		if err != nil {
			return err
		}
		return nil
	}
	`)

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

/*
func genSubscriberClientCode() {
	template = ```
	func (s *inner${_svbName}Client) Subscribe(ctx context.Context, req *HelloWorldRequest) error {
	subscriptionName := "helloworldsubscription"
	topicName := "helloworldtopic"
	// TODO: メッセージの処理時間の延長を実装する必要がある
	callback := func(ctx context.Context, msg *pubsub.Message) {
		var event HelloWorldRequest
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			msg.Nack()
			return
		}
		if err := service.HelloWorld(ctx, &event); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	}

	sub := client.Subscription(subScriptionName)
	err := sub.Receive(ctx, callback)
	if err != nil {
		return err
	}
	return nil
```
}
*/

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
