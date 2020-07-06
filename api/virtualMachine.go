package main

import (
	"context"
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strconv"
	"strings"
)

type VirtualMachine struct {
	Name          string
	Id            int
	StatusCode    protocol.StatusVmResponse_Status
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

func GetVirtualMachine(id int) (VirtualMachine, error) {
	virtualMachines := database.Collection("virtualMachines")
	vm := VirtualMachine{}
	err := virtualMachines.FindOne(context.Background(), bson.M{"id": id}).Decode(&vm)
	return vm, err
}

func (vm *VirtualMachine) Created() bool {
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
	vmRef, err := GetVmRefById(vm.Id)
	if err != nil {
		return fmt.Errorf("Start vm error : %s", err.Error())
	}
	_, err = proxmoxClient.StartVm(vmRef)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VirtualMachine) Stop() error {
	vmRef, err := GetVmRefById(vm.Id)
	if err != nil {
		return fmt.Errorf("Stop vm error : %s", err.Error())
	}
	_, err = proxmoxClient.StopVm(vmRef)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VirtualMachine) Delete() error {
	if vm.Created() {
		vmRef, err := GetVmRefById(vm.Id)
		if err != nil {
			return fmt.Errorf("Delete vm error : %s", err.Error())
		}
		_, err = proxmoxClient.DeleteVm(vmRef)
		if err != nil {
			return err
		}
	}
	c := context.Background()
	virtualMachines := database.Collection("virtualMachines")

	_, _ = virtualMachines.DeleteOne(c, bson.M{"id": vm.Id})

	return nil
}

func (vm *VirtualMachine) Sync() error {
	virtualMachines := database.Collection("virtualMachines")
	_, err := virtualMachines.UpdateOne(context.Background(), bson.M{"id": vm.Id}, bson.M{"$set": vm})
	return err
}

func (spec *Specification) CheckMinimumResources() error {
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

func GetVmRefById(id int) (*proxmox.VmRef, error) {
	vmRef := proxmox.NewVmRef(id)
	err := proxmoxClient.CheckVmRef(vmRef)
	if err != nil {
		/*
			if err.Error() == fmt.Sprintf("Vm '%d' not found", id) {
				virtualMachines := database.Collection("virtualMachines")
				_, _ = virtualMachines.DeleteOne(context.Background(), bson.M{"id": id})
			}*/
		return nil, err
	}
	return vmRef, nil
}

func (spec *Specification) CheckFreeResources(user User) error {
	freeResources, err := user.GetFreeResources()
	if err != nil {
		return err
	}
	if spec.Cores > freeResources.Cores {
		return fmt.Errorf("You will exceed your CPU quota by %d core", spec.Cores-freeResources.Cores)
	}
	if spec.Storage > freeResources.Storage {
		return fmt.Errorf("You will exceed your storage quota by %d Mb", spec.Storage-freeResources.Storage)
	}
	if spec.Memory > freeResources.Memory {
		return fmt.Errorf("You will exceed your memory quota by %d Mb", spec.Memory-freeResources.Memory)
	}

	return nil
}

func (user *User) GetUsedResources() (Specification, error) {
	c := context.Background()

	virtualMachines := database.Collection("virtualMachines")
	vms, err := virtualMachines.Find(c, bson.M{"owner": user.Login})
	if err != nil {
		return Specification{}, err
	}

	var result Specification
	for vms.Next(c) {

		var vm VirtualMachine
		err := vms.Decode(&vm)
		if err != nil {
			return Specification{}, err
		}

		if !vm.Created() {
			continue
		}

		vmRef, err := GetVmRefById(vm.Id)
		if err != nil {
			continue
		}

		config, _ := proxmoxClient.GetVmConfig(vmRef)

		if vm.UseOwnerQuota {
			diskInfo, conv := config["scsi0"].(string)
			if !conv {
				return Specification{}, err
			}
			cores, conv := config["cores"].(float64)
			if !conv {
				return Specification{}, err
			}
			memory, conv := config["memory"].(float64)
			if !conv {
				return Specification{}, err
			}
			storage, err := strconv.Atoi(diskInfo[strings.Index(diskInfo, "=")+1 : len(diskInfo)-1])
			if err != nil {
				return Specification{}, err
			}

			result.Cores += int(cores)
			result.Memory += int(memory)
			result.Storage += storage
		}
	}

	return result, nil
}

func (user *User) GetFreeResources() (Specification, error) {
	result := user.Quota
	used, err := user.GetUsedResources()
	if err != nil {
		return Specification{}, err
	}

	result.Cores -= used.Cores
	result.Memory -= used.Memory
	result.Storage -= used.Storage

	return result, nil
}
