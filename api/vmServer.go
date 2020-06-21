package main

import (
	"context"
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type VmServer struct {
	protocol.UnimplementedVmServiceServer
}

type VirtualMachine struct {
	Name       string
	Id         int
	StatusCode protocol.StatusVmResponse_Status
	Spec       Spec
	Error      string
}

type Spec struct {
	cpu  int
	ram  int
	disk int
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
	_, err := CheckToken(request.Token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_BAD_TOKEN, Name: "", Id: 0}, nil
		} else {
			return nil, status.Errorf(codes.Aborted, err.Error())
		}
	}
	if request.Spec == nil {
		return nil, status.Errorf(codes.Aborted, "Spec does not exist")
	}
	spec := Spec{int(request.Spec.Cpus), int(request.Spec.Ram), int(request.Spec.Disk)}

	err = CheckSpec(spec)
	if err != nil {
		return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_NOT_ENOUGH_RESOURCES, Name: err.Error(), Id: 0}, nil
	}

	vmId, err := proxmoxClient.NextId()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	virtualMachines := database.Collection("virtualMachines")

	_, _ = virtualMachines.DeleteOne(c, bson.M{"id": vmId})

	vm := VirtualMachine{Name: request.Name, Id: vmId, StatusCode: protocol.StatusVmResponse_PREPARED, Error: "", Spec: spec}
	_, err = virtualMachines.InsertOne(c, vm)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	go VmCreationProcess(vm)

	return &protocol.CreateVmResponse{Code: protocol.CreateVmResponse_OK, Name: request.Name, Id: int32(vmId)}, nil
}

func CheckSpec(spec Spec) error {
	if spec.cpu < 1 {
		return fmt.Errorf("Vm must have at least 1 CPU")
	}
	if spec.disk < 2252 {
		return fmt.Errorf("Vm must have at least 2252 Mb HDD")
	}
	if spec.ram < 512 {
		return fmt.Errorf("Vm must have at least 512 Mb RAM")
	}

	return nil
}

func VmCreationProcess(vm VirtualMachine) {
	err := CreateVM(vm)
	if err != nil {
		vm.Fatal(fmt.Errorf("Create error : %s", err.Error()))
		return
	}

	err = SetupVM(vm)
	if err != nil {
		vm.Fatal(fmt.Errorf("Setup error : %s", err.Error()))
		return
	}
	vm.StatusCode = protocol.StatusVmResponse_STOPPED
	err = vm.Sync()
	if err != nil {
		vm.Fatal(fmt.Errorf("VmCreationProcess : %s", err.Error()))
		return
	}

}

func CreateVM(vm VirtualMachine) error {
	vm.StatusCode = protocol.StatusVmResponse_CREATED
	err := vm.Sync()
	if err != nil {
		return err
	}
	vmParent, err := proxmoxClient.GetVmRefByName("VM 9000")
	if err != nil {
		return err
	}

	cloneParams := map[string]interface{}{
		"newid":  vm.Id,
		"name":   vm.Name,
		"target": "spacex",
		"full":   false,
	}

	_, err = proxmoxClient.CloneQemuVm(vmParent, cloneParams)
	if err != nil {
		return err
	}

	timeout := 1 * time.Minute
	start := time.Now().Unix()
	for !IsCreated(vm) {
		if time.Now().Unix()-start > int64(timeout.Seconds()) {
			return fmt.Errorf("VM not created")
		}
		time.Sleep(1 * time.Second)
		log.Printf("%d vm did not find", vm.Id)
	}

	return nil
}

func SetupVM(vm VirtualMachine) error {
	vm.StatusCode = protocol.StatusVmResponse_SETUP
	err := vm.Sync()
	if err != nil {
		return fmt.Errorf("Syncing error : %s", err.Error())
	}
	vmRef := proxmox.NewVmRef(vm.Id)
	err = proxmoxClient.CheckVmRef(vmRef)
	if err != nil {
		return fmt.Errorf("Setup vm error : %s", err.Error())
	}

	param := map[string]interface{}{
		"ciuser":     "louis",
		"cipassword": "louis",
		"memory":     vm.Spec.ram,
		"cores":      vm.Spec.cpu,
	}

	_, err = proxmoxClient.SetVmConfig(vmRef, param)
	if err != nil {
		return fmt.Errorf("Setup vm error : %s", err.Error())
	}

	_, err = proxmoxClient.ResizeQemuDisk(vmRef, "scsi0", vm.Spec.disk-2252)
	if err != nil {
		return fmt.Errorf("Resize disk error : %s", err.Error())
	}

	return nil
}

func UpdateStatus(vm VirtualMachine) error {
	virtualMachines := database.Collection("virtualMachines")
	_, err := virtualMachines.UpdateOne(context.Background(), bson.M{"id": vm.Id}, bson.M{"$set": vm})
	return err
}

func IsCreated(vm VirtualMachine) bool {
	_, err := proxmoxClient.VMIdExists(vm.Id)
	return err != nil
}

func GetVirtualMachine(id int) (VirtualMachine, error) {
	virtualMachines := database.Collection("virtualMachines")
	vm := VirtualMachine{}
	err := virtualMachines.FindOne(context.Background(), bson.M{"id": id}).Decode(&vm)
	return vm, err
}

func (vm *VirtualMachine) Fatal(err error) {
	vm.StatusCode = protocol.StatusVmResponse_ABORTED
	vm.Error = err.Error()
	_ = vm.Sync()
	log.Printf(err.Error())
}

func (vm *VirtualMachine) Start() error {
	vmRef := proxmox.NewVmRef(vm.Id)
	err := proxmoxClient.CheckVmRef(vmRef)
	if err != nil {
		return fmt.Errorf("Setup vm error : %s", err.Error())
	}
	_, err = proxmoxClient.StartVm(vmRef)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VirtualMachine) Sync() error {
	return UpdateStatus(*vm)
}
