package main

import (
	"github.com/robfig/cron/v3"
	"github.com/six3six/SpaceHoster/api/protocol"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	var err error

	address := "localhost:8080"
	dockerClient = connectDocker()
	proxmoxClient = connectProxmox()
	connectDB()

	c := cron.New()
	CleanTokensCronId, err = c.AddFunc("@every 30m", CleanTokens)
	if err != nil {
		log.Fatalf("failed to add cron: %v", err)
	}

	c.Start()
	CleanTokens()
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Server listening at : %s", address)
	s := grpc.NewServer()
	protocol.RegisterLoginServiceServer(s, &loginServer{})
	protocol.RegisterVmServiceServer(s, &VmServer{})
	log.Printf("Server available at : grpc://%s", address)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
