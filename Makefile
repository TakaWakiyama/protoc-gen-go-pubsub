generate_option:
	protoc --go_out=./option --go_opt=paths=source_relative  ./proto/option.proto
	mv ./option/proto/option.pb.go ./option/option.pb.go
	rm -rf ./option/proto

copy_option:
	mkdir -p ./example/p
	cp proto/option.proto ./example/proto/option/

generate_sample:
	protoc \
	-I ./example/proto \
	--experimental_allow_proto3_optional \
	--go_out=./example/generated \
	--go_opt=paths=source_relative \
	--go-pubsub_out=./example/generated \
	--go-pubsub_opt=paths=source_relative \
	./example/proto/event.proto
