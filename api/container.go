package main

import (
	"github.com/gin-gonic/gin"
)

type Container struct {
	DockerName string
	Image      string
	Owner      string
	Managers   []string
	PublicName string
}

func CreateNewVm() {

}

func CreateNewVmHandler(c *gin.Context) {
}

/*
func CreateNewVM(c *gin.Context) {
	vmParamsCollection := database.Collection("vmParams")

	var vmCountParam struct {
		data string
		vmId int
	}
	err := vmParamsCollection.FindOne(c, bson.M{"data": "count"}).Decode(&vmCountParam)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			vmCountParam = struct {
				data string
				vmId int
			}{"count", 400}
			_, err := vmParamsCollection.InsertOne(c, vmCountParam)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"context": "Inserting vm enumaration",
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"context": "Searching for vm enumaration",
				"message": err.Error(),
			})
			return
		}
	}

	vmName := "testvm"
	vmNode := "spacex"
	vmId := vmCountParam.vmId
	user := "louis"
	password := "password"

	vmCountParam.vmId++
	_, err = vmParamsCollection.UpdateOne(c, bson.M{"data": "count"}, vmCountParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Updating count",
			"message": err.Error(),
		})
		return
	}

	modelRef, err := proxmoxClient.GetVmRefByName("VM 9000")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Searching for model",
			"message": err.Error(),
		})
		return
	}

	vmParams := map[string]interface{}{
		"newid":  vmId,
		"name":   vmName,
		"target": vmNode,
		"full":   false,
	}
	c.JSON(http.StatusOK, vmParams)
	exitCloneStatus, err := proxmoxClient.CloneQemuVm(modelRef, vmParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Cloning model",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": exitCloneStatus})
	err = nil
	var vmRef *proxmox.VmRef
	for i := 0; i < 50; i++ {
		vmRef, err = proxmoxClient.GetVmRefByName(vmName)
		if err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Searching cloned vm",
			"message": err.Error(),
		})
		return
	}
	vmConfig := map[string]interface{}{
		"ciuser":     user,
		"cipassword": password,
	}
	exitConfigStatus, err := proxmoxClient.SetVmConfig(vmRef, vmConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Configuring cloned vm",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": exitConfigStatus})
}
*/
