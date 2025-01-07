package tokens

import (
	"context"
	"errors"

	"github.com/gophercloud/gophercloud/v2"
	os "github.com/gophercloud/gophercloud/v2/openstack/identity/v2/tokens"
)

var (
	ErrUsernameRequired = errors.New("You must supply a Username in your AuthOptions.")
	// ErrPasswordProvided is returned if both a password and an API key are provided to Create.
	ErrPasswordProvided = errors.New("Please provide either a password or an API key.")
)

// AuthOptions wraps the OpenStack AuthOptions struct to be able to customize the request body
// when API key authentication is used.
type AuthOptions struct {
	gophercloud.AuthOptions
}

func WrapOptions(opts gophercloud.AuthOptions) AuthOptions {
	return AuthOptions{AuthOptions: opts}
}

// ToTokenCreateMap serializes an AuthOptions into a request body. If an API key is provided, it
// will be used, otherwise
func (auth AuthOptions) ToTokenV2CreateMap() (map[string]interface{}, error) {
	// Verify that other required attributes are present.
	if auth.Username == "" {
		return nil, ErrUsernameRequired
	}

	authMap := make(map[string]interface{})

	authMap["RAX-KSKEY:apiKeyCredentials"] = map[string]interface{}{
		"username": auth.Username,
		"apiKey":   auth.Password,
	}

	if auth.TenantID != "" {
		authMap["tenantId"] = auth.TenantID
	}
	if auth.TenantName != "" {
		authMap["tenantName"] = auth.TenantName
	}

	return map[string]interface{}{"auth": authMap}, nil
}

func (auth AuthOptions) CanReauth() bool {
	return auth.AllowReauth
}

// Create authenticates to Rackspace's identity service and attempts to acquire a Token. Rather
// than interact with this service directly, users should generally call
// rackspace.AuthenticatedClient().
func Create(ctx context.Context, client *gophercloud.ServiceClient, auth AuthOptions) os.CreateResult {
	return os.Create(ctx, client, auth)
}
