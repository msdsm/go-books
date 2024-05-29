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
```proto
syntax = "proto3";
```
- packageはGoと同じ扱い(他のprotoファイルで定義された型をパッケージ名.型名で参照可能)
```proto
package myapp;
```
- gRPCで呼び出そうとするProcedure(関数)をメソッド、そしてそのメソッドをいくつかまとめてひとくくりにしたものをサービスという
  - 以下の例では2つのことを行っている
    - 引数にHelloRequest型、戻り値にHelloResponse型を持つメソッドHelloを定義
    - Helloメソッド一つを持つGreetingServiceサービスを定義
```proto
// サービスの定義
service GreetingService {
	// サービスが持つメソッドの定義
	rpc Hello (HelloRequest) returns (HelloResponse); 
}
```
- 上記のHelloRequest, HelloResponse型を以下のように定義
```proto
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
## コード自動生成
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


## gRPCサーバー実装
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

## gRPCクライアント実装
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

## ストリーミング処理
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


## サーバーストリーミングの実装
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

## クライアントストリーミングの実装
- Client Streaming RPC : 複数リクエスト1レスポンス
- protoファイルにメソッド追加
```proto
service GreetingService {
	// サービスが持つメソッドの定義
	rpc Hello (HelloRequest) returns (HelloResponse);
	// サーバーストリーミングRPC
	rpc HelloServerStream (HelloRequest) returns (stream HelloResponse);
	// クライアントストリーミングRPC
	rpc HelloClientStream (stream HelloRequest) returns (HelloResponse);
}
```
- コードを自動生成するとサーバーサイドは以下のように生成される
```go
type GreetingServiceServer interface {
	// サービスが持つメソッドの定義
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	// サーバーストリーミングRPC
	HelloServerStream(*HelloRequest, GreetingService_HelloServerStreamServer) error
	// クライアントストリーミングRPC
	HelloClientStream(GreetingService_HelloClientStreamServer) error
	mustEmbedUnimplementedGreetingServiceServer()
}

