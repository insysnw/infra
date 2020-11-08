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
	DOKey     *digitalocean.SshKey
}

func (key SshKey) ToString() string {
	return key.KeyType + " " + key.KeyItself + " " + key.Comment
}

func (key SshKey) GetUsername() string {
	return strings.Split(key.Comment, "@")[0]
}

func (key SshKey) initDO(ctx *pulumi.Context) error {
	name := "insys-key: " + key.Comment
	doKey, err := digitalocean.NewSshKey(ctx, name, &digitalocean.SshKeyArgs{
		Name:      pulumi.String(name),
		PublicKey: pulumi.String(key.toString()),
	})
	if err != nil {
		return err
	}
	key.DOKey = doKey
	return nil
}

func GetKeys(ctx *pulumi.Context, keysFilePath string) ([]SshKey, error) {
	//keys with DO part instantiated
	var keys []SshKey

	localKeys, err := ReadKeys(keysFilePath)
	if err != nil {
		return keys, err
	}

	for _, localKey := range localKeys {
		err := localKey.initDO(ctx)
		if err != nil {
			return keys, err
		} else {
			keys = append(keys, localKey)
		}
	}
	return keys, nil
}

func readKeys(keysFilePath string) ([]SshKey, error) {
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
