# TODO: create docker file

FROM golang:1.19
COPY . /src
WORKDIR /src

# Build the Go app
# RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 \
    # go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

# CMD [ "echo", "hello world" ]
