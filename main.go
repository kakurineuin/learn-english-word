package main

import (
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/kakurineuin/learn-english-word/pb"
	"github.com/kakurineuin/learn-english-word/pkg/endpoint"
	"github.com/kakurineuin/learn-english-word/pkg/service"
	"github.com/kakurineuin/learn-english-word/pkg/transport"
)

const PORT = ":8090"

func main() {
	logger := log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		level.Error(logger).Log("msg", "net listen fail", "err", err)
		os.Exit(1)
	}

	wordService := service.WordService{}
	wordEndpoints := endpoint.MakeEndpoints(wordService, logger)
	myGrpcServer := transport.NewGRPCServer(wordEndpoints, logger)

	grpcServer := grpc.NewServer()
	pb.RegisterWordServiceServer(grpcServer, myGrpcServer)
	reflection.Register(grpcServer)
	level.Info(logger).Log("msg", "Starting gRPC server at "+PORT)
	err = grpcServer.Serve(listener)

	if err != nil {
		level.Error(logger).Log("msg", "grpcServer serve fail", "err", err)
	}
}
