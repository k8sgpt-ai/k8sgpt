package server

import (
	rpc "buf.build/gen/go/ronaldpetty/ronk8sgpt/grpc/go/schema/v1/schemav1grpc"
)

type handler struct {
	rpc.UnimplementedServerServiceServer
}
