package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()
  
  // add the token to the backend
  // need to do this here or it is not picked up
  os.WriteFile("./credentials.tfrc.json", []byte(fmt.Sprintf(`  
{
  "credentials": {
    "app.terraform.io": {
      "token": "%s"
    }
  }
}`, os.Getenv("TFE_TOKEN"))), 0644)

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

	_, err = client.Container().
		Build(src).
		Publish(ctx, "nicholasjackson/dagger-example:latest")

	if err != nil {
		fmt.Printf("Error creating and pushing container: %s", err)
		os.Exit(1)
	}

	deploy := client.Host().Workdir().
		Directory("./deploy").
		WithoutDirectory("cdktf.out")

  creds := client.Host().Workdir().File("./credentials.tfrc.json")

	cdktf := client.Container().From("nicholasjackson/cdktf:1.3.4").
		WithEnvVariable("DIGITALOCEAN_TOKEN", os.Getenv("DIGITALOCEAN_TOKEN")).
    WithEnvVariable("CACHEBUST", fmt.Sprintf("%d", time.Now().Nanosecond())).
    WithMountedFile("/root/.terraform.d/credentials.tfrc.json", creds).
		WithMountedDirectory("/src", deploy).
		WithWorkdir("/src").
		WithEntrypoint([]string{})

	_,err = cdktf.Exec(
		dagger.ContainerExecOpts{
			Args: []string{"cdktf", "get"},
		},
	).Exec(
		dagger.ContainerExecOpts{
			Args: []string{"cdktf", "apply", "--auto-approve"},
		},
	).ExitCode(ctx)

  // remove the credentials
  os.Remove("./credentials.tfrc.json")
	
  if err != nil {
		fmt.Printf("Error deploying application: %s", err)
		os.Exit(1)
	}
}