type GreetingService_HelloClientStreamServer interface {
	SendAndClose(*HelloResponse) error
	Recv() (*HelloRequest, error)
	grpc.ServerStream
}
```
- Client Streaming RPCをUnary RPCと比較すると、引数がHelloRequest型ではなく、`GreetingService_HelloClientStreamServer`インターフェースになっていて、戻り値からHelloResponseが消えている
- `GreetingService_HelloClientStreamServer`は`SendAndClose()`メソッドと`Recv()`メソッドを持つ
- ビジネスロジックを記述すると以下
```go
func (s *myServer) HelloClientStream(stream hellopb.GreetingService_HelloClientStreamServer) error {
	nameList := make([]string, 0)
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			message := fmt.Sprintf("Hello, %v!", nameList)
			return stream.SendAndClose(&hellopb.HelloResponse{
				Message: message,
			})
		}
		if err != nil {
			return err
		}
		nameList = append(nameList, req.GetName())
	}
}
```
- このように`Recv()`を呼び出してリクエストを取得する
- また、その返り値のerrorがio.EOFになっているときにストリーム終端
- レスポンスは`SendAndClose()`の引数にHelloResponseを渡して返す
- 動作確認:`grpcurl -plaintext -d '{"name": "hsaki"}{"name": "a-san"}{"name": "b-san"}{"name": "c-san"}{"name": "d-san"}' localhost:8080 myapp.GreetingService.HelloClientStream`
```json
{
  "message": "Hello, [hsaki a-san b-san c-san d-san]!"
}
```
- クライアントは以下のように生成されている
```go
type GreetingServiceClient interface {
	// サービスが持つメソッドの定義
	Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
	// サーバーストリーミングRPC
	HelloServerStream(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (GreetingService_HelloServerStreamClient, error)
	// クライアントストリーミングRPC
	HelloClientStream(ctx context.Context, opts ...grpc.CallOption) (GreetingService_HelloClientStreamClient, error)
}
type GreetingService_HelloClientStreamClient interface {
	Send(*HelloRequest) error
	CloseAndRecv() (*HelloResponse, error)
	grpc.ClientStream
}
```
- この`Send()`と`CloseAndRecv()`を使って以下のようにかける
```go
// Unary RPCがリクエストを送るところ
func Hello() {
	// (一部抜粋)
	// Helloメソッドの実行
	res, err := client.Hello(context.Background(), req)
}

// Client Stream RPCがリクエストを送るところ
func HelloClientStream() {
	// (一部抜粋)
	// サーバーに複数回リクエストを送るためのストリームを得る
	stream, err := client.HelloClientStream(context.Background())

	for i := 0; i < sendCount; i++ {
		// ストリームを通じてリクエストを送信
		stream.Send(&hellopb.HelloRequest{
			Name: name,
		})
	}
  // ストリームからレスポンスを得る
	res, err := stream.CloseAndRecv()
}
```

## 双方向ストリーミングの実装
- Bidirectional streaming RPC : 複数リクエスト複数レスポンス
- protoファイル定義
```proto
service GreetingService {
	// サービスが持つメソッドの定義
	rpc Hello (HelloRequest) returns (HelloResponse);
	// サーバーストリーミングRPC
	rpc HelloServerStream (HelloRequest) returns (stream HelloResponse);
	// クライアントストリーミングRPC
	rpc HelloClientStream (stream HelloRequest) returns (HelloResponse);
	// 双方向ストリーミングRPC
	rpc HelloBiStreams (stream HelloRequest) returns (stream HelloResponse);
}
```
- サーバーサイドで以下のコードが生成される
```go
type GreetingServiceServer interface {
	// サービスが持つメソッドの定義
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	// サーバーストリーミングRPC
	HelloServerStream(*HelloRequest, GreetingService_HelloServerStreamServer) error
	// クライアントストリーミングRPC
	HelloClientStream(GreetingService_HelloClientStreamServer) error
	// 双方向ストリーミングRPC
	HelloBiStreams(GreetingService_HelloBiStreamsServer) error
	mustEmbedUnimplementedGreetingServiceServer()
}
type GreetingService_HelloBiStreamsServer interface {
	Send(*HelloResponse) error
	Recv() (*HelloRequest, error)
	grpc.ServerStream
}
```
- ビジネスロジックの実装は、`Send()`メソッドと`Recv()`メソッドを用いてServer Streaming, Client Streamingと同じ様にかける
```go
func (s *myServer) HelloBiStreams(stream hellopb.GreetingService_HelloBiStreamsServer) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		message := fmt.Sprintf("Hello, %v!", req.GetName())
		if err := stream.Send(&hellopb.HelloResponse{
			Message: message,
		}); err != nil {
			return err
		}
	}
}
```
- clientは以下の様に生成されている
```go
type GreetingServiceClient interface {
	// サービスが持つメソッドの定義
	Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
	// サーバーストリーミングRPC
	HelloServerStream(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (GreetingService_HelloServerStreamClient, error)
	// クライアントストリーミングRPC
	HelloClientStream(ctx context.Context, opts ...grpc.CallOption) (GreetingService_HelloClientStreamClient, error)
	// 双方向ストリーミングRPC
	HelloBiStreams(ctx context.Context, opts ...grpc.CallOption) (GreetingService_HelloBiStreamsClient, error)
}
type GreetingService_HelloBiStreamsClient interface {
	Send(*HelloRequest) error
	Recv() (*HelloResponse, error)
	grpc.ClientStream
}
```
- clientも同じ様に`Send()`, `Recv()`メソッドを用いて実装するが、送信の終端に達した時は`CloseSend()`メソッドを使用する
  - `CloseSend()`は`grpc.ClientStream`インターフェース由来のメソッド
```go
func HelloBiStreams() {
	stream, err := client.HelloBiStreams(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	sendNum := 5
	fmt.Printf("Please enter %d names.\n", sendNum)

	var sendEnd, recvEnd bool
	sendCount := 0
	for !(sendEnd && recvEnd) {
		// 送信処理
		if !sendEnd {
			scanner.Scan()
			name := scanner.Text()

			sendCount++
			if err := stream.Send(&hellopb.HelloRequest{
				Name: name,
			}); err != nil {
				fmt.Println(err)
				sendEnd = true
			}

			if sendCount == sendNum {
				sendEnd = true
				if err := stream.CloseSend(); err != nil {
					fmt.Println(err)
				}
			}
		}

		// 受信処理
		if !recvEnd {
			if res, err := stream.Recv(); err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Println(err)
				}
				recvEnd = true
			} else {
				fmt.Println(res.GetMessage())
			}
		}
	}
}
```

## gRPCのステータスコード
- gRPCではメソッドの呼び出しに成功した場合には中で何が起きてもHTTPレスポンスステータスコードは200を返す
- gRPCでは独自のエラ〜コードが17個ある
  - 0  : OK               	: 正常
  - 1  : Canceled         	: 処理がキャンセルされた
  - 2  : Unknown          	: 不明なエラー
  - 3  : InvalidArgument  	: 無効な引数でメソッドを呼び出した
  - 4  : DeadlineExceeded 	: タイムアウト
  - 5  : NotFound         	: 要求されたエンティティが存在しなかった
  - 6  : AlreadyExists	  	: 既に存在しているエンティティを作成するようなリクエストだったため失敗
  - 7  : PermissionDenied 	: そのメソッドを実行するための権限がない
  - 8  : ResourceExhausted	: リクエストを処理するためのquotaが枯渇した
  - 9  : FailedPrecondition : 処理を実行できる状態ではないためリクエストが拒否された (例: 中身があるディレクトリをrmdirしようとした)
  - 10 : Aborted	        : トランザクションがコンフリクトしたなどして、処理が異常終了させられた
  - 11 : OutOfRange	        : 有効範囲外の操作をリクエストされた (例: ファイルサイズを超えたオフセットからのreadを指示された)
  - 12 : Unimplemented	    : サーバーに実装されていないサービス・メソッドを呼び出そうとした
  - 13 : Internal	        : サーバー内で重大なエラーが発生した
  - 14 : Unavailable	    : メソッドを実行するための用意ができていない
  - 15 : DataLoss	        : NWの問題で伝送中にパケットが失われた
  - 16 : Unauthenticated	: ユーザー認証に失敗した
  - https://grpc.io/docs/guides/error/#error-status-codes
- `google.golang.org/grpc/status`パッケージを用意して以下の様にerrorの作成ができる
```go
err := status.Error(codes.Unknown, "unknown error occurred")
```
- 動作確認として`grpcurl -plaintext -d '{"name": "hsaki"}' localhost:8080 myapp.GreetingService.Hello`をたたくと以下のように返ってくる
```
ERROR:
  Code: Unknown
  Message: unknown error occurred
