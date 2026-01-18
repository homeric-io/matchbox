package rpc

import (
	"golang.org/x/net/context"

	"github.com/homeric-io/tinm/tinm/rpc/rpcpb"
	"github.com/homeric-io/tinm/tinm/server"
	pb "github.com/homeric-io/tinm/tinm/server/serverpb"
)

// genericServer takes a tinm Server and implements a gRPC GenericServer.
type genericServer struct {
	srv server.Server
}

func newGenericServer(s server.Server) rpcpb.GenericServer {
	return &genericServer{
		srv: s,
	}
}

func (s *genericServer) GenericPut(ctx context.Context, req *pb.GenericPutRequest) (*pb.GenericPutResponse, error) {
	_, err := s.srv.GenericPut(ctx, req)
	return &pb.GenericPutResponse{}, grpcError(err)
}

func (s *genericServer) GenericGet(ctx context.Context, req *pb.GenericGetRequest) (*pb.GenericGetResponse, error) {
	template, err := s.srv.GenericGet(ctx, req)
	return &pb.GenericGetResponse{Config: []byte(template)}, grpcError(err)
}

func (s *genericServer) GenericDelete(ctx context.Context, req *pb.GenericDeleteRequest) (*pb.GenericDeleteResponse, error) {
	err := s.srv.GenericDelete(ctx, req)
	return &pb.GenericDeleteResponse{}, grpcError(err)
}
