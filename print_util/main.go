package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/rodaine/table"
	"os"
	"strconv"
	"strings"
)

func main() {
	studentsKeysFilePath := "../students.keys"
	doToken := "thisisakey"
	f, err := os.Open(studentsKeysFilePath)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	defer f.Close()

	var keys []string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		pubKey := scanner.Text()
		keys = append(keys, pubKey)
	}

	client := godo.NewFromToken(doToken)
	ctx := context.TODO()

	// create table to fill
	tbl := table.New("Droplet", "Student", "IPv4")

	droplets, err := DropletList(ctx, client)
	if err != nil {
		fmt.Println("Unable to get droplets list")
	} else {
		for _, droplet := range droplets {
			ipv4, err := droplet.PublicIPv4()
			if err != nil {
				fmt.Println("Failed to get droplets IPv4 address")
			}
			preindex := strings.Split(droplet.Name, "-")
			index, err := strconv.Atoi(strings.TrimPrefix(preindex[0], "insys"))
			if err != nil {
				fmt.Println("Failed to parse droplet index:\n\tGot \"%s\" instead of a number")
			}
			tbl.AddRow(droplet.Name, strings.Split(keys[index], " ")[2], ipv4)

		}
		tbl.Print()
	}

}

func DropletList(ctx context.Context, client *godo.Client) ([]godo.Droplet, error) {
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(ctx, opt)
		if err != nil {
			return nil, err
		}

		// append the current page's droplets to our list
		list = append(list, droplets...)

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}
