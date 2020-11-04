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

type SshKey struct {
	KeyType   string
	KeyItself string
	Comment   string
}

func (key SshKey) toString() string {
	return key.KeyType + " " + key.KeyItself + " " + key.Comment
}

func (key SshKey) toDO(ctx *pulumi.Context) (*digitalocean.SshKey, error) {
	name := "insys-key: " + key.Comment
	newKey, err := digitalocean.NewSshKey(ctx, name, &digitalocean.SshKeyArgs{
		Name:      pulumi.String(name),
		PublicKey: pulumi.String(key.toString()),
	})
	if err != nil {
		return newKey, err
	}
	return newKey, nil
}

func GetKeys(ctx *pulumi.Context, keysFilePath string) ([]*digitalocean.SshKey, error) {
	var Keys []*digitalocean.SshKey

	localKeys, err := ReadKeys(keysFilePath)
	if err != nil {
		return Keys, err
	}

	for _, localKey := range localKeys {
		newKey, err := localKey.toDO(ctx)
		if err != nil {
			return Keys, err
		} else {
			Keys = append(Keys, newKey)
		}
	}
	return Keys, nil
}

func ReadKeys(keysFilePath string) ([]SshKey, error) {
	f, err := os.Open(keysFilePath)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)

	var Keys []SshKey

	for scanner.Scan() {
		pubKey := scanner.Text()
		parsedPubKey := strings.Split(pubKey, " ")
		// TODO: add some form of validation
		parsedComment := strings.Join(parsedPubKey[2:], " ")
		Keys = append(Keys, SshKey{
			KeyType:   parsedPubKey[0],
			KeyItself: parsedPubKey[1],
			Comment:   parsedComment,
		})
	}
	return Keys, nil

}

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
		StudentsKeys, err := GetKeys(ctx, studentsKeysFilePath)
		if err != nil {
			fmt.Println("Unable to read keys from " + studentsKeysFilePath)
			fmt.Println(err)
		}

		// read teachers keys file
		TeachersKeys, err := GetKeys(ctx, teachersKeysFilePath)
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
