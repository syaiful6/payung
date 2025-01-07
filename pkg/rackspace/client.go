package rackspace

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	os "github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/utils"
	v2tokens "github.com/syaiful6/payung/pkg/rackspace/identity/v2/tokens"
)

const (
	// RacksapaceUS Identity is an identity endpoint located in the United States.
	RackspaceUSIdentity = "https://identity.api.rackspacecloud.com/v2.0/"

	// RackspaceUKIdentity is an identity endpoint located in the UK.
	RackspaceUKIdentity = "https://lon.identity.api.rackspacecloud.com/v2.0/"
)

const (
	v20 = "v2.0"
)

// NewClient creates a client that's prepared to communicate with the Rackspace API,
// but is not yet authenticated. Users will probably prefer using the AuthenticatedClient function
// instead.
func NewClient(endpoint string) (*gophercloud.ProviderClient, error) {
	if endpoint == "" {
		return os.NewClient(RackspaceUSIdentity)
	}

	return os.NewClient(endpoint)
}

func AuthenticatedClient(ctx context.Context, options gophercloud.AuthOptions) (*gophercloud.ProviderClient, error) {
	client, err := NewClient(options.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	err = Authenticate(ctx, client, options)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func Authenticate(ctx context.Context, client *gophercloud.ProviderClient, options gophercloud.AuthOptions) error {
	versions := []*utils.Version{
		{ID: v20, Priority: 20, Suffix: "/v2.0/"},
	}

	chosen, endpoint, err := utils.ChooseVersion(ctx, client, versions)
	if err != nil {
		return err
	}

	switch chosen.ID {
	case v20:
		return v2auth(ctx, endpoint, client, options)
	default:
		return fmt.Errorf("Unrecognized identity version: %s", chosen.ID)
	}
}

func AuthenticateV2(ctx context.Context, client *gophercloud.ProviderClient, options gophercloud.AuthOptions) error {
	return v2auth(ctx, "", client, options)
}

func v2auth(ctx context.Context, endpoint string, client *gophercloud.ProviderClient, options gophercloud.AuthOptions) error {
	v2Client := NewIdentityV2(client)
	if endpoint != "" {
		v2Client.Endpoint = endpoint
	}

	result := v2tokens.Create(ctx, v2Client, v2tokens.WrapOptions(options))
	token, err := result.ExtractToken()
	if err != nil {
		return err
	}

	catalog, err := result.ExtractServiceCatalog()
	if err != nil {
		return err
	}

	if options.AllowReauth {
		client.ReauthFunc = func(ctx context.Context) error {
			return AuthenticateV2(ctx, client, options)
		}
	}

	client.TokenID = token.ID
	client.EndpointLocator = func(opts gophercloud.EndpointOpts) (string, error) {
		return os.V2EndpointURL(catalog, opts)
	}

	return nil
}

func NewIdentityV2(client *gophercloud.ProviderClient) *gophercloud.ServiceClient {
	v2Endpoint := client.IdentityBase + "v2.0/"

	return &gophercloud.ServiceClient{
		ProviderClient: client,
		Endpoint:       v2Endpoint,
	}
}

func NewObjectStorageV1(client *gophercloud.ProviderClient, eo gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	return os.NewObjectStorageV1(client, eo)
}
