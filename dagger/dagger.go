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

  src,err := client.Host().Workdir().Read().ID(ctx)
  if err != nil {
    fmt.Printf("Error getting reference to host directory: %s", err)
    os.Exit(1)
  }

  golang := client.Container().From("golang:latest")
  golang = golang.WithMountedDirectory("/src", src).WithWorkdir("/src").WithEnvVariable("CGO_ENABLED","0")

  golang = golang.Exec(dagger.ContainerExecOpts{
    Args: []string{"go", "build", "-o", "build/"},
  })
  

  path := "build/"
  err = os.MkdirAll(filepath.Join(".", path), os.ModePerm)
  if err != nil {
    fmt.Printf("Error creating output folder: %s", err)
    os.Exit(1)
  }

  build,err := golang.Directory(path).ID(ctx)
  
  if err != nil {
    fmt.Printf("Error fetching data from container: %s", err)
    os.Exit(1)
  }

  workdir := client.Host().Workdir() 
  _,err = workdir.Write(ctx, build, dagger.HostDirectoryWriteOpts{
    Path: path,
  })
  
  if err != nil {
    fmt.Printf("Error writing directory: %s", err)
    os.Exit(1)
  }

  cn,err := client.Container().Build(src).Publish(ctx, "nicholasjackson/dagger-example:latest")
  if err != nil {
    fmt.Printf("Error creating and pushing container: %s", err)
    os.Exit(1)
  }

  fmt.Printf("Succesfully created new container: %s", cn)

}
