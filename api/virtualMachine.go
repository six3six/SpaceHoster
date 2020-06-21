package main

import (
	"context"
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

type VirtualMachine struct {
	Name          string
	Id            int
	StatusCode    protocol.StatusVmResponse_Status
	Spec          Specification
	Error         string
	Owner         Login
	UseOwnerQuota bool
	Editors       []Login
}

type Specification struct {
	Cores   int
	Memory  int
	Storage int
}

var defaultSpecification = Specification{
	2,
	1024,
	30000,
}

func GetVirtualMachine(id int) (VirtualMachine, error) {
	virtualMachines := database.Collection("virtualMachines")
	vm := VirtualMachine{}
	err := virtualMachines.FindOne(context.Background(), bson.M{"id": id}).Decode(&vm)
	return vm, err
}

func (vm *VirtualMachine) IsCreated() bool {
	_, err := proxmoxClient.VMIdExists(vm.Id)
	return err != nil
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
	virtualMachines := database.Collection("virtualMachines")
	_, err := virtualMachines.UpdateOne(context.Background(), bson.M{"id": vm.Id}, bson.M{"$set": vm})
	return err
}

func (spec *Specification) CheckSpec() error {
	if spec.Cores < 1 {
		return fmt.Errorf("Vm must have at least 1 CPU")
	}
	if spec.Storage < 2252 {
		return fmt.Errorf("Vm must have at least 2252 Mb HDD")
	}
	if spec.Memory < 512 {
		return fmt.Errorf("Vm must have at least 512 Mb RAM")
	}

	return nil
}
