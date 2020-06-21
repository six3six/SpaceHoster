package main

import (
	"context"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VmServer struct {
	protocol.UnimplementedVmServiceServer
}

func (*VmServer) Start(c context.Context, request *protocol.VmRequest) (*protocol.StatusVmResponse, error) {
	_, err := CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.StatusVmResponse{Code: protocol.StatusVmResponse_BAD_TOKEN, Status: protocol.StatusVmResponse_ABORTED}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	vm, err := GetVirtualMachine(int(request.Id))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	err = vm.Start()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	return &protocol.StatusVmResponse{Status: vm.StatusCode, Code: protocol.StatusVmResponse_OK, Message: vm.Error}, nil
}

func (s *VmServer) Status(c context.Context, request *protocol.VmRequest) (*protocol.StatusVmResponse, error) {
	_, err := CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.StatusVmResponse{Code: protocol.StatusVmResponse_BAD_TOKEN, Status: protocol.StatusVmResponse_ABORTED}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	vm, err := GetVirtualMachine(int(request.Id))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	if vm.StatusCode == protocol.StatusVmResponse_STOPPED || vm.StatusCode == protocol.StatusVmResponse_RUNNING {
		vmRef := proxmox.NewVmRef(vm.Id)
		err := proxmoxClient.CheckVmRef(vmRef)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
		state, err := proxmoxClient.GetVmState(vmRef)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}

		if state["status"] == "running" {
			vm.StatusCode = protocol.StatusVmResponse_RUNNING
		} else {
			vm.StatusCode = protocol.StatusVmResponse_STOPPED
		}

		err = vm.Sync()
		if err != nil {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	return &protocol.StatusVmResponse{Status: vm.StatusCode, Code: protocol.StatusVmResponse_OK, Message: vm.Error}, nil
}

func (s *VmServer) Create(c context.Context, request *protocol.CreateVmRequest) (*protocol.CreateVmResponse, error) {
	login, err := CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_BAD_TOKEN, Name: "", Id: 0}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}
	if request.Spec == nil {
		return nil, status.Errorf(codes.Aborted, "Specification does not exist")
	}
	spec := Specification{int(request.Spec.Core), int(request.Spec.Memory), int(request.Spec.Storage)}

	err = spec.CheckSpec()
	if err != nil {
		return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_NOT_ENOUGH_RESOURCES, Name: err.Error(), Id: 0}, nil
	}

	vmId, err := proxmoxClient.NextId()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	virtualMachines := database.Collection("virtualMachines")

	_, _ = virtualMachines.DeleteOne(c, bson.M{"id": vmId})

	vm := VirtualMachine{request.Name, vmId, protocol.StatusVmResponse_PREPARED, spec, "", login.Login, true, []Login{}}

	_, err = virtualMachines.InsertOne(c, vm)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	go VmCreationProcess(vm)

	return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_OK, Name: request.Name, Id: int32(vmId)}, nil
}
