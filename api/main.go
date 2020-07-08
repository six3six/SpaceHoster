package main

import (
	"github.com/robfig/cron/v3"
	"github.com/six3six/SpaceHoster/api/login"
	"github.com/six3six/SpaceHoster/api/protocol"
	"github.com/six3six/SpaceHoster/api/virtual_machine"
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
	loginService := login.Service{Database: database}
	_, err = c.AddFunc("@every 30m", loginService.CleanTokens)
	if err != nil {
		log.Fatalf("failed to add cron: %v", err)
	}

	c.Start()
	loginService.CleanTokens()
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Server listening at : %s", address)
	s := grpc.NewServer()
	protocol.RegisterLoginServiceServer(s, &login.Server{Database: database})
	protocol.RegisterVmServiceServer(s, &virtual_machine.VmServer{Database: database, LoginService: &loginService, Proxmox: proxmoxClient})
	log.Printf("Server available at : grpc://%s", address)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
