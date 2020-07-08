package virtual_machine

import (
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/six3six/SpaceHoster/api/login"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strconv"
	"strings"
)

type VirtualMachine struct {
	Name          string
	Id            int
	StatusCode    protocol.Status
	Error         string
	Owner         login.Login
	UseOwnerQuota bool
	Editors       []login.Login
	Proxmox       *proxmox.Client
}

func (vm *VirtualMachine) Created() bool {
	_, err := vm.Proxmox.VMIdExists(vm.Id)
	return err != nil
}

func (vm *VirtualMachine) Fatal(database *mongo.Database, err error) {
	vm.StatusCode = protocol.Status_ABORTED
	vm.Error = err.Error()
	_ = Sync(database, *vm)
	log.Printf(err.Error())
}

func (vm *VirtualMachine) Start() error {
	vmRef, err := GetVmRefById(vm.Proxmox, vm.Id)
	if err != nil {
		return fmt.Errorf("Start vm error : %s", err.Error())
	}
	_, err = vm.Proxmox.StartVm(vmRef)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VirtualMachine) Stop() error {
	vmRef, err := GetVmRefById(vm.Proxmox, vm.Id)
	if err != nil {
		return fmt.Errorf("Stop vm error : %s", err.Error())
	}
	_, err = vm.Proxmox.StopVm(vmRef)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VirtualMachine) EditSpecification(specification Specification) error {

	err := specification.CheckMinimumResources()
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}

	vmRef, err := GetVmRefById(vm.Proxmox, vm.Id)
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

	_, err = vm.Proxmox.SetVmConfig(vmRef, param)
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}

	_, err = vm.Proxmox.ResizeQemuDisk(vmRef, "scsi0", specification.Storage-actualSpec.Storage)
	if err != nil {
		return fmt.Errorf("Resize storage error : %s", err.Error())
	}

	return nil
}

func (vm *VirtualMachine) EditLogin(login string, password string) error {

	vmRef, err := GetVmRefById(vm.Proxmox, vm.Id)
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}

	param := map[string]interface{}{
		"ciuser":     login,
		"cipassword": password,
	}

	_, err = vm.Proxmox.SetVmConfig(vmRef, param)
	if err != nil {
		return fmt.Errorf("Edit vm error : %s", err.Error())
	}
	return nil
}

func (vm *VirtualMachine) GetSpecification() (Specification, error) {
	vmRef, err := GetVmRefById(vm.Proxmox, vm.Id)
	if err != nil {
		return Specification{}, fmt.Errorf("GetSpecification error : %s", err.Error())
	}

	config, err := vm.Proxmox.GetVmConfig(vmRef)

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
