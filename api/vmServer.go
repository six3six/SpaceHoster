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
	user, err := CheckToken(request.Token)
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
	if vm.Owner != user.Login {
		return &protocol.StatusVmResponse{Code: protocol.StatusVmResponse_NOT_ALLOWED, Status: protocol.StatusVmResponse_ABORTED}, nil
	}

	err = vm.Start()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	return &protocol.StatusVmResponse{Status: vm.StatusCode, Code: protocol.StatusVmResponse_OK, Message: vm.Error}, nil
}

func (*VmServer) Stop(c context.Context, request *protocol.VmRequest) (*protocol.StatusVmResponse, error) {
	user, err := CheckToken(request.Token)
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
	if vm.Owner != user.Login {
		return &protocol.StatusVmResponse{Code: protocol.StatusVmResponse_NOT_ALLOWED, Status: protocol.StatusVmResponse_ABORTED}, nil
	}

	err = vm.Stop()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	return &protocol.StatusVmResponse{Status: vm.StatusCode, Code: protocol.StatusVmResponse_OK, Message: vm.Error}, nil
}

func (s *VmServer) Status(c context.Context, request *protocol.VmRequest) (*protocol.StatusVmResponse, error) {
	user, err := CheckToken(request.Token)
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

	if vm.Owner != user.Login {
		return &protocol.StatusVmResponse{Code: protocol.StatusVmResponse_NOT_ALLOWED, Status: protocol.StatusVmResponse_ABORTED}, nil
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
	user, err := CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_BAD_TOKEN, Name: "", Id: 0}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}
	if request.Specification == nil {
		return nil, status.Errorf(codes.Aborted, "Specification does not exist")
	}
	spec := Specification{int(request.Specification.Core), int(request.Specification.Memory), int(request.Specification.Storage)}

	err = spec.CheckSpec()
	if err != nil {
		return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_NOT_ENOUGH_RESOURCES, Name: err.Error(), Id: 0}, nil
	}

	err = spec.CheckFreeResources(*user)
	if err != nil {
		return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_NOT_ENOUGH_RESOURCES, Name: err.Error(), Id: 0}, nil
	}

	vmId, err := proxmoxClient.NextId()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	virtualMachines := database.Collection("virtualMachines")

	_, _ = virtualMachines.DeleteOne(c, bson.M{"id": vmId})

	vm := VirtualMachine{request.Name, vmId, protocol.StatusVmResponse_PREPARED, "", user.Login, true, []Login{}}

	_, err = virtualMachines.InsertOne(c, vm)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	go VmCreationProcess(vm, string(user.Login), user.EncodedPassword)

	return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_OK, Name: request.Name, Id: int32(vmId)}, nil
}

func (s *VmServer) List(c context.Context, request *protocol.JustTokenRequest) (*protocol.ListVmResponse, error) {
	user, err := CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.ListVmResponse{Code: protocol.ListVmResponse_BAD_TOKEN, Id: []int32{}}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}
	virtualMachines := database.Collection("virtualMachines")
	vms, err := virtualMachines.Find(c, bson.M{"owner": user.Login})
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	var id []int32
	for vms.Next(c) {

		var vm VirtualMachine
		err := vms.Decode(&vm)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}

		if vm.Created() {
			id = append(id, int32(vm.Id))
		}
	}

	return &protocol.ListVmResponse{Code: protocol.ListVmResponse_OK, Id: id}, nil
}

func (s *VmServer) FreeResources(c context.Context, request *protocol.JustTokenRequest) (*protocol.FreeResourcesResponse, error) {
	user, err := CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.FreeResourcesResponse{Code: protocol.FreeResourcesResponse_BAD_TOKEN, Free: nil, Total: nil}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	free, err := user.GetFreeResources()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	total := user.Quota

	freePrt := protocol.VmSpecification{Storage: int32(free.Storage), Memory: int32(free.Memory), Core: int32(free.Cores)}
	totalPrt := protocol.VmSpecification{Storage: int32(total.Storage), Memory: int32(total.Memory), Core: int32(total.Cores)}

	return &protocol.FreeResourcesResponse{Code: protocol.FreeResourcesResponse_OK, Free: &freePrt, Total: &totalPrt}, nil
}
