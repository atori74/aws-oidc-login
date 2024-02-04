package callback

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/atori74/aws-oidc-login/platform/authenticator"
	"github.com/atori74/aws-oidc-login/platform/credential"
)

// Handler for our callback.
func Handler(auth *authenticator.Authenticator, done chan interface{}) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		if ctx.Query("state") != session.Get("state") {
			ctx.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

		// Exchange an authorization code for a token.
		token, err := auth.Exchange(ctx.Request.Context(), ctx.Query("code"))
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Failed to convert an authorization code into a token.")
			return
		}

		idToken, err := auth.VerifyIDToken(ctx.Request.Context(), token)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		roleSessionName, ok := profile["name"].(string)
		if !ok {
			ctx.String(http.StatusInternalServerError, "Failed to get name claim.")
			return
		}

		cred, err := credential.AssumeRoleWithWebIdentity(
			os.Getenv("AWS_ROLE_ARN"),
			roleSessionName,
			token.Extra("id_token").(string),
		)
		if err != nil {
			ctx.String(http.StatusUnauthorized, fmt.Sprintf("Failed to assumeRoleWithWebIdentity. %s", err.Error()))
			return
		}

		credentialFilePath := os.Getenv("AWS_CREDENTIALS_FILE")
		if credentialFilePath == "" {
			credentialFilePath = config.DefaultSharedCredentialsFilename()
		}
		err = cred.SetCredentialFile(credentialFilePath)
		if err != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to set credential file. %s", err.Error()))
			return
		}

		fmt.Println("Successfully Authenticated.")
		fmt.Println("You can also set credentials as environment variables like below.")
		fmt.Println("==========")
		fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", cred.AccessKeyID)
		fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", cred.SecretAccessKey)
		fmt.Printf("export AWS_SESSION_TOKEN=%s\n", cred.SessionToken)

		ctx.HTML(http.StatusOK, "authenticated.html", profile)
		close(done)
	}
}

func Template() string {
	return `
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<script src="http://code.jquery.com/jquery-3.1.0.min.js" type="text/javascript"></script>

	<!-- font awesome from BootstrapCDN -->
	<link href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" rel="stylesheet">
	<link href="//maxcdn.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" rel="stylesheet">
	<style>
		body {
		  font-family: "proxima-nova", sans-serif;
		  text-align: center;
		  font-size: 300%;
		  font-weight: 100;
		}
		h1,
		h2,
		h3 {
		  font-weight: 100;
		}
		h2 {
		  margin-top: 30px;
		  margin-bottom: 40px;
		  font-size: 200%;
		}
	</style>
</head>
<body class="home">
	<div class="container">
		<div class="login-page clearfix">
			<div class="logged-in-box auth0-box logged-in">
				<h2>Authenticated</h2>
				<h2>Hello {{ .nickname }}</h3>
			</div>
		</div>
	</div>
</body>
</html>
`
}
