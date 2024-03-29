package main

import (
	"context"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
)

func fetchSecretsFromVault() {
	if len(mongo_user) == 0 && len(mongo_password) == 0 {
		config := vault.DefaultConfig() // modify for more granular configuration
		config.Address = vault_addr
		client, err := vault.NewClient(config)
		if err != nil {
			panic(fmt.Errorf("unable to initialize Vault client: %w", err))
		}

		k8sAuth, err := auth.NewKubernetesAuth(
			"mongodb",
			auth.WithServiceAccountTokenPath(jwt_path),
		)
		if err != nil {
			panic(fmt.Errorf("unable to initialize Kubernetes auth method: %w", err))
		}

		authInfo, err := client.Auth().Login(ctx, k8sAuth)
		if err != nil {
			panic(fmt.Errorf("unable to log in with Kubernetes auth: %w", err))
		}
		if authInfo == nil {
			panic(fmt.Errorf("no auth info was returned after login"))
		}

		// get secret from Vault, from the default mount path for KV v2 in dev mode, "secret"
		secret, err := client.KVv2("secret").Get(context.Background(), "mongodb/config")
		if err != nil {
			panic(fmt.Errorf("unable to read secret: %w", err))
		}

		// data map can contain more than one key-value pair,
		// in this case we're just grabbing one of them
		username, ok := secret.Data["username"].(string)
		if !ok {
			panic(fmt.Sprintf("value type assertion failed: %T %#v", secret.Data["username"], secret.Data["username"]))
		}
		password, ok := secret.Data["password"].(string)
		if !ok {
			panic(fmt.Sprintf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"]))
		}

		mongo_user = strings.Trim(username, " ")
		mongo_password = strings.Trim(password, " ")
	}
}
