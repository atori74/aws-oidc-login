package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/skratchdot/open-golang/open"

	"github.com/atori74/aws-oidc-login/platform/authenticator"
	"github.com/atori74/aws-oidc-login/platform/router"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	done := make(chan interface{})
	rtr := router.New(auth, done)

	go http.ListenAndServe("0.0.0.0:3000", rtr)

	open.Run("http://localhost:3000/login")
	<-done
}
