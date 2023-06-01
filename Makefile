generate_option:
	mkdir -p ./option
	protoc --go_out=./option --go_opt=paths=source_relative --go_opt=Mproto/option.proto=github.com/TakaWakiyama/protoc-gen-go-pubsub  ./proto/option.proto
	mkdir -p ./cmd/generator/option
	mv ./option/proto/option.pb.go ./cmd/generator/option/option.pb.go
	rm -rf ./option/proto

copy_option:
	mkdir -p ./example/p
	cp proto/option.proto ./example/proto/option.proto

generate_example_subscriber:
	protoc \
	-I ./example/proto \
	--experimental_allow_proto3_optional \
	--go_out=./example/generated \
	--go_opt=paths=source_relative \
	--go_opt=Moption.proto=github.com/TakaWakiyama/protoc-gen-go-pubsub/example \
	--go-pubsub_out=./example/generated \
	--go-pubsub_opt=paths=source_relative \
	--go-pubsub_opt=is_publisher=0 \
	--go-pubsub_opt=Moption.proto=github.com/TakaWakiyama/protoc-gen-go-pubsub/example \
	./example/proto/sub.proto

generate_example_publisher:
	protoc \
	-I ./example/proto \
	--experimental_allow_proto3_optional \
	--go_out=./example/generated \
	--go_opt=paths=source_relative \
	--go_opt=Moption.proto=github.com/TakaWakiyama/protoc-gen-go-pubsub/example \
	--go-pubsub_out=./example/generated \
	--go-pubsub_opt=paths=source_relative \
	--go-pubsub_opt=is_publisher=1 \
	--go-pubsub_opt=Moption.proto=github.com/TakaWakiyama/protoc-gen-go-pubsub/example \
	./example/proto/pub.proto

generate_example_option:
	protoc \
	-I ./example/proto \
	--experimental_allow_proto3_optional \
	--go_out=./example/generated \
	--go_opt=paths=source_relative \
	--go_opt=Moption.proto=github.com/TakaWakiyama/protoc-gen-go-pubsub/example \
	./example/proto/option.proto


generate_example: generate_example_subscriber generate_example_publisher generate_example_option

install:
	cd generator && go install

phony: generate_option copy_option generate_example_subscriber generate_example_publisher generate_example_option install