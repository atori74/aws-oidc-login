package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/skratchdot/open-golang/open"

	"github.com/atori74/aws-oidc-login/platform/authenticator"
	"github.com/atori74/aws-oidc-login/platform/credential"
	"github.com/atori74/aws-oidc-login/platform/options"
	"github.com/atori74/aws-oidc-login/platform/router"
)

func main() {
	opts := options.Parse()
	err := opts.Validate()
	if err != nil {
		fmt.Println(err.Error())
		options.Usage()
		return
	}

	if err := godotenv.Load(filepath.Join(opts.EnvDir, opts.EnvFilename)); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	if opts.IsCredentialProcess {
		cred, err := credential.GetCache()
		if err == nil {
			fmt.Printf("%s", cred)
			return
		}
	}

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	done := make(chan interface{})
	rtr, err := router.New(auth, opts, done)
	if err != nil {
		log.Fatalf("Failed to initialize router: %v", err)
	}

	go http.ListenAndServe("0.0.0.0:3000", rtr)

	open.Run("http://localhost:3000/login")
	<-done
}
