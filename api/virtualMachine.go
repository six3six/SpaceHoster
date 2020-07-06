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

func (vm *VirtualMachine) EditSpecification(specification Specification) error {

	err := specification.CheckMinimumResources()
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}

	vmRef, err := GetVmRefById(vm.Id)
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}

	actualSpec, err := vm.GetSpecification()
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}

	if specification.Storage < actualSpec.Storage {
		return fmt.Errorf("Edit vm error : cannot decrease disk size")
	}

	param := map[string]interface{}{
		"memory": specification.Memory,
		"cores":  specification.Cores,
	}

	_, err = proxmoxClient.SetVmConfig(vmRef, param)
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}

	_, err = proxmoxClient.ResizeQemuDisk(vmRef, "scsi0", specification.Storage-actualSpec.Storage)
	if err != nil {
		return fmt.Errorf("Resize storage error : %s", err.Error())
	}

	return nil
}

func (vm *VirtualMachine) EditLogin(login string, password string) error {

	vmRef, err := GetVmRefById(vm.Id)
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}

	param := map[string]interface{}{
		"ciuser":     login,
		"cipassword": password,
	}

	_, err = proxmoxClient.SetVmConfig(vmRef, param)
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}
	return nil
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

func (spec *Specification) CheckFreeResourcesWithout(user User, without Specification) error {
	freeResources, err := user.GetFreeResourcesWithout(without)
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

func (spec *Specification) CheckFreeResources(user User) error {
	return spec.CheckFreeResourcesWithout(user, Specification{0, 0, 0})
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

		if vm.UseOwnerQuota {
			specs, err := vm.GetSpecification()
			if err != nil {
				return Specification{}, err
			}

			result.Storage -= specs.Storage
			result.Memory -= specs.Memory
			result.Cores -= specs.Cores
		}

	}

	return result, nil
}

func (vm VirtualMachine) GetSpecification() (Specification, error) {
	vmRef, err := GetVmRefById(vm.Id)
	if err != nil {
		return Specification{}, fmt.Errorf("GetSpecification error : %s", err.Error())
	}

	config, err := proxmoxClient.GetVmConfig(vmRef)

	diskInfo, conv := config["scsi0"].(string)
	if !conv {
		return Specification{}, fmt.Errorf("Can't convert scsi0 var from %d vm", vm.Id)
	}
	cores, conv := config["cores"].(float64)
	if !conv {
		return Specification{}, fmt.Errorf("Can't convert cores var from %d vm", vm.Id)
	}
	memory, conv := config["memory"].(float64)
	if !conv {
		return Specification{}, fmt.Errorf("Can't convert memory var from %d vm", vm.Id)
	}
	storage, err := strconv.Atoi(diskInfo[strings.Index(diskInfo, "=")+1 : len(diskInfo)-1])
	if err != nil {
		return Specification{}, fmt.Errorf("GetSpecification error : %s", err.Error())
	}

	var result Specification
	result.Cores += int(cores)
	result.Memory += int(memory)
	result.Storage += storage

	return result, nil
}

func (user *User) GetFreeResourcesWithout(without Specification) (Specification, error) {
	result := user.Quota
	used, err := user.GetUsedResources()
	if err != nil {
		return Specification{}, err
	}

	result.Cores -= used.Cores - without.Cores
	result.Memory -= used.Memory - without.Cores
	result.Storage -= used.Storage - without.Cores

	return result, nil
}

func (user *User) GetFreeResources() (Specification, error) {
	return user.GetFreeResourcesWithout(Specification{0, 0, 0})
}
