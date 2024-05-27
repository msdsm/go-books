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

### ストリーミング処理
- gRPCで可能な4種類の通信方式
  - Unary RPC
    - 1リクエスト1レスポンスの通信方式
  - Server streaming RPC
    - 1リクエスト複数レスポンスの通信方式
    - サーバー側からプッシュ通知を受け取るなど
  - Client streaming RPC
    - 複数リクエスト1レスポンスの通信方式
    - クライアント側から複数回に分けてデータをアップロードして全て受け取るとサーバーが1回だけokと返すなど
  - Bidirectional streaming RPC
    - サーバー・クライアントともに任意のタイミングでリクエスト・レスポンスを送ることができる通信方式


### サーバーストリーミングの実装
- server streaming RPC : 1リクエスト複数レスポンス
- protoファイルにサーバーストリーミングのメソッド記述
```proto
service GreetingService {
	// サービスが持つメソッドの定義
	rpc Hello (HelloRequest) returns (HelloResponse);
	// サーバーストリーミングRPC
	rpc HelloServerStream (HelloRequest) returns (stream HelloResponse);
}
```
- コードを自動生成すると、`GreetingServiceServer`インターフェースにメソッドが追加されている
```go
type GreetingServiceServer interface {
	// サービスが持つメソッドの定義
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	// サーバーストリーミングRPC
	HelloServerStream(*HelloRequest, GreetingService_HelloServerStreamServer) error
	mustEmbedUnimplementedGreetingServiceServer()
}
```
- 返り値のHelloResponseが消えて第二引数にインターフェースが加わっていることがわかる
```go
// 自動生成された、サーバーストリーミングのためのインターフェース(for サーバー)
type GreetingService_HelloServerStreamServer interface {
	Send(*HelloResponse) error
	grpc.ServerStream
}
```
- ビジネスロジックを実装すると以下
```go
func (s *myServer) HelloServerStream(req *hellopb.HelloRequest, stream hellopb.GreetingService_HelloServerStreamServer) error {
	resCount := 5
	for i := 0; i < resCount; i++ {
		if err := stream.Send(&hellopb.HelloResponse{
			Message: fmt.Sprintf("[%d] Hello, %s!", i, req.GetName()),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
	}
	return nil // ストリームの終わり
}
```
- このようにUnary RPCは直接レスポンスを関数の返り値として返していたが、Server Streaming RPCでは第二引数のインターフェースのSendメソッドを用いてレスポンスを渡している
  - Sendメソッドの引数にHelloResponseを渡している
- また、ストリームの終端はHelloServerStreamメソッドをreturn文で終わらせることで実現できる
- 実際に`grpcurl -plaintext -d '{"name": "hsaki"}' localhost:8080 myapp.GreetingService.HelloServerStream`を叩くと以下の様にServer Streamingを実現できていることがわかる(1リクエスト5レスポンス)
```json
{
  "message": "[0] Hello, hsaki!"
}
{
  "message": "[1] Hello, hsaki!"
}
{
  "message": "[2] Hello, hsaki!"
}
{
  "message": "[3] Hello, hsaki!"
}
{
  "message": "[4] Hello, hsaki!"
}
```
- grpcurlを使わずにサーバーにリクエストを送信するクライアントコードを実装する方法は以下
- 自動生成されたコードは以下のようになっている
```go
type GreetingServiceClient interface {
	// サービスが持つメソッドの定義
	Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
	// サーバーストリーミングRPC
	HelloServerStream(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (GreetingService_HelloServerStreamClient, error)
}

type GreetingService_HelloServerStreamClient interface {
	Recv() (*HelloResponse, error)
	grpc.ClientStream
}
```
- メソッドの返り値がHelloResponseではなく、`GreetingService_HelloServerStreamClient`インターフェースになっていることがわかる
  - この`Recv()`メソッドがHelloResponseを返している
- そのため、Unary RPCとは異なりServer Stream RPCでは以下のように記述
```go
// Unary RPCがレスポンスを受け取るところ
func Hello() {
	// (一部抜粋)
	// Helloメソッドの実行 -> HelloResponse型のレスポンスresを入手
	res, err := client.Hello(context.Background(), req)
}

// Server Stream RPCがレスポンスを受け取るところ
func HelloServerStream() {
	// (一部抜粋)
	// サーバーから複数回レスポンスを受け取るためのストリームを得る
	stream, err := client.HelloServerStream(context.Background(), req)

	for {
		// ストリームからレスポンスを得る
		res, err := stream.Recv()
	}
}
```
- また、ストリームの終端は`Recv()`メソッドの返り値の2つ目のerrorがio.EOFの時である
## 自分用メモ
### HTTP/2とは
- 簡潔にHTTP/2の特徴は以下
  - ストリームという概念を導入したことでHTTP/1.1に比べて効率的に通信可能
  - ヘッダー圧縮を行うことでHTTP/1.1より通信量を減らすことが可能
  - サーバープッシュによりリクエストされる前にサーバーからリソースを送信することでHTTP/1.1よりラウンドトリップ回数を減らせる
    - ラウンドトリップとは通信において、ある場所からある場所へ何かを送信したときにそれが返ってくるまでの時間のこと
- HTTP/1.1では1回目のリクエスト->1回目のレスポンス->2回目のリクエスト->2回目のレスポンスというように一度リクエストを送信するとそのレスポンスを受け取ってから次のリクエストを行う必要がある
- HTTP/2ではストリームを導入することでリクエストとレスポンスを並列で行える
- 送受信するデータをフレームという単位に分割して扱う
  - フレームフォーマットはLenght,Type,Flags,ID,payloadから構成される
  - type,flagsが重要
  - typeには10種類ある
  - DATAフレームとHEADERSフレームの2種類がよく使われる
    - DATAフレーム : リクエスト/レスポンスのボディを送信するフレーム
    - HEADERSフレーム : リクエスト/レスポンスのヘッダーを送信するフレーム
  - HTTP/2では1つの送受信データを複数個のフレームに分割してやりとりするため、分割したデータの最後のフレームに最後であるという情報を加える必要がある
    - flagsにEND_STREAMフラグをつけることで実現している
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