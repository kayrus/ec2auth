package pkg

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/ec2tokens"
)

type EC2Creds struct {
	Access  string
	Secret  string
	AuthURL string
}

func OpenStackEC2Auth(ec2creds EC2Creds, debug bool) error {
	provider, err := openstack.NewClient(ec2creds.AuthURL)
	if err != nil {
		return err
	}

	if debug {
		provider.HTTPClient = http.Client{
			Transport: &RoundTripper{
				Rt:     &http.Transport{},
				Logger: &logger{},
			},
		}
	}

	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return err
	}

	// auth against openstack
	authOptions := &ec2tokens.AuthOptions{
		Access: ec2creds.Access,
		Secret: ec2creds.Secret,
	}

	res := ec2tokens.Create(identityClient, authOptions)
	if res.Err != nil {
		return res.Err
	}

	user, err := res.ExtractUser()
	if err != nil {
		return err
	}

	project, err := res.ExtractProject()
	if err != nil {
		return err
	}

	tokenID, err := res.ExtractTokenID()
	if err != nil {
		return err
	}

	if debug {
		log.Printf("User: %s", user.Name)
		log.Printf("Project: %s", project.Name)
	}

	fmt.Println(tokenID)

	return nil
}
