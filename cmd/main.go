package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/ec2tokens"
	"github.com/kayrus/ec2auth/pkg"
)

func main() {
	ao := &ec2tokens.AuthOptions{}
	var authURL string
	var host string
	var debug bool
	var showErr bool
	var insecureTls bool
	var threads uint
	flag.StringVar(&authURL, "auth-url", "", "Keystone auth URL")
	flag.StringVar(&host, "host", "", "override keystone HOST")
	flag.StringVar(&ao.Access, "access", "", "EC2 access")
	flag.StringVar(&ao.Secret, "secret", "", "EC2 secret")
	flag.UintVar(&threads, "threads", 0, "Whether to run an infinite loop with an amount of threads")
	flag.BoolVar(&insecureTls, "insecure-tls", false, "Whether to ignore server TLS certificate verification")
	flag.BoolVar(&debug, "debug", false, "show debug logs")
	flag.BoolVar(&showErr, "show-error", false, "show error type on auth failure")
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

	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureTls,
	}
	var logger pkg.ILogger
	if debug {
		logger = &pkg.Logger{}
	} else {
		logger = &pkg.NoopLogger{}
	}
	provider.HTTPClient = http.Client{
		Transport: &pkg.RoundTripper{
			Rt: &http.Transport{
				TLSClientConfig: tlsConfig,
				Dial: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   9 * time.Second,
				ResponseHeaderTimeout: 9 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Host:   &host,
			Logger: logger,
		},
	}

	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		log.Fatal(err)
	}

	lck := &sync.RWMutex{}
	errs := make(map[string]uint64)
	totalReq := new(uint64)
	totalErr := new(uint64)
	fps := new(uint64)
	ops := new(uint64)
	auth := func(limiter chan struct{}) {
		atomic.AddUint64(ops, 1)
		res, err := pkg.OpenStackEC2Auth(identityClient, ao)
		if err != nil {
			atomic.AddUint64(fps, 1)
			if limiter == nil {
				log.Print(err)
				os.Exit(1)
			}
			if showErr {
				var errType string
				if e, ok := err.(*url.Error); ok && e.Err != nil {
					errType = fmt.Sprintf("%v", e.Err)
				} else {
					errType = fmt.Sprintf("%T", err)
				}
				lck.Lock()
				errs[errType] += 1
				lck.Unlock()
			}
			<-limiter
			return
		}

		if debug {
			log.Printf("User: %s", res.Username)
			log.Printf("Project: %s", res.Project)
		}

		if limiter == nil {
			fmt.Println(res.TokenID)
			return
		}

		<-limiter
	}

	if threads == 0 {
		auth(nil)
		os.Exit(0)
	}

	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				f := atomic.SwapUint64(fps, 0)
				s := atomic.SwapUint64(ops, 0)
				tS := atomic.AddUint64(totalReq, s)
				tF := atomic.AddUint64(totalErr, f)
				var perc uint64
				var tPerc uint64
				if s > 0 {
					perc = 100 * f / s
				}
				if tS > 0 {
					tPerc = 100 * tF / tS
				}
				log.Printf("%d rps, %d failed (%d%%)", s, f, perc)
				log.Printf("total %d rps, %d failed: %d%%", tS, tF, tPerc)
				if showErr {
					lck.RLock()
					for k, v := range errs {
						log.Printf("ERROR: %s -> %d", k, v)
					}
					lck.RUnlock()
				}
			}
		}
	}()

	limiter := make(chan struct{}, threads)
	for {
		limiter <- struct{}{}
		go auth(limiter)
	}
}
