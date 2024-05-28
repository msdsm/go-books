package main

import (
	"errors"
	"io"
	"log"

	"google.golang.org/grpc"
)

func myStreamServerInterceptor1(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// ストリームがopenされたときの前処理
	log.Println("[pre stream] mystream server interceptor 1: ", info.FullMethod)

	// ストリーム処理
	err := handler(srv, &myServerStreamWrapper1{ss})

	// ストリームがcloseされるときの後処理
	log.Println("[post stream] my stream server interceptor 1: ")
	return err
}

type myServerStreamWrapper1 struct {
	grpc.ServerStream
}

func (s *myServerStreamWrapper1) RecvMsg(m interface{}) error {
	// ストリームからリクエスト受信
	err := s.ServerStream.RecvMsg(m)

	// 受信したリクエストをハンドラで処理する前に差し込む前処理
	if !errors.Is(err, io.EOF) {
		log.Println("[pre message] my stream server interceptor 1: ", m)
	}
	return err
}

func (s *myServerStreamWrapper1) SendMsg(m interface{}) error {
	log.Println("[post message] my stream server interceptor 1: ", m)
	return s.ServerStream.SendMsg(m)
}

func myStreamServerInterceptor2(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// ストリームがopenされたときの前処理
	log.Println("[pre stream] mystream server interceptor 2: ", info.FullMethod)

	// ストリーム処理
	err := handler(srv, &myServerStreamWrapper2{ss})

	// ストリームがcloseされるときの後処理
	log.Println("[post stream] my stream server interceptor 2: ")
	return err
}

type myServerStreamWrapper2 struct {
	grpc.ServerStream
}

func (s *myServerStreamWrapper2) RecvMsg(m interface{}) error {
	// ストリームからリクエスト受信
	err := s.ServerStream.RecvMsg(m)

	// 受信したリクエストをハンドラで処理する前に差し込む前処理
	if !errors.Is(err, io.EOF) {
		log.Println("[pre message] my stream server interceptor 2: ", m)
	}
	return err
}

func (s *myServerStreamWrapper2) SendMsg(m interface{}) error {
	log.Println("[post message] my stream server interceptor 2: ", m)
	return s.ServerStream.SendMsg(m)
}
