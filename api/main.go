package main

import (
	"github.com/six3six/SpaceHoster/api/protocol"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {

	dockerClient = connectDocker()
	proxmoxClient = connectProxmox()
	connectDB()

	lis, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	protocol.RegisterLoginServiceServer(s, &loginServer{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
