package pkg

import (
	"bufio"
	"fmt"
	"github.com/pulumi/pulumi-digitalocean/sdk/v3/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"os"
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
