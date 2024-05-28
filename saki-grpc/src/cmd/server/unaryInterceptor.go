package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

func myUnaryServerInterceptor1(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// ハンドラの前処理
	log.Println("[pre] my unary server interceptor 1: ", info.FullMethod, req)

	// 本来の処理
	res, err := handler(ctx, req)

	// ハンドラの後処理
	log.Println("[post] my unary serer interceptor 1: ", res)
	return res, err
}

func myUnaryServerInterceptor2(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// ハンドラの前処理
	log.Println("[pre] my unary server interceptor 2: ", info.FullMethod, req)

	// 本来の処理
	res, err := handler(ctx, req)

	// ハンドラの後処理
	log.Println("[post] my unary serer interceptor 2: ", res)
	return res, err
}
