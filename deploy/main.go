package main

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	digitalocean "github.com/hashicorp/cdktf-provider-digitalocean-go/digitalocean/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

func NewMyStack(scope constructs.Construct, id string) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)

	digitalocean.NewDigitaloceanProvider(stack, jsii.String("digitalocean"), &digitalocean.DigitaloceanProviderConfig{})

	app := digitalocean.NewApp(stack, jsii.String("dagger"), &digitalocean.AppConfig{
		Spec: &digitalocean.AppSpec{
			Name:   jsii.String("dagger"),
			Region: jsii.String("ams"),
			Service: []*digitalocean.AppSpecService{
				&digitalocean.AppSpecService{
					Name:     jsii.String("dagger"),
					HttpPort: jsii.Number(9090),
					Image: &digitalocean.AppSpecServiceImage{
						RegistryType: jsii.String("DOCKER_HUB"),
						Registry:     jsii.String("nicholasjackson"),
						Repository:   jsii.String("dagger-example"),
						Tag:          jsii.String("latest"),
					},
					InstanceSizeSlug: jsii.String("basic-xxs"),
				},
			},
		},
	})

	cdktf.NewTerraformOutput(stack, jsii.String("live_url"), &cdktf.TerraformOutputConfig{Value: app.LiveUrl()})

	return stack
}

func main() {
	app := cdktf.NewApp(nil)

	NewMyStack(app, "src")

	app.Synth()
}
