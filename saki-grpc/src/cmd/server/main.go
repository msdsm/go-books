package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	hellopb "mygrpc/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}

func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	// メタデータ参照
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Println(md)
	}

	// メタデータ送信
	headerMD := metadata.New(map[string]string{"type": "unary", "from": "server", "in": "header"})
	if err := grpc.SetHeader(ctx, headerMD); err != nil {
		return nil, err
	}
	trailerMD := metadata.New(map[string]string{"type": "unary", "from": "server", "in": "trailer"})
	if err := grpc.SetTrailer(ctx, trailerMD); err != nil {
		return nil, err
	}

	// リクエストからnameを取り出す
	// "Hello, [名前]!"というレスポンスを返す
	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil

	/*
		// error返す
		err := status.Error(codes.Unknown, "unknown error occurred")
	*/

	/*
		// gRPCステータスのdetailsフィールドの利用
		stat := status.New(codes.Unknown, "unknown error occurred")
		stat, _ = stat.WithDetails(&errdetails.DebugInfo{
			Detail: "detail reason of err",
		})
		err := stat.Err()
		return nil, err
	*/
}

// HelloServerStreamメソッドのビジネスロジック実装
func (s *myServer) HelloServerStream(req *hellopb.HelloRequest, stream hellopb.GreetingService_HelloServerStreamServer) error {
	// メタデータ参照
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Println(md)
	}

	resCount := 5
	for i := 0; i < resCount; i++ {
		if err := stream.Send(&hellopb.HelloResponse{
			Message: fmt.Sprintf("[%d] Hello, %s!", i, req.GetName()),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
	}

	// ストリーム終了
	return nil
}

// HelloClientStreamメソッドのビジネスロジック実装
func (s *myServer) HelloClientStream(stream hellopb.GreetingService_HelloClientStreamServer) error {
	// メタデータ参照
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Println(md)
	}

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

func (s *myServer) HelloBiStreams(stream hellopb.GreetingService_HelloBiStreamsServer) error {
	// メタデータ参照
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Println(md)
	}

	// メタデータ送信
	headerMD := metadata.New(map[string]string{"type": "stream", "from": "server", "in": "header"})
	/*
		// すぐに送信
		if err := stream.SendHeader(headerMD); err != nil {
			return err
		}
	*/
	// 本来ヘッダーを送るタイミングで送る場合
	if err := stream.SetHeader(headerMD); err != nil {
		return err
	}

	trailerMD := metadata.New(map[string]string{"type": "stream", "from": "server", "in": "trailer"})
	stream.SetTrailer(trailerMD)

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) { // リクエストのストリーム終端
			return nil // return nilでレスポンスのストリーム終端とする
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

func NewMyServer() *myServer {
	return &myServer{}
}

func main() {
	// 1. 8080番portのListenerを作成
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// 2. gRPCサーバーを作成
	s := grpc.NewServer(
		/*
			grpc.ChainUnaryInterceptor(
				myUnaryServerInterceptor1,
				myUnaryServerInterceptor2,
			),
		*/
		grpc.ChainStreamInterceptor(
			myStreamServerInterceptor1,
			myStreamServerInterceptor2,
		),
		//grpc.UnaryInterceptor(myUnaryServerInterceptor1),
		//grpc.StreamInterceptor(myStreamServerInterceptor1),
	)

	// 3. gRPCサーバーにGreetingServiceを登録
	hellopb.RegisterGreetingServiceServer(s, NewMyServer())

	// 4. サーバーリフレクションの設定
	reflection.Register(s)

	// 5. 作成したgRPCサーバーを8080番ポートで稼働
	go func() {
		log.Printf("start gRPC server port: %v", port)
		s.Serve(listener)
	}()

	// 6. Ctrl+Cが入力され得たらGraceful shutwodnされるようにする
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}
