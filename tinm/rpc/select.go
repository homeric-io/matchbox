package rpc

import (
	"golang.org/x/net/context"

	"github.com/homeric-io/tinm/tinm/rpc/rpcpb"
	"github.com/homeric-io/tinm/tinm/server"
	pb "github.com/homeric-io/tinm/tinm/server/serverpb"
)

// selectServer wraps a tinm Server to be suitable for gRPC registration.
type selectServer struct {
	srv server.Server
}

func newSelectServer(s server.Server) rpcpb.SelectServer {
	return &selectServer{
		srv: s,
	}
}

func (s *selectServer) SelectGroup(ctx context.Context, req *pb.SelectGroupRequest) (*pb.SelectGroupResponse, error) {
	group, err := s.srv.SelectGroup(ctx, req)
	return &pb.SelectGroupResponse{Group: group}, grpcError(err)
}

func (s *selectServer) SelectProfile(ctx context.Context, req *pb.SelectProfileRequest) (*pb.SelectProfileResponse, error) {
	profile, err := s.srv.SelectProfile(ctx, req)
	return &pb.SelectProfileResponse{Profile: profile}, grpcError(err)
}
