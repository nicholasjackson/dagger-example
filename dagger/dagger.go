package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
  err := apply()
  if err != nil {
    fmt.Println(err)
		os.Exit(1)
  }

	fmt.Println("Succesfully deployed")
}

func apply() error {
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

  os.WriteFile("./credentials.tfrc.json", []byte(fmt.Sprintf(`  
{
  "credentials": {
    "app.terraform.io": {
      "token": "%s"
    }
  }
}`, os.Getenv("TFE_TOKEN"))), 0644)

  defer func() {
    os.Remove("./credentials.tfrc.json")
  }()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return fmt.Errorf("Error connecting to Dagger Engine: %s", err)
	}

	defer client.Close()

	src := client.Host().Workdir()
	if err != nil {
		return fmt.Errorf("Error getting reference to host directory: %s", err)
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
	build := golang.Directory(path)

  _, err = client.Container().From("alpine:latest").
    WithMountedDirectory("/tmp", build).
    Exec(dagger.ContainerExecOpts{ 
      Args: []string{"cp", "/tmp/dagger-example","/bin/dagger-example"},
    }).
    WithEntrypoint([]string{"/bin/dagger-example"}).
		Publish(ctx, "nicholasjackson/dagger-example:latest")

	if err != nil {
		return fmt.Errorf("Error creating and pushing container: %s", err)
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
    WithMountedFile("/root/.terraform.d/credentials.tfrc.json", creds).
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
  
	if err != nil {
		return fmt.Errorf("Error deploying application to DigitalOcean: %s", err)
	}

  return nil
}