```
- `status.Error()`はコードと文字列からerrorを作成するがその逆変換が`status.FromError()`
```go
func FromError(err error) (s *Status, ok bool)
```
- この返り値の`Code()`, `Message()`メソッドを使って取り出せる
```go
if stat, ok := status.FromError(err); ok {
	fmt.Printf("code: %s\n", stat.Code())
	fmt.Printf("message: %s\n", stat.Message())
}
```
- gRPCで返すerrorではcode,messageに加えてdetailsフィールドもあり、以下のようにかける
	1. `status.New()`でステータス型の作成
	2. `WithDetails()`でステータス型に詳細情報を付加
	3. ステータス型の`Err()`でエラー生成
```go
stat := status.New(codes.Unknown, "unknown error occurred")
stat, _ = stat.WithDetails(&errdetails.DebugInfo{
	Detail: "detail reason of err",
})
err := stat.Err()
```
- grpcurlを叩くと以下の様になる
```
ERROR:
  Code: Unknown
  Message: unknown error occurred
  Details:
  1)    {
          "@type": "type.googleapis.com/google.rpc.DebugInfo",
          "detail": "detail reason of err"
        }
```
- またクライアントの方では上述の`FromError`の返り値のステータス型から`Details()`メソッドでdetailを取り出せる
  - Server側でDetailの作成はprotoファイルを元にしている(errdetails.DebugInfoはprotoファイルから自動生成されたコードのパッケージ)なのでdetailのメッセージ型をでシリアライズするためにclient側で`"google.golang.org/genproto/googleapis/rpc/errdetails"`をimportする必要がある


## インターセプタ(サーバーサイド)
- gRPCでは、ハンドラ処理の前後に追加処理を挟むミドルウェアのことをインターセプタと呼ぶ
- Unary RPCとStreaming RPCでInterceptorの引数の型が異なる
### Unary RPCのInterceptor
- 以下のように前処理と後処理を記述できる
```go
func myUnaryServerInterceptor1(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("[pre] my unary server interceptor 1: ", info.FullMethod) // ハンドラの前に割り込ませる前処理
	res, err := handler(ctx, req) // 本来の処理
	log.Println("[post] my unary server interceptor 1: ", m) // ハンドラの後に割り込ませる後処理
	return res, err
}
```
- 以下のようにサーバーに導入できる
```go
s := grpc.NewServer(
	grpc.UnaryInterceptor(myUnaryServerInterceptor1),
)
```

### Stream RPCのInterceptor
```go
func myStreamServerInterceptor1(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// ストリームがopenされたときに行われる前処理
	log.Println("[pre stream] my stream server interceptor 1: ", info.FullMethod)

	err := handler(srv, &myServerStreamWrapper1{ss}) // 本来のストリーム処理

	// ストリームがcloseされるときに行われる後処理
	log.Println("[post stream] my stream server interceptor 1: ")
	return err
}

