package main

import (
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/six3six/SpaceHoster/api/protocol"
	"log"
	"time"
)

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
	for !vm.IsCreated() {
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
		"memory":     vm.Spec.Memory,
		"cores":      vm.Spec.Cores,
	}

	_, err = proxmoxClient.SetVmConfig(vmRef, param)
	if err != nil {
		return fmt.Errorf("Setup vm error : %s", err.Error())
	}

	_, err = proxmoxClient.ResizeQemuDisk(vmRef, "scsi0", vm.Spec.Storage-2252)
	if err != nil {
		return fmt.Errorf("Resize storage error : %s", err.Error())
	}

	return nil
}
