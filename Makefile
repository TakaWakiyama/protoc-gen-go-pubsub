generate:
	protoc --go_out=./option --go_opt=paths=source_relative  ./proto/option.proto
	mv ./option/proto/option.pb.go ./option/option.pb.go
	rm -rf ./option/proto
# make build
# make test
# make install
# make clean
# make uninstall