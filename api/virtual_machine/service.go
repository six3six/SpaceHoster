package virtual_machine

import (
	"context"
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetVirtualMachine(database *mongo.Database, id int) (VirtualMachine, error) {
	virtualMachines := database.Collection("virtualMachines")
	vm := VirtualMachine{}
	err := virtualMachines.FindOne(context.Background(), bson.M{"id": id}).Decode(&vm)
	return vm, err
}

func GetVmRefById(proxmoxClient *proxmox.Client, id int) (*proxmox.VmRef, error) {
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

func Delete(proxmoxClient *proxmox.Client, database *mongo.Database, vm VirtualMachine) error {
	if vm.Created() {
		vmRef, err := GetVmRefById(proxmoxClient, vm.Id)
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

func Sync(database *mongo.Database, vm VirtualMachine) error {
	virtualMachines := database.Collection("virtualMachines")
	_, err := virtualMachines.UpdateOne(context.Background(), bson.M{"id": vm.Id}, bson.M{"$set": vm})
	return err
}
