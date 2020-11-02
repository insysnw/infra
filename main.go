package main

import (
	"bufio"
	"fmt"
	"github.com/pulumi/pulumi-digitalocean/sdk/v3/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"os"
	"strconv"
	"strings"
)

func main() {
	installation := "\n#!/bin/bash\nsudo apt update\nsudo apt install --assume-yes ngin postgresql"
	pulumi.Run(func(ctx *pulumi.Context) error {

		// read students keys file
		fStudents, err := os.Open("students.keys")

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
		fTeachers, err := os.Open("teachers.keys")

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

		// create Droplet for each key
		for index, key := range StudentsKeys {
			var SshKeysArray pulumi.StringArray
			SshKeysArray = append(SshKeysArray, key.Fingerprint)
			for _, teachersKey := range TeachersKeys {
				SshKeysArray = append(SshKeysArray, teachersKey.Fingerprint)
			}
			_, err := digitalocean.NewDroplet(ctx, "insys"+strconv.Itoa(index), &digitalocean.DropletArgs{
				Image:   pulumi.String("ubuntu-20-04-x64"),
				Region:  pulumi.String("fra1"),
				Size:    pulumi.String("s-1vcpu-1gb"),
				SshKeys: SshKeysArray,
				UserData: pulumi.String(installation),
			})
			if err != nil {
				fmt.Println("Unable to create droplet")
				fmt.Println(err)
			}
		}
		return nil
	})
}
