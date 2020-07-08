package virtual_machine

import (
	"context"
	"fmt"
	"github.com/six3six/SpaceHoster/api/common"
	"github.com/six3six/SpaceHoster/api/login"
	"go.mongodb.org/mongo-driver/bson"
)

type Specification common.Resource

func (spec *Specification) CheckFreeResources(user login.User) error {
	freeResources, err := GetFreeResources(user)
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

func (spec *Specification) CheckFreeResourcesWithout(user login.User, without Specification) error {
	freeResources, err := GetFreeResourcesWithout(user, without)
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

func GetUsedResources(user login.User) (Specification, error) {
	c := context.Background()

	virtualMachines := user.Database.Collection("virtualMachines")
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

func GetFreeResourcesWithout(user login.User, without Specification) (Specification, error) {
	result := user.Quota
	used, err := GetUsedResources(user)
	if err != nil {
		return Specification{}, err
	}

	result.Cores -= used.Cores - without.Cores
	result.Memory -= used.Memory - without.Cores
	result.Storage -= used.Storage - without.Cores

	return Specification(result), nil
}

func GetFreeResources(user login.User) (Specification, error) {
	return GetFreeResourcesWithout(user, Specification{})
}
