package main

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/guths/cube/task"
)

func main() {
	fmt.Println("Running")

	dockerTask, createResult := createContainer()
	if createResult.Error != nil {
		fmt.Printf("%v", createResult.Error)
		os.Exit(1)
	}

	time.Sleep(time.Second * 5)
	fmt.Printf("stopping container %s\n", createResult.ContainerId)
	_ = stopContainer(dockerTask, createResult.ContainerId)
}

func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=cube",
			"POSTGRES_PASSWORD=secret",
		},
	}

	dc, _ := client.NewClientWithOpts(client.FromEnv)

	d := task.Docker{
		Client: dc,
		Config: c,
	}

	result := d.Run()

	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil, nil
	}

	fmt.Printf("Container %s is running with confi %v\n", result.ContainerId, c)

	return &d, &result
}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
	result := d.Stop(id)
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil
	}

	fmt.Printf(
		"Container %s has been stopped and removed\n", result.ContainerId)
	return &result
}
