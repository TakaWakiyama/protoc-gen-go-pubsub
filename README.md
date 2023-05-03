# protoc-gen-go-pubsub

## 概要

このコードは、Go言語でGoogle Cloud Pub/Subのサブスクライバーとパブリッシャーを生成するプラグインを実装するためのものです。

このプラグインは、ProtobufのRPCサービス定義に注釈を追加することで、各メソッドに対するPub/Subトピックとサブスクリプションの情報を提供します。そして、これらの情報を利用して、Pub/Subトピックからデータを取得してRPCメソッドを呼び出すサブスクライバーと、RPCメソッドの引数をPub/Subトピックにパブリッシュするパブリッシャーを生成します。

このコードでは、pubsubGenerator構造体を使用して、サブスクライバーとパブリッシャーそれぞれのコードを生成します。サブスクライバーには、RPCメソッドの情報に基づいてPub/Subトピックからデータを取得してRPCメソッドを呼び出すためのコードが含まれます。パブリッシャーには、RPCメソッドの引数をPub/Subトピックにパブリッシュするためのコードが含まれます。

このコードは、Go言語でのGoogle Cloud Pub/Subの使用に関心がある開発者にとって役立つものです。また、Pub/Subを使用することで、大規模なシステムでの非同期メッセージングの実装を簡素化することができます。

## 使い方

Usage:
  protoc-gen-go-pubsub [options] [directory...]

Options:
  --is_publisher=true|false
        Set to true for generating publisher code, set to false for generating subscriber code. Defaults to false.

```bash
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
```
