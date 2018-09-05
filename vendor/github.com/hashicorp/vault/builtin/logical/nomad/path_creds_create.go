package nomad

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathCredsCreate(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the role",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathTokenRead,
		},
	}
}

func (b *backend) pathTokenRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)

	role, err := b.Role(ctx, req.Storage, name)
	if err != nil {
		return nil, errwrap.Wrapf("error retrieving role: {{err}}", err)
	}
	if role == nil {
		return logical.ErrorResponse(fmt.Sprintf("role %q not found", name)), nil
	}

	// Determine if we have a lease configuration
	leaseConfig, err := b.LeaseConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if leaseConfig == nil {
		leaseConfig = &configLease{}
	}

	// Get the nomad client
	c, err := b.client(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	// Generate a name for the token
	tokenName := fmt.Sprintf("vault-%s-%s-%d", name, req.DisplayName, time.Now().UnixNano())

	// Handling nomad maximum token lenght
	// https://github.com/hashicorp/nomad/blob/d9276e22b3b74674996fb548cdb6bc4c70d5b0e4/nomad/structs/structs.go#L115
	if len(tokenName) > 64 {
		tokenName = tokenName[0:63]
	}

	// Create it
	token, _, err := c.ACLTokens().Create(&api.ACLToken{
		Name:     tokenName,
		Type:     role.TokenType,
		Policies: role.Policies,
		Global:   role.Global,
	}, nil)
	if err != nil {
		return nil, err
	}

	// Use the helper to create the secret
	resp := b.Secret(SecretTokenType).Response(map[string]interface{}{
		"secret_id":   token.SecretID,
		"accessor_id": token.AccessorID,
	}, map[string]interface{}{
		"accessor_id": token.AccessorID,
	})
	resp.Secret.TTL = leaseConfig.TTL

	return resp, nil
}
