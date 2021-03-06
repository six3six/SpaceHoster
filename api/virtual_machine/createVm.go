package virtual_machine

import (
	"fmt"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func VmCreationProcess(vm VirtualMachine, database *mongo.Database, login string, password string, specification Specification) {
	err := CreateVM(vm, database)
	if err != nil {
		vm.Fatal(database, fmt.Errorf("Create error : %s", err.Error()))
		return
	}

	err = SetupVM(vm, login, password, specification, database)
	if err != nil {
		vm.Fatal(database, fmt.Errorf("Setup error : %s", err.Error()))
		return
	}
	vm.StatusCode = protocol.Status_STOPPED
	err = Sync(database, vm)
	if err != nil {
		vm.Fatal(database, fmt.Errorf("VmCreationProcess : %s", err.Error()))
		return
	}
}

func CreateVM(vm VirtualMachine, database *mongo.Database) error {
	vm.StatusCode = protocol.Status_CREATED
	err := Sync(database, vm)
	if err != nil {
		return err
	}
	vmParent, err := vm.Proxmox.GetVmRefByName("VM 9000")
	if err != nil {
		return err
	}

	cloneParams := map[string]interface{}{
		"newid":  vm.Id,
		"name":   vm.Name,
		"target": "spacex",
		"full":   false,
	}

	_, err = vm.Proxmox.CloneQemuVm(vmParent, cloneParams)
	if err != nil {
		return err
	}

	timeout := 1 * time.Minute
	start := time.Now().Unix()
	for !vm.Created() {
		if time.Now().Unix()-start > int64(timeout.Seconds()) {
			return fmt.Errorf("VM not created")
		}
		time.Sleep(1 * time.Second)
		log.Printf("%d vm did not find", vm.Id)
	}

	return nil
}

func SetupVM(vm VirtualMachine, login string, password string, specification Specification, database *mongo.Database) error {
	vm.StatusCode = protocol.Status_SETUP
	err := Sync(database, vm)
	if err != nil {
		return fmt.Errorf("Syncing error : %s", err.Error())
	}

	err = vm.EditSpecification(specification)
	if err != nil {
		return fmt.Errorf("Setup vm error (spec mod) : %s", err.Error())
	}

	err = vm.EditLogin(login, password)
	if err != nil {
		return fmt.Errorf("Setup vm error (login mod) : %s", err.Error())
	}

	return nil
}
