package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/ec2tokens"
	"github.com/kayrus/ec2auth/pkg"
)

func main() {
	ao := &ec2tokens.AuthOptions{}
	var authURL string
	var debug bool
	flag.StringVar(&authURL, "auth-url", "", "Keystone auth URL")
	flag.StringVar(&ao.Access, "access", "", "EC2 access")
	flag.StringVar(&ao.Secret, "secret", "", "EC2 secret")
	flag.BoolVar(&debug, "debug", false, "show debug logs")
	flag.Parse()

	if authURL == "" {
		authURL = os.Getenv("OS_AUTH_URL")
	}

	if ao.Access == "" {
		ao.Access = os.Getenv("AWS_ACCESS_KEY_ID")
	}

	if ao.Secret == "" {
		ao.Secret = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	var errors []error
	if authURL == "" {
		errors = append(errors, fmt.Errorf("Please define --auth-url parameter or OS_AUTH_URL environment variable"))
	}

	if ao.Access == "" {
		errors = append(errors, fmt.Errorf("Please define --access parameter or AWS_ACCESS_KEY_ID environment variable"))
	}

	if ao.Secret == "" {
		errors = append(errors, fmt.Errorf("Please define --secret parameter or AWS_SECRET_ACCESS_KEY environment variable"))
	}

	if errors != nil {
		for _, e := range errors {
			log.Printf("%s", e)
		}
		os.Exit(1)
	}

	provider, err := openstack.NewClient(authURL)
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		provider.HTTPClient = http.Client{
			Transport: &pkg.RoundTripper{
				Rt:     &http.Transport{},
				Logger: &pkg.Logger{},
			},
		}
	}

	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		log.Fatal(err)
	}

	res, err := pkg.OpenStackEC2Auth(identityClient, ao)
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		log.Printf("User: %s", res.Username)
		log.Printf("Project: %s", res.Project)
	}

	fmt.Println(res.TokenID)
}
