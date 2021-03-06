package main

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/rodaine/table"
	"os"
	"strings"
)

func main() {
	doToken := os.Getenv("DO_TOKEN")

	client := godo.NewFromToken(doToken)
	ctx := context.TODO()

	// create table to fill
	tbl := table.New("Droplet", "IPv4")

	droplets, err := DropletList(ctx, client)
	if err != nil {
		fmt.Println("Unable to get droplets list")
		fmt.Println(err)
	} else {
		for _, droplet := range droplets {
			if strings.Contains(droplet.Name, "insys") {
				if strings.Contains(droplet.Name, "internal") {
					ipv4, err := droplet.PrivateIPv4()
					if err != nil {
						fmt.Println("Failed to get droplets IPv4 address")
					}
					tbl.AddRow(droplet.Name, ipv4)
				} else {
					ipv4, err := droplet.PublicIPv4()
					if err != nil {
						fmt.Println("Failed to get droplets IPv4 address")
					}
					tbl.AddRow(droplet.Name, ipv4)
				}
			}
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
