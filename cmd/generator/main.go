package main

import (
	"flag"

	"github.com/TakaWakiyama/protoc-gen-go-pubsub/cmd/generator"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	usage := "true if publisher false if subscriber"
	var flags flag.FlagSet
	isPublisher := flags.Bool("is_publisher", false, usage)

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			g := generator.NewFileGenerator(gen, f, *isPublisher)
			g.Generate()
		}
		return nil
	})
}
