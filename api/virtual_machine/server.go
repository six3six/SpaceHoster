package virtual_machine

import (
	"context"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/six3six/SpaceHoster/api/login"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VmServer struct {
	protocol.UnimplementedVmServiceServer
	Database     *mongo.Database
	LoginService *login.Service
	Proxmox      *proxmox.Client
}

func (s *VmServer) Start(c context.Context, request *protocol.VmRequest) (*protocol.StatusVmResponse, error) {
	user, err := s.LoginService.CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.StatusVmResponse{Code: protocol.Code_BAD_TOKEN, Status: protocol.Status_ABORTED}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	vm, err := GetVirtualMachine(s.Database, int(request.Id))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	if vm.Owner != user.Login {
		return &protocol.StatusVmResponse{Code: protocol.Code_NOT_ENOUGH_RESOURCES, Status: protocol.Status_ABORTED}, nil
	}

	err = vm.Start()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	return &protocol.StatusVmResponse{Status: vm.StatusCode, Code: protocol.Code_OK, Message: vm.Error}, nil
}

func (s *VmServer) Stop(c context.Context, request *protocol.VmRequest) (*protocol.StatusVmResponse, error) {
	user, err := s.LoginService.CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.StatusVmResponse{Code: protocol.Code_BAD_TOKEN, Status: protocol.Status_ABORTED}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	vm, err := GetVirtualMachine(s.Database, int(request.Id))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	if vm.Owner != user.Login {
		return &protocol.StatusVmResponse{Code: protocol.Code_NOT_ALLOWED, Status: protocol.Status_ABORTED}, nil
	}

	err = vm.Stop()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	return &protocol.StatusVmResponse{Status: vm.StatusCode, Code: protocol.Code_OK, Message: vm.Error}, nil
}

func (s *VmServer) Status(c context.Context, request *protocol.VmRequest) (*protocol.StatusVmResponse, error) {
	user, err := s.LoginService.CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.StatusVmResponse{Code: protocol.Code_BAD_TOKEN, Status: protocol.Status_ABORTED}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	vm, err := GetVirtualMachine(s.Database, int(request.Id))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	if vm.Owner != user.Login {
		return &protocol.StatusVmResponse{Code: protocol.Code_NOT_ALLOWED, Status: protocol.Status_ABORTED}, nil
	}

	if vm.StatusCode == protocol.Status_STOPPED || vm.StatusCode == protocol.Status_RUNNING {
		vmRef := proxmox.NewVmRef(vm.Id)
		err := s.Proxmox.CheckVmRef(vmRef)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
		state, err := s.Proxmox.GetVmState(vmRef)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}

		if state["status"] == "running" {
			vm.StatusCode = protocol.Status_RUNNING
		} else {
			vm.StatusCode = protocol.Status_STOPPED
		}

		err = Sync(s.Database, vm)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	return &protocol.StatusVmResponse{Status: vm.StatusCode, Code: protocol.Code_OK, Message: vm.Error}, nil
}

func (s *VmServer) Create(c context.Context, request *protocol.CreateVmRequest) (*protocol.CreateVmResponse, error) {
	user, err := s.LoginService.CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.CreateVmResponse{Code: protocol.Code_BAD_TOKEN, Name: "", Id: 0}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}
	if request.Specification == nil {
		return nil, status.Errorf(codes.Aborted, "Specification does not exist")
	}
	spec := Specification{Cores: int(request.Specification.Core), Memory: int(request.Specification.Memory), Storage: int(request.Specification.Storage)}

	err = spec.CheckMinimumResources()
	if err != nil {
		return &protocol.CreateVmResponse{Code: protocol.Code_NOT_ENOUGH_RESOURCES, Name: err.Error(), Id: 0}, nil
	}

	err = spec.CheckFreeResources(*user)
	if err != nil {
		return &protocol.CreateVmResponse{Code: protocol.Code_NOT_ENOUGH_RESOURCES, Name: err.Error(), Id: 0}, nil
	}

	vmId, err := s.Proxmox.NextId()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	virtualMachines := s.Database.Collection("virtualMachines")

	_, _ = virtualMachines.DeleteOne(c, bson.M{"id": vmId})

	vm := VirtualMachine{request.Name, vmId, protocol.Status_PREPARED, "", user.Login, true, []login.Login{}, s.Proxmox}

	_, err = virtualMachines.InsertOne(c, vm)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	go VmCreationProcess(vm, s.Database, string(user.Login), user.EncodedPassword, spec)

	return &protocol.CreateVmResponse{Code: protocol.Code_OK, Name: request.Name, Id: int32(vmId)}, nil
}

func (s *VmServer) List(c context.Context, request *protocol.JustTokenRequest) (*protocol.ListVmResponse, error) {
	user, err := s.LoginService.CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.ListVmResponse{Code: protocol.Code_BAD_TOKEN, Id: []int32{}}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}
	virtualMachines := s.Database.Collection("virtualMachines")
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

	return &protocol.ListVmResponse{Code: protocol.Code_OK, Id: id}, nil
}

func (s *VmServer) FreeResources(c context.Context, request *protocol.JustTokenRequest) (*protocol.FreeResourcesResponse, error) {
	user, err := s.LoginService.CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.FreeResourcesResponse{Code: protocol.Code_BAD_TOKEN, Free: nil, Total: nil}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	free, err := GetFreeResources(*user)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	total := user.Quota

	freePrt := protocol.VmSpecification{Storage: int32(free.Storage), Memory: int32(free.Memory), Core: int32(free.Cores)}
	totalPrt := protocol.VmSpecification{Storage: int32(total.Storage), Memory: int32(total.Memory), Core: int32(total.Cores)}

	return &protocol.FreeResourcesResponse{Code: protocol.Code_OK, Free: &freePrt, Total: &totalPrt}, nil
}

func (s *VmServer) Delete(c context.Context, request *protocol.VmRequest) (*protocol.StatusVmResponse, error) {
	user, err := s.LoginService.CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.StatusVmResponse{Code: protocol.Code_BAD_TOKEN, Status: protocol.Status_ABORTED}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	vm, err := GetVirtualMachine(s.Database, int(request.Id))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	if vm.Owner != user.Login {
		return &protocol.StatusVmResponse{Code: protocol.Code_NOT_ALLOWED, Status: protocol.Status_ABORTED}, nil
	}

	err = Delete(s.Proxmox, s.Database, vm)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	return &protocol.StatusVmResponse{Code: protocol.Code_OK, Status: protocol.Status_DELETED}, nil
}

func (s *VmServer) Modify(c context.Context, request *protocol.ModifyVmRequest) (*protocol.StatusVmResponse, error) {
	user, err := s.LoginService.CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.StatusVmResponse{Code: protocol.Code_BAD_TOKEN, Status: protocol.Status_ABORTED, Message: ""}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	vm, err := GetVirtualMachine(s.Database, int(request.Id))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	if vm.Owner != user.Login {
		return &protocol.StatusVmResponse{Code: protocol.Code_NOT_ALLOWED, Status: protocol.Status_ABORTED}, nil
	}

	originalSpec, err := vm.GetSpecification()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	spec := Specification{Cores: int(request.Specification.Core), Memory: int(request.Specification.Memory), Storage: int(request.Specification.Storage)}
	if vm.UseOwnerQuota {
		err = spec.CheckFreeResourcesWithout(*user, originalSpec)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}

	err = vm.EditSpecification(spec)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	return &protocol.StatusVmResponse{Code: protocol.Code_OK, Status: protocol.Status_PREPARED, Message: ""}, nil
}
