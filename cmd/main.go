package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kayrus/ec2auth/pkg"
)

func main() {
	var ec2creds pkg.EC2Creds
	var debug bool
	flag.StringVar(&ec2creds.AuthURL, "auth-url", "", "Keystone auth URL")
	flag.StringVar(&ec2creds.Access, "access", "", "EC2 access")
	flag.StringVar(&ec2creds.Secret, "secret", "", "EC2 secret")
	flag.BoolVar(&debug, "debug", false, "show debug logs")
	flag.Parse()

	if ec2creds.AuthURL == "" {
		ec2creds.AuthURL = os.Getenv("OS_AUTH_URL")
	}

	if ec2creds.Access == "" {
		ec2creds.Access = os.Getenv("AWS_ACCESS_KEY_ID")
	}

	if ec2creds.Secret == "" {
		ec2creds.Secret = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	var errors []error
	if ec2creds.AuthURL == "" {
		errors = append(errors, fmt.Errorf("Please define --auth-url parameter or OS_AUTH_URL environment variable"))
	}

	if ec2creds.Access == "" {
		errors = append(errors, fmt.Errorf("Please define --access parameter or AWS_ACCESS_KEY_ID environment variable"))
	}

	if ec2creds.Secret == "" {
		errors = append(errors, fmt.Errorf("Please define --secret parameter or AWS_SECRET_ACCESS_KEY environment variable"))
	}

	if errors != nil {
		for _, e := range errors {
			log.Printf("%s", e)
		}
		os.Exit(1)
	}

	err := pkg.OpenStackEC2Auth(ec2creds, debug)
	if err != nil {
		log.Fatal(err)
	}
}
