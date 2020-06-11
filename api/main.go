package main

import (
	"github.com/six3six/SpaceHoster/api/protocol"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	address := "localhost:8080"
	dockerClient = connectDocker()
	proxmoxClient = connectProxmox()
	connectDB()

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Server listening at : %s", address)
	s := grpc.NewServer()
	protocol.RegisterLoginServiceServer(s, &loginServer{})
	log.Printf("Server available at : grpc://%s", address)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
