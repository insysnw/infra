package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/pulumi/pulumi-digitalocean/sdk/v3/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"os"
	"strconv"
	"strings"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		f, err := os.Open("authorized_keys")

		if err != nil {
			fmt.Println("Error: ", err)
		}

		defer f.Close()

		scanner := bufio.NewScanner(f)

		pat := "it_should_not_be_here"

		clientx := godo.NewFromToken(pat)
		ctxx := context.TODO()

		opt := &godo.ListOptions{
			Page:    1,
			PerPage: 200,
		}

		keys, _, er := clientx.Keys.List(ctxx, opt)

		if er != nil {
			fmt.Println("Unable to retrieve keys")
			fmt.Println(er)
		} else {
			fmt.Println(keys)
		}

		var keysToUse []*digitalocean.SshKey
		for scanner.Scan() {
			pubKey := scanner.Text()
			for _, key := range keys {
				if key.PublicKey == pubKey {
					pulumiKey, err := digitalocean.GetSshKey(ctx, key.Name, pulumi.ID(key.ID), nil)
					if err != nil {
						fmt.Println("Unable to retrieve keys")
						fmt.Println(er)
					} else {
						keysToUse = append(keysToUse, pulumiKey)
						break
					}
				}
			}
			parsedPubKey := strings.Split(pubKey, " ")
			keyComment := parsedPubKey[2]
			name := "insys: " + keyComment
			newKey, err := digitalocean.NewSshKey(ctx, name, &digitalocean.SshKeyArgs{
				Name:      pulumi.String(name),
				PublicKey: pulumi.String(pubKey),
			})
			if err != nil {
				fmt.Println("Unable to add key for " + keyComment)
				fmt.Println(err)
			} else {
				keysToUse = append(keysToUse, newKey)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error: ", err)
		}

		for index, key := range keysToUse {
			var SshKeysArray pulumi.StringArray
			SshKeysArray = append(SshKeysArray, key.Fingerprint)
			_, err := digitalocean.NewDroplet(ctx, "insys"+strconv.Itoa(index), &digitalocean.DropletArgs{
				Image:   pulumi.String("ubuntu-20-04-x64"),
				Region:  pulumi.String("fra1"),
				Size:    pulumi.String("s-1vcpu-1gb"),
				SshKeys: SshKeysArray,
			})
			if err != nil {
				fmt.Println("Unable to create droplet")
				fmt.Println(err)
			}
		}
		return nil
	})
}
