package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
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

		if opts.IsMfa {
			cred, err = credential.GetAWSCredentialWithMfa()
			if err != nil {
				log.Fatalf("Failed to get the cached credential: %v", err)
			}
			fmt.Printf("%s", cred)
			return
		}
	}

	if opts.IsMfa {
		cred, err := credential.GetCredentialWithMfa()
		if err != nil {
			log.Fatalf("Failed to get temporary credential with MFA: %v", err)
		}
		credentialFilePath := os.Getenv("AWS_CREDENTIALS_FILE")
		if credentialFilePath == "" {
			credentialFilePath = config.DefaultSharedCredentialsFilename()
		}
		err = cred.SetCredentialFile(credentialFilePath)
		if err != nil {
			log.Fatalf("Failed to set credential file. %v", err)
		}

		fmt.Println("Successfully Authenticated.")
		fmt.Println("You can also set credentials as environment variables like below.")
		fmt.Println("==========")
		fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", cred.AccessKeyID)
		fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", cred.SecretAccessKey)
		fmt.Printf("export AWS_SESSION_TOKEN=%s\n", cred.SessionToken)
		return
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
