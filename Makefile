generate_option:
	protoc --go_out=./option --go_opt=paths=source_relative  ./proto/option.proto
	mv ./option/proto/option.pb.go ./option/option.pb.go
	rm -rf ./option/proto

copy_option:
	mkdir -p ./example/p
	cp proto/option.proto ./example/proto/option/

generate_sample_subscriber:
	protoc \
	-I ./example/proto \
	--experimental_allow_proto3_optional \
	--go_out=./example/generated \
	--go_opt=paths=source_relative \
	--go-pubsub_out=./example/generated \
	--go-pubsub_opt=paths=source_relative \
	--go-pubsub_opt=is_publisher=0 \
	./example/proto/sub.proto

generate_sample_publisher:
	protoc \
	-I ./example/proto \
	--experimental_allow_proto3_optional \
	--go_out=./example/generated \
	--go_opt=paths=source_relative \
	--go-pubsub_out=./example/generated \
	--go-pubsub_opt=paths=source_relative \
	--go-pubsub_opt=is_publisher=1 \
	./example/proto/pub.proto

generate_sample_option:
	protoc \
	-I ./example/proto \
	--experimental_allow_proto3_optional \
	--go_out=./example/generated \
	--go_opt=paths=source_relative \
	./example/proto/option.proto


phony: generate_option copy_option generate_sample_subscriber generate_sample_publisher generate_sample_option