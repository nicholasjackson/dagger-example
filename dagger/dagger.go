package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		fmt.Printf("Error connecting to Dagger Engine: %s", err)
		os.Exit(1)
	}

	defer client.Close()

	src := client.Host().Workdir()
	if err != nil {
		fmt.Printf("Error getting reference to host directory: %s", err)
		os.Exit(1)
	}

	golang := client.Container().From("golang:latest")
	golang = golang.WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0")

	golang = golang.Exec(
		dagger.ContainerExecOpts{
			Args: []string{"go", "build", "-o", "build/"},
		},
	)

	path := "build/"
	err = os.MkdirAll(filepath.Join(".", path), os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating output folder: %s", err)
		os.Exit(1)
	}

	build := golang.Directory(path)

	_, err = build.Export(ctx, path)
	if err != nil {
		fmt.Printf("Error writing directory: %s", err)
		os.Exit(1)
	}

	cn, err := client.Container().
		Build(src).
		Publish(ctx, "nicholasjackson/dagger-example:latest")

	if err != nil {
		fmt.Printf("Error creating and pushing container: %s", err)
		os.Exit(1)
	}

	deploy := client.Host().Workdir().
		Directory("./deploy").
		WithoutDirectory("cdktf.out")

	cdktf := client.Container().From("nicholasjackson/cdktf:latest").
		WithEnvVariable("DIGITALOCEAN_TOKEN", os.Getenv("DIGITALOCEAN_TOKEN")).
		WithMountedDirectory("/src", deploy).
		WithWorkdir("/src").
		WithEntrypoint([]string{})

	cdktf = cdktf.Exec(
		dagger.ContainerExecOpts{
			Args: []string{"cdktf", "get"},
		},
	).Exec(
		dagger.ContainerExecOpts{
			Args: []string{"cdktf", "apply", "--auto-approve"},
		},
	)

	state := cdktf.File("./terraform.src.tfstate")
	_, err = state.Export(ctx, "./deploy/terraform.src.tfstate")

	if err != nil {
		fmt.Printf("Error deploying application to DigitalOcean: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Succesfully created new container: %s", cn)
}
