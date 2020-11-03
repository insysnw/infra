package main

import (
	"bufio"
	"fmt"
	"github.com/pulumi/pulumi-digitalocean/sdk/v3/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
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
		fStudents, err := os.Open(studentsKeysFilePath)

		if err != nil {
			fmt.Println("Error: ", err)
		}

		defer fStudents.Close()

		scanner := bufio.NewScanner(fStudents)

		// create each key
		var StudentsKeys []*digitalocean.SshKey
		for scanner.Scan() {
			pubKey := scanner.Text()
			parsedPubKey := strings.Split(pubKey, " ")
			keyComment := parsedPubKey[2]
			name := "insys-key: " + keyComment
			newKey, err := digitalocean.NewSshKey(ctx, name, &digitalocean.SshKeyArgs{
				Name:      pulumi.String(name),
				PublicKey: pulumi.String(pubKey),
			})
			if err != nil {
				fmt.Println("Unable to add key for " + keyComment)
				fmt.Println(err)
			} else {
				StudentsKeys = append(StudentsKeys, newKey)
			}
		}

		// read teachers keys file
		fTeachers, err := os.Open(teachersKeysFilePath)

		if err != nil {
			fmt.Println("Error: ", err)
		}

		defer fTeachers.Close()

		scanner = bufio.NewScanner(fTeachers)

		if err := scanner.Err(); err != nil {
			fmt.Println("Error: ", err)
		}

		var TeachersKeys []*digitalocean.SshKey
		for scanner.Scan() {
			pubKey := scanner.Text()
			parsedPubKey := strings.Split(pubKey, " ")
			keyComment := parsedPubKey[2]
			name := "insys-key: " + keyComment
			newKey, err := digitalocean.NewSshKey(ctx, name, &digitalocean.SshKeyArgs{
				Name:      pulumi.String(name),
				PublicKey: pulumi.String(pubKey),
			})
			if err != nil {
				fmt.Println("Unable to add key for " + keyComment)
				fmt.Println(err)
			} else {
				TeachersKeys = append(TeachersKeys, newKey)
			}
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
		for index, key := range StudentsKeys {
			var SshKeysArray pulumi.StringArray
			SshKeysArray = append(SshKeysArray, key.Fingerprint)
			dropletArgs.SshKeys = SshKeysArray
			for _, teachersKey := range TeachersKeys {
				SshKeysArray = append(SshKeysArray, teachersKey.Fingerprint)
			}
			_, err := digitalocean.NewDroplet(ctx, "insys"+strconv.Itoa(index), &dropletArgs)
			if err != nil {
				fmt.Println("Unable to create droplet")
				fmt.Println(err)
			}
		}

		// create one droplet for teachers
		var SshKeysArray pulumi.StringArray
		for _, teachersKey := range TeachersKeys {
			SshKeysArray = append(SshKeysArray, teachersKey.Fingerprint)
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