type myServerStreamWrapper1 struct {
	grpc.ServerStream
}

func (s *myServerStreamWrapper1) RecvMsg(m interface{}) error {
	// ストリームから、リクエストを受信
	err := s.ServerStream.RecvMsg(m)
	// 受信したリクエストを、ハンドラで処理する前に差し込む前処理
	if !errors.Is(err, io.EOF) {
		log.Println("[pre message] my stream server interceptor 1: ", m)
	}
	return err
}

func (s *myServerStreamWrapper1) SendMsg(m interface{}) error {
	// ハンドラで作成したレスポンスを、ストリームから返信する直前に差し込む後処理
	log.Println("[post message] my stream server interceptor 1: ", m)
	return s.ServerStream.SendMsg(m)
}
```
- ストリーミングRPCでは以下のような処理になる
  - ストリームをopenする
  - 以下を繰り返す
    - ストリームからリクエストを受信
    - ハンドラ内で、リクエストに対するレスポンスを生成
    - ストリームを通じてレスポンスを送信
  - ストリームをclose
- そのため、前処理・後処理といってもストリームopen/closeのときの処理なのか、ストリームから実際にデータを送受信するときの処理なのかという選択肢が生まれる
- ストリームopen/closeに着目した前処理・後処理はUnary RPCと同様にhandlerの前後に書く
```go
func myStreamServerInterceptor1(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 前処理をここに書く

	err := handler(srv, &myServerStreamWrapper1{ss}) // 本来のストリーム処理

	// 後処理をここに書く

	return err
}
```
- 送受信に着目した前処理・後処理は、`grpc.ServerStream`インターフェース型の`RecvMsg`, `SendMsg`メソッドで行われる
- そのため、リクエスト受信時・レスポンス送信時に自分のやりたい処理を入れ込むには以下をやる
  - `grpc.ServerStream`インターフェース型を満たす独自構造体を作成
  - 独自構造体の`RecvMsg`,`SendMsg`メソッドを自分がやりたい処理を入れ込む形でオーバーライド
- Streaming RPCのインターセプタは以下の様に導入できる
```go
s := grpc.NewServer(
	grpc.StreamInterceptor(myStreamServerInterceptor1),
)
```
### 複数のInterceptor
複数個のインターセプタを導入することもできる
```go
// Unary RPC
s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			myUnaryServerInterceptor1,
			myUnaryServerInterceptor2,
		),
	)
