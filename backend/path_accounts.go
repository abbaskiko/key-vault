package backend

import (
	"context"
	"encoding/hex"

	vault "github.com/bloxapp/eth2-key-manager"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/pkg/errors"

	"github.com/bloxapp/key-vault/backend/store"
)

// Endpoints patterns
const (
	// AccountsPattern is the path pattern for list accounts endpoint
	AccountsPattern = "accounts/"
)

func accountsPaths(b *backend) []*framework.Path {
	return []*framework.Path{
		&framework.Path{
			Pattern:         AccountsPattern,
			HelpSynopsis:    "List wallet accounts",
			HelpDescription: ``,
			Fields:          map[string]*framework.FieldSchema{},
			ExistenceCheck:  b.pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: b.pathWalletAccountsList,
			},
		},
	}
}

func (b *backend) pathWalletAccountsList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Load config
	config, err := b.configured(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}

	storage := store.NewHashicorpVaultStore(ctx, req.Storage, config.Network)
	options := vault.KeyVaultOptions{}
	options.SetStorage(storage)

	portfolio, err := vault.OpenKeyVault(&options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open key vault")
	}

	wallet, err := portfolio.Wallet()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve wallet by name")
	}

	var accounts []map[string]string
	for _, a := range wallet.Accounts() {
		accObj := map[string]string{
			"id":               a.ID().String(),
			"name":             a.Name(),
			"validationPubKey": hex.EncodeToString(a.ValidatorPublicKey().Marshal()),
			"withdrawalPubKey": hex.EncodeToString(a.WithdrawalPublicKey().Marshal()),
		}
		accounts = append(accounts, accObj)
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"accounts": accounts,
		},
	}, nil
}
