# protoc-gen-go-pubsub

## Overview

This plugin generates Go code from Protobuf for Google Cloud Pub/Sub subscribers and publishers.
It provides flexible behavior by specifying topics, subscription names, and other options with protobuf Options.
It also provides features such as interceptors and graceful shutdown.

## Installation

```bash
git clone github.com/TakaWakiyama/protoc-gen-go-pubsub
make install

Usage:
  protoc-gen-go-pubsub [options] [directory...]

Options:
  --is_publisher=true|false
        Set to true for generating publisher code, set to false for generating subscriber code. Defaults to false.

```

## Usage

1. Copy `option.proto` to your project.
2. Create a proto file for your publisher or subscriber.

  refer to `example/proto/pub.proto` and `example/proto/sub.proto`.

  ```proto
  rpc HelloWorld(HelloWorldEvent) returns (Empty) {
      // see details in example/proto/option.proto
      option (option.subOption) = {
      // if you have not prepared a subscription, you can create it automatically.
      topic: "pubsub_topic_name"
      subscription: "subscription_name"
  };```
3. Generate code with `protoc-gen-go-pubsub`

```bash
# 1. generate pubsub option
generate_example_option:
protoc \
-I ./example/proto \
--experimental_allow_proto3_optional \
--go_out=./example/generated \
--go_opt=paths=source_relative \
--go_opt=Moption.proto=github.com/TakaWakiyama/protoc-gen-go-pubsub/example \ # set your option.proto path
./example/proto/option.proto

# 1. generate pubsub subscriber

generate_example_subscriber:
protoc \
--experimental_allow_proto3_optional \
--go_out=./example/generated \
--go_opt=paths=source_relative \
--go_opt=Moption.proto=github.com/TakaWakiyama/protoc-gen-go-pubsub/example \ # set your option.proto path
--go-pubsub_out=./example/generated \
--go-pubsub_opt=paths=source_relative \
--go-pubsub_opt=is_publisher=0 \ # set 0 generate subscriber code
--go-pubsub_opt=Moption.proto=github.com/TakaWakiyama/protoc-gen-go-pubsub/example \ # set your option.proto path
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
```

4. Use generated code

see `example/main.go`

5. Run Sample Code

```bash
cd example
docker-compose up -d pubsub subscriber
docker-compose run --rm publisher
```

6. Development

```bash
cd generator
make install
make generate_example_{subscriber,publisher, option}
docker-compose run --rm publisher
```