```
```go
// Streaming RPC
s := grpc.NewServer(
	grpc.ChainStreamInterceptor(
		myStreamServerInterceptor1,
		myStreamServerInterceptor2,
	),
)
```
- このときインターセプタの実行順序は以下のようになる(前処理は記述順、後処理は逆順)
	1. インターセプタ1の前処理
	2. インターセプタ2の前処理
	3. ハンドラによる本処理
	4. インターセプタ2の後処理
	5. インターセプタ1の後処理

## インターセプタ(クライアントサイド)
- サーバーと同じことをクライアントでもできる
- クライアントがリクエストを送信する前、レスポンスを受信する前に処理を挟める
### Unary RPCのInterceptor
- grpcパッケージで以下のような形であるべきと定められている
```go
type UnaryClientInterceptor func(ctx context.Context, method string, req, reply interface{}, cc *ClientConn, invoker UnaryInvoker, opts ...CallOption) error
```
- Unary RPCのInterceptorは以下のように実装できる
```go
func myUnaryClientInteceptor1(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	fmt.Println("[pre] my unary client interceptor 1", method, req) // リクエスト送信前に割り込ませる前処理
	err := invoker(ctx, method, req, res, cc, opts...) // 本来のリクエスト
	fmt.Println("[post] my unary client interceptor 1", res) // リクエスト送信後に割り込ませる後処理
	return err
}
```
- このInterceptorは以下のように導入できる
```go
conn, err := grpc.Dial(
		address,
		grpc.WithUnaryInterceptor(myUnaryClientInteceptor1), // ここ

		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
```
### Stream RPCのInterceptor
- 以下のように定められている
```go
type StreamClientInterceptor func(ctx context.Context, desc *StreamDesc, cc *ClientConn, method string, streamer Streamer, opts ...CallOption) (ClientStream, error)
```
- これを用いて以下のように実装できる
```go
func myStreamClientInterceptor1(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	// ストリームがopenされる前に行われる前処理
	log.Println("[pre] my stream client interceptor 1", method)

	stream, err := streamer(ctx, desc, cc, method, opts...)
	return &myClientStreamWrapper1{stream}, err
}

type myClientStreamWrapper1 struct {
	grpc.ClientStream
}

func (s *myClientStreamWrapper1) SendMsg(m interface{}) error {
	// リクエスト送信前に割り込ませる前処理
	log.Println("[pre message] my stream client interceptor 1: ", m)

	// リクエスト送信
	return s.ClientStream.SendMsg(m)
}

func (s *myClientStreamWrapper1) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m) // レスポンス受信処理

	// レスポンス受信後に割り込ませる後処理
	if !errors.Is(err, io.EOF) {
		log.Println("[post message] my stream client interceptor 1: ", m)
	}
	return err
}

func (s *myClientStreamWrapper1) CloseSend() error {
	err := s.ClientStream.CloseSend() // ストリームをclose

	// ストリームがcloseされた後に行われる後処理
	log.Println("[post] my stream client interceptor 1")
	return err
}
```
- まず、ストリームopen前の処理はInterceptor関数の中にかく
- クライアントInterceptorは返り値として`grpc.ClientStream`を返し、この返り値で得られるClientStreamは以下のように定義されている
```go
type ClientStream interface {
	// (一部抜粋)
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
	CloseSend() error
}
```
- 上から順に、リクエスト送信、レスポンス受信、ストリームclose処理である
- そのため、これらの前後に処理を割り込ませるために、独自のクライアントストリム構造体(Wrapper)を作ってメソッドをオーバーライドさせる
- Stream Interceptorは以下のように導入できる
```go
func main() {
	conn, err := grpc.Dial(
		address,
		grpc.WithStreamInterceptor(myStreamClientInteceptor1), // ここ

		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}
```
### 複数のInterceptor
- Unary RPCの場合
```go
conn, err := grpc.Dial(
	address,
	// これ
	grpc.WithChainUnaryInterceptor(
		myUnaryClientInteceptor1,
		myUnaryClientInteceptor2,
	),
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithBlock(),
)
```
- Stream RPCの場合
```go
conn, err := grpc.Dial(
	address,
	// これ
	grpc.WithChainStreamInterceptor(
		myStreamClientInteceptor1,
		myStreamClientInteceptor2,
	),
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithBlock(),
)
```
- また、処理の順序についてはサーバーサイドと同じく以下のようになる
	1. インターセプタ1の前処理
	2. インターセプタ2の前処理
	3. ハンドラによる本処理
	4. インターセプタ2の後処理
	5. インターセプタ1の後処理


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