package vault

import (
	"context"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

var client *vault.Client

func Get() *vault.Client {
	if client == nil {
		panic("vault client is not initialized")
	}
	return client
}

func Init(address, token string) error {
	var err error
	client, err = vault.New(
		vault.WithAddress(address),
		vault.WithRequestTimeout(10*time.Second),
	)
	if err != nil {
		return err
	}

	if err = client.SetToken(token); err != nil {
		return err
	}
	return nil
}

func SetStr(key string, value string) error {
	_, err := Get().Secrets.KvV2Write(
		context.Background(),
		key,
		schema.KvV2WriteRequest{Data: map[string]any{"value": value}},
		vault.WithMountPath("secret"),
	)
	return err
}

func GetStr(key string) (string, error) {
	secret, err := Get().Secrets.KvV2Read(
		context.Background(),
		key,
		vault.WithMountPath("secret"),
	)
	if err != nil {
		return "", err
	}
	return secret.Data.Data["value"].(string), nil
}

func SetMap(key string, data map[string]any) error {
	_, err := Get().Secrets.KvV2Write(
		context.Background(),
		key,
		schema.KvV2WriteRequest{Data: data},
		vault.WithMountPath("secret"),
	)
	return err
}

func GetMap(key string) (map[string]any, error) {
	secret, err := Get().Secrets.KvV2Read(
		context.Background(),
		key,
		vault.WithMountPath("secret"),
	)
	if err != nil {
		return map[string]any{}, err
	}
	return secret.Data.Data, nil
}

func Del(key string) error {
	ctx := context.Background()
	_, err := Get().Secrets.KvV2Delete(
		ctx, key, vault.WithMountPath("secret"),
	)
	if err != nil {
		return err
	}
	_, err = Get().Secrets.KvV2DeleteMetadataAndAllVersions(
		ctx, key, vault.WithMountPath("secret"),
	)
	return err
}
