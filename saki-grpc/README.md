# 作ってわかる！はじめてのgRPC
- https://zenn.dev/hsaki/books/golang-grpc-starting

## RPC
- Remote Procedure Callの略
- ローカルからリモートの関数を呼び出すもの全般を指す
- `main.go`でmain関数とhello関数を定義していてmain関数からhello関数を呼び出しているとき、Procedure Callという
- それに対して、呼び出し元と呼び出されるProcedure(関数)が別の場所・別のサーバー上にあるときに、RemoteがついてRPC(Remote Procedure Call)という

## gRPC
- RPCを実現するプロトコルの1つ
  - RPCを実現するために様々なプロトコルが考えられた
- Googleが開発したもの
- gRPCがRPCを実現するために使っている技術は大きく分けて以下の2つ
  - HTTP/2
  - Protocol Buffers
- RPCを行うためには以下2つを行う必要がある
  - クライアントからサーバーに呼び出す関数と引数の情報を伝える
  - サーバーからクライアントに戻り値の情報を伝える
- gRPCではHTTP/2のPOSTリクエストとそのレスポンスを使って実現
  - 呼び出す関数の情報:HTTPリクエストのパスに含める
  - 呼び出す関数に渡す引数:HTTPリクエストボディに含める
  - 呼び出した関数の戻り値:HTTPレスポンスボディに含める
- 呼び出した関数の引数・戻り値の情報は、Protocol Buffersというシリアライズ方式を用いて変換した内容をリクエスト・レスポンスボディに含める

## protoファイル書き方
- proto2,proto3のバージョンが存在していて、proto3を使うには明示的にバージョン指定をする必要がある
```
syntax = "proto3";
```
- packageはGoと同じ扱い(他のprotoファイルで定義された型をパッケージ名.型名で参照可能)
```
package myapp;
```
- gRPCで呼び出そうとするProcedure(関数)をメソッド、そしてそのメソッドをいくつかまとめてひとくくりにしたものをサービスという
  - 以下の例では2つのことを行っている
    - 引数にHelloRequest型、戻り値にHelloResponse型を持つメソッドHelloを定義
    - Helloメソッド一つを持つGreetingServiceサービスを定義
```
// サービスの定義
service GreetingService {
	// サービスが持つメソッドの定義
	rpc Hello (HelloRequest) returns (HelloResponse); 
}
```
- 上記のHelloRequest, HelloResponse型を以下のように定義
```
// 型の定義
message HelloRequest {
	string name = 1;
}

message HelloResponse {
	string message = 1;
}
```
- string以外にもint, bool, enumなどprotobufにはいろいろ用意されている
- 他にもGoogleが定義してパッケージとして公開した便利型の集合であるWell Known Typesというものもある
### コード自動生成
- gRPC通信を実装したgoのコードをprotoファイルから自動生成する
- `protoc`コマンドを使う
  - `brew install protoc`
  - `brew install protoc-gen-go`
  - `brew install protoc-gen-go-grpc`
```
protoc --go_out=../pkg/grpc --go_opt=paths=source_relative \
	--go-grpc_out=../pkg/grpc --go-grpc_opt=paths=source_relative \
	hello.proto
```
- オプションの説明は以下
  - `--go_out` : hello.pb.goファイルの出力先
  - `--go_opt` : hello.pb.goファイル生成時のオプション
  - `--go-grpc_out` : hello_grpc.pb.goファイルの出力先
  - `--go-grpc_opt` : hello_grpc.pb.goファイルの生成時オプション
- このコマンドによって以下2つ生成される
  - `hello.pb.go`:protoファイルから自動生成されたリクエスト/レスポンス型を定義した部分のコード
    - 構造体とそのゲッターなどが作成される
  - `hello_grpc.pb.go`:protoファイルから自動生成されたサービス部分のコード
    - サーバーサイド用コードとクライアント用コードが作成される
- サーバーサイド用コードは以下
  - protoのserviceがgoのinterfaceになり、protoのrpcがgoのメソッドになっている
  - `RegisterGreetingServiceServer`は第一引数で渡したServer上で第二引数で渡したgRPCサービスを動かすための関数
```go
// 生成されたGoのコード
type GreetingServiceServer interface {
	// サービスが持つメソッドの定義
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	mustEmbedUnimplementedGreetingServiceServer()
}

func RegisterGreetingServiceServer(s grpc.ServiceRegistrar, srv GreetingServiceServer)
```
- クライアント用コードは以下
```go
// リクエストを送るクライアントを作るコンストラクタ
func NewGreetingServiceClient(cc grpc.ClientConnInterface) GreetingServiceClient {
	return &greetingServiceClient{cc}
}

// クライアントが呼び出せるメソッド一覧をインターフェースで定義
type GreetingServiceClient interface {
	Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
}

```
- 以下公式
  - protoファイル上のメッセージ型がどんなgoの型になるのか
    - https://protobuf.dev/reference/go/go-generated/
  - protoファイル上のメソッド定義がどんなサーバー/クライアント用のコードになるのか
    - https://grpc.io/docs/languages/go/generated-code/
## 自分用メモ
### HTTP/2とは
- 簡潔にHTTP/2の特徴は以下
  - ストリームという概念を導入したことでHTTP/1.1に比べて効率的に通信可能
  - ヘッダー圧縮を行うことでHTTP/1.1より通信量を減らすことが可能
  - サーバープッシュによりリクエストされる前にサーバーからリソースを送信することでHTTP/1.1よりラウンドトリップ回数を減らせる
    - ラウンドトリップとは通信において、ある場所からある場所へ何かを送信したときにそれが返ってくるまでの時間のこと
- HTTP/1.1では1回目のリクエスト->1回目のレスポンス->2回目のリクエスト->2回目のレスポンスというように一度リクエストを送信するとそのレスポンスを受け取ってから次のリクエストを行う必要がある
- HTTP/2ではストリームを導入することでリクエストとレスポンスを並列で行える

### Protocol Buffers
- 構造化データを定義するIDLとしての役割とその構造化データをネットワーク経由で送信可能なバイト列へシリアライズする機能、またその逆のシリアライズする機能を備えている
  - Protocol Buffersはデータをシリアライズする際にJSONやXMLのようなテキスト形式ではなくバイナリ形式にシリアライズする
  - IDL(インタフェース記述言語)(Interface Description Language)
  - .protoファイル
    - vscodeでprotocol buffersの拡張機能をインストールすると.protoファイルに色をつけられる
### シリアライズとは
- シリアライズ(直列化)とは、複数の並列データを直列化して送信すること