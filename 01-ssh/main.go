package main

import (
	"fmt"
	"github.com/insysnw/infra/pkg"
	"github.com/pulumi/pulumi-digitalocean/sdk/v3/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"io/ioutil"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		studentsKeysFilePath := "../students.keys"
		teachersKeysFilePath := "../teachers.keys"
		region := "fra1"

		// Script to execute on every machine
		content, err := ioutil.ReadFile("script.sh")
		if err != nil {
			fmt.Println("Error: ", err)
		}

		installationScript := string(content)

		// read students keys file
		StudentsKeys, err := pkg.GetKeys(ctx, studentsKeysFilePath)
		if err != nil {
			fmt.Println("Unable to read keys from " + studentsKeysFilePath)
			fmt.Println(err)
		}

		// read teachers keys file
		TeachersKeys, err := pkg.GetKeys(ctx, teachersKeysFilePath)
		if err != nil {
			fmt.Println("Unable to read keys from " + teachersKeysFilePath)
			fmt.Println(err)
		}

		insysVpc, err := digitalocean.NewVpc(ctx, "insysVpc", &digitalocean.VpcArgs{
			Region: pulumi.String(region),
		})
		if err != nil {
			fmt.Println("Unable to create VPC")
			fmt.Println(err)
			return err
		}

		// generate regular droplet args
		dropletArgs := digitalocean.DropletArgs{
			Image:    pulumi.String("ubuntu-20-04-x64"),
			Region:   pulumi.String(region),
			Size:     pulumi.String("s-1vcpu-1gb"),
			UserData: pulumi.String(installationScript),
			VpcUuid:  insysVpc.ID(),
		}

		// create Droplet for each students key
		for _, key := range StudentsKeys {
			var SshKeysArray pulumi.StringArray
			SshKeysArray = append(SshKeysArray, key.DOKey.Fingerprint)
			dropletArgs.SshKeys = SshKeysArray
			for _, teachersKey := range TeachersKeys {
				SshKeysArray = append(SshKeysArray, teachersKey.DOKey.Fingerprint)
			}
			_, err := digitalocean.NewDroplet(ctx, key.GetUsername(), &dropletArgs)
			if err != nil {
				fmt.Println("Unable to create droplet")
				fmt.Println(err)
			}
		}

		// create one droplet for teachers
		var SshKeysArray pulumi.StringArray
		for _, teachersKey := range TeachersKeys {
			SshKeysArray = append(SshKeysArray, teachersKey.DOKey.Fingerprint)
		}
		dropletArgs.SshKeys = SshKeysArray
		_, err = digitalocean.NewDroplet(ctx, "insys-lead", &dropletArgs)
		if err != nil {
			fmt.Println("Unable to create droplet")
			fmt.Println(err)
		}
		return nil
	})
}
