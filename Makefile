generate_option:
	protoc --go_out=./option --go_opt=paths=source_relative  ./proto/option.proto
	mv ./option/proto/option.pb.go ./option/option.pb.go
	rm -rf ./option/proto

generate_sample:
	mkdir -p ./example/proto/option
	cp proto/option.proto ./example/proto/option/
	protoc \
	-I ./example/proto \
	--go_out=./example/generated \
	--go-pubsub_out=./example/generated \
	./example/proto/event.proto

# make test
# make install
# make clean
# make uninstall