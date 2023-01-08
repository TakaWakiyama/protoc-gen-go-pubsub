package main

import (
	"errors"
	"strings"

	"github.com/TakaWakiyama/protoc-gen-go-pubsub/option"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	GO_IMPORT_FMT          = "fmt"
	GO_IMPORT_CONTEXT      = "context"
	GO_IMPORT_PUBSUB       = "cloud.google.com/go/pubsub"
	GO_IMPORT_PROTO        = "google.golang.org/protobuf/proto"
	GO_IMPORT_PROTOREFLECT = "google.golang.org/protobuf/reflect/protoreflect"
)

var allImportMods = []string{
	GO_IMPORT_FMT,
	GO_IMPORT_CONTEXT,
	GO_IMPORT_PUBSUB,
	GO_IMPORT_PROTO,
	GO_IMPORT_PROTOREFLECT,
}

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f)
		}
		return nil
	})
}

// generateFile generates a _ascii.pb.go file containing gRPC service definitions.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + "_event.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated  by protoc-gen-go-event. DO NOT EDIT.")
	g.P("// versions:")
	// g.P("//  protoc-gen-go ", file) // TODO: get version from protogen
	// g.P("//  protoc        v3.21.9")                            // TODO: get version from protogen
	g.P("// source: ", file.Proto.GetName())
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	g.P("import (")
	for _, mod := range allImportMods {
		g.P(`"`, mod, `"`)
	}
	g.P(")")

	if len(file.Services) == 0 {
		return g
	}
	svc := file.Services[0]

	svcName := svc.GoName
	// Create Client Code

	// Create Server Code
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
		opt, _ := getPubSubOption(m)
		g.P("func listen", m.GoName, "(ctx context.Context, service ", svcName, ", client *pubsub.Client) error {")
		g.P("subscriptionName := ", `"`, opt.Subscription, `"`)
		g.P("topicName := ", `"`, opt.Topic, `"`)
		g.P("//", "TODO: メッセージの処理時間の延長を実装する必要がある")
		// https://christina04.hatenablog.com/entry/cloud-pubsub
		g.P(`callback := func(ctx context.Context, msg *pubsub.Message) {
			msg.Ack()
		`)
		g.P("var event ", m.Input.GoIdent)
		g.P("if err := proto.Unmarshal(msg.Data, &event); err != nil {")
		g.P("fmt.Println(err)")
		// error 処理
		g.P("}")
		g.P("if err := service.", m.GoName, "(ctx, &event); err != nil {")
		// 再送信させる
		g.P("msg.Nack()")
		g.P("}")
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
		// topic := client.Topic(topicName)
		fmt.Printf("topicName: %v\n", topicName)
		err := sub.Receive(ctx, callback)
		if err != nil {
			return err
		}
		return nil
	}
	`)

	// Create Client Code
	genClientCode(svcName, svc.Methods, g)
	// Create Publish Function
	genCreateTopicFunction(g)

	return g
}

func getPubSubOption(m *protogen.Method) (*option.PubSubOption, error) {
	options := m.Desc.Options().(*descriptorpb.MethodOptions)
	ext := proto.GetExtension(options, option.E_PubSubOption)
	opt, ok := ext.(*option.PubSubOption)
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

	func (c *inner{{.Name}}Client) publish(topic string, event protoreflect.ProtoMessage) (string, error) {
		ctx := context.Background()
		// TODO: メモリに持っておく
		t := c.client.Topic(topic)

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
		opt, _ := getPubSubOption(m)
		g.P("func (c *inner", svcName, "Client) Publish", m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") (string, error) {")
		g.P("return c.publish(", `"`, opt.Topic, `"`, ", req)")
		g.P("}")
	}
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
