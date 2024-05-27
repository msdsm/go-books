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


### gRPCサーバー実装
- `src/cmd/server/main.go`にあるように`grpc.NewServer()`でgRPCサーバーを作成できる
```go
// 2. gRPCサーバーを作成
s := grpc.NewServer()
```
- サーバーにサービスの登録をするにはprotoから生成した`RegisterGreetingServiceServer(s grpc.ServiceRegistrar, srv GreetingServiceServer)`を使う
- `GreetingServiceServer`はinterfaceでソースを見ると以下のようになっている
```go
// GreetingServiceServerインターフェース型の定義
type GreetingServiceServer interface {
	// Helloメソッドを持つ
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	mustEmbedUnimplementedGreetingServiceServer()
}
```
- このinterfadeを実装するために新しい構造体を作成する
```go
type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}
```
- ここで`UnimplementedGreetingServiceServer`はprotocコマンドによって生成された構造体で、`GreetingServiceServer`インターフェースを満たすためのメソッドが定義されている
```go
func (UnimplementedGreetingServiceServer) Hello(context.Context, *HelloRequest) (*HelloResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Hello not implemented")
}
func (UnimplementedGreetingServiceServer) mustEmbedUnimplementedGreetingServiceServer() {}
```
- そのため`Hello`のビジネスロジックを記述すればよい
```go
func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	// リクエストからnameフィールドを取り出して
	// "Hello, [名前]!"というレスポンスを返す
	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}
```
- gRPCサーバーに登録する
```go
hellopb.RegisterGreetingServiceServer(s, NewMyServer())
```
- 動作確認のためにはgRPCurlを使う
  - gRPCの通信はProtocol Bufferでシリアライズされている
  - そのシリアライズ・デシリアライズを行うためにはprotoファイルによって書かれたシリアライズのルールを知る必要がある
  - protoファイルによるメッセージ型の定義を知らないgRPCulコマンドは代わりにgRPCサーバーそのものからprotoファイルの情報を取得することでシリアライズのルールを知り通信する
  - gRPCサーバーそのものからprotoファイルの情報を取得する必要があり、その機能がサーバーリフレクション
```go
// 4. サーバーリフレクションの設定
reflection.Register(s)
```
- gRPCサーバーで稼働しているサービスの確認
  - `grpcurl -plaintext localhost:8080 list`
- サービスが持つメソッド一覧を表示(ここではGreetingService)
  - `grpcurl -plaintext localhost:8080 list myapp.GreetingService`
- `grpcurl -plaintext -d '{"name": "hsaki"}' localhost:8080 myapp.GreetingService.Hello`
をたたくと正しくレスポンスかえってきた　
```
{
  "message": "Hello, hsaki!"
}
```

### gRPCクライアント実装
- gRPCurlコマンドを使わずにプログラムからgRPCサーバーにリクエストを送る
- gRPCのコネクションを得るには`grpc.Dial()`関数を使う
```go
conn, err := grpc.Dial(
	address,

	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithBlock(),
)
```
- 第一引数にgRPCのサーバーアドレスを渡す
  - 第二引数以降はオプション
  - `grpc.WithTransportCredentials(insecure.NewCredentials())` : SSL/TSLを使用しない
  - `grpc.WithBlock()` : コネクションが確立されるまで待機する(同期)
- `hello_grpc.pb.go`で生成された`NewGreetingServiceClient`を使用してクライアントを作成
```go
client = hellopb.NewGreetingServiceClient(conn)
```
- `hello_grpc.pb.go`を見ると以下のようにクライアントはサービスの`Hello`メソッドにリクエストを送信するための`Hello`メソッドを持つことがわかる
```go
type GreetingServiceClient interface {
	// サービスが持つメソッドの定義
	Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
}

type greetingServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGreetingServiceClient(cc grpc.ClientConnInterface) GreetingServiceClient {
	return &greetingServiceClient{cc}
}

func (c *greetingServiceClient) Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error) {
	out := new(HelloResponse)
	err := c.cc.Invoke(ctx, GreetingService_Hello_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
```
- これを使用してリクエストを送信できる
```go
// リクエストに使うHelloRequest型の生成
req := &hellopb.HelloRequest{
	Name: name,
}
// Helloメソッドの実行 -> HelloResponse型のレスポンスresを入手
res, err := client.Hello(context.Background(), req)
if err != nil {
	fmt.Println(err)
} else {
	// resの内容を標準出力に出す
	fmt.Println(res.GetMessage())
}
```
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

### Graceful Shutdown
- アプリ停止の際にリクエストの受付を停止して、処理中のプロセスが完全に終わるまで待ってからアプリを終了すること