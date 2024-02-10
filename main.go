package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/skratchdot/open-golang/open"
	flag "github.com/spf13/pflag"

	"github.com/atori74/aws-oidc-login/platform/authenticator"
	"github.com/atori74/aws-oidc-login/platform/router"
)

var (
	envDir      = flag.StringP("envdir", "d", "", "directory where env file exists")
	envFilename = ".env"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: aws-oidc-login [flags] [env]\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.Arg(0) != "" {
		envFilename = flag.Arg(0) + ".env"
	}

	if err := godotenv.Load(filepath.Join(*envDir, envFilename)); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	done := make(chan interface{})
	rtr, err := router.New(auth, done)
	if err != nil {
		log.Fatalf("Failed to initialize router: %v", err)
	}

	go http.ListenAndServe("0.0.0.0:3000", rtr)

	open.Run("http://localhost:3000/login")
	<-done
}
