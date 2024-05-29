package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

func myUnaryClientInterceptor1(
	ctx context.Context,
	method string,
	req, res interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	// リクエスト送信前の前処理
	fmt.Println("[pre] my unary client interceptor 1", method, req)

	// リクエスト
	err := invoker(ctx, method, req, res, cc, opts...)

	// リクエスト送信後の後処理
	fmt.Println("[post] my unary client interceptor 1", res)

	return err
}

func myUnaryClientInterceptor2(
	ctx context.Context,
	method string,
	req, res interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	// リクエスト送信前の前処理
	fmt.Println("[pre] my unary client interceptor 2", method, req)

	// リクエスト
	err := invoker(ctx, method, req, res, cc, opts...)

	// リクエスト送信後の後処理
	fmt.Println("[post] my unary client interceptor 2", res)

	return err
}
