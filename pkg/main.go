package pkg

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/ec2tokens"
)

type AuthResult struct {
	Username string
	Project  string
	TokenID  string
}

func OpenStackEC2Auth(identityClient *gophercloud.ServiceClient, ao *ec2tokens.AuthOptions) (*AuthResult, error) {
	res := ec2tokens.Create(identityClient, ao)
	if res.Err != nil {
		return nil, res.Err
	}

	user, err := res.ExtractUser()
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("empty user")
	}

	project, err := res.ExtractProject()
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, fmt.Errorf("empty project scope")
	}

	tokenID, err := res.ExtractTokenID()
	if err != nil {
		return nil, err
	}

	return &AuthResult{user.Name, project.Name, tokenID}, nil
}
