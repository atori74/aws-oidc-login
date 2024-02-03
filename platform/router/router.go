package router

import (
	"encoding/gob"
	"io/ioutil"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/atori74/aws-oidc-login/platform/authenticator"
	"github.com/atori74/aws-oidc-login/web/app/callback"
	"github.com/atori74/aws-oidc-login/web/app/login"
)

// New registers the routes and returns the router.
func New(auth *authenticator.Authenticator, done chan interface{}) *gin.Engine {
	gin.DefaultWriter = ioutil.Discard
	router := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	router.Static("/public", "web/static")
	router.LoadHTMLGlob("web/template/*")

	router.GET("/login", login.Handler(auth))
	router.GET("/callback", callback.Handler(auth, done))

	return router
}
