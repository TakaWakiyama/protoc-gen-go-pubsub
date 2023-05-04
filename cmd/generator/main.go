package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
)

const version = "1.0.0"

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
			g := NewFileGenerator(gen, f, *isPublisher)
			g.Generate()
		}
		return nil
	})
}
