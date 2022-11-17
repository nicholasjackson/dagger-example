package main

import (
	"flag"
	"fmt"
	"os"

	"dagger.io/dagger"
)

var doDeploy = flag.Bool("do-deploy", false, "Should we deploy the application")
var ref = flag.String("ref", "dev", "tag or branch or sha")

func main() {
	flag.Parse()

	build, err := NewBuild()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	apply(build)
	if build.HasError() {
		os.Exit(1)
	}
}

func apply(build *Build) {
	clean := generateCredentials()
	defer clean()

	app := buildApplication(build)

	packageApplication(build, app, *ref)
}

func generateCredentials() func() {
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

	return func() {
		os.Remove("./credentials.tfrc.json")
	}
}

func buildApplication(build *Build) *dagger.File {
	done := build.LogStart("Application Build")
	defer done()

	src := build.client.Host().Workdir()

	golang := build.client.Container().From("golang:latest")
	golang = golang.WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0")

	golang = golang.Exec(
		dagger.ContainerExecOpts{
			Args: []string{"go", "build", "-o", "build/dagger-example"},
		},
	)

	return golang.Directory("./build").File("dagger-example")
}

func packageApplication(build *Build, app *dagger.File, branch string) string {
	if build.Cancelled() {
		return ""
	}

	done := build.LogStart("Package Application")
	defer done()

	_, err := build.client.Container().
		From("alpine:latest").
		WithMountedFile("/tmp", app).
		Exec(dagger.ContainerExecOpts{
			Args: []string{"cp", "/tmp/dagger-example", "/bin/dagger-example"},
		}).
		WithEntrypoint([]string{"/bin/dagger-example"}).
		Publish(build.ContextWithTimeout(DefaultTimeout), "nicholasjackson/dagger-example:"+branch)

	if err != nil {
		build.LogError(fmt.Errorf("Error creating and pushing container: %s", err))
	}

	return "nicholasjackson/dagger-example:" + branch
}

func deployApplicaton(build *Build) {
	if !*doDeploy {
		return
	}

	if build.HasError() {
		return
	}

	done := build.LogStart("Deploy Application")
	defer done()

	deploy := build.client.Host().Workdir().
		Directory("./deploy").
		WithoutDirectory("cdktf.out")

	creds := build.client.Host().Workdir().File("./credentials.tfrc.json")

	cdktf := build.client.Container().From("nicholasjackson/cdktf:1.3.4").
		WithEnvVariable("DIGITALOCEAN_TOKEN", os.Getenv("DIGITALOCEAN_TOKEN")).
		WithMountedDirectory("/src", deploy).
		WithMountedFile("/root/.terraform.d/credentials.tfrc.json", creds).
		WithWorkdir("/src").
		WithEntrypoint([]string{})

	_, err := cdktf.Exec(
		dagger.ContainerExecOpts{
			Args: []string{"cdktf", "get"},
		},
	).Exec(
		dagger.ContainerExecOpts{
			Args: []string{"cdktf", "apply", "--auto-approve"},
		},
	).ExitCode(build.ContextWithTimeout(DefaultTimeout))

	if err != nil {
		build.LogError(fmt.Errorf("Error deploying application to DigitalOcean: %s", err))
	}
}
