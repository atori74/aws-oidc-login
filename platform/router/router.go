package router

import (
	"encoding/gob"
	"html/template"
	"io/ioutil"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/atori74/aws-oidc-login/platform/authenticator"
	"github.com/atori74/aws-oidc-login/web/app/callback"
	"github.com/atori74/aws-oidc-login/web/app/login"
)

// New registers the routes and returns the router.
func New(auth *authenticator.Authenticator, done chan interface{}) (*gin.Engine, error) {
	gin.DefaultWriter = ioutil.Discard
	router := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	//router.Static("/public", "web/static")
	//router.LoadHTMLGlob("web/template/*")

	t, err := loadTemplate("authenticated.html", callback.Template())
	if err != nil {
		return nil, err
	}
	router.SetHTMLTemplate(t)

	router.GET("/login", login.Handler(auth))
	router.GET("/callback", callback.Handler(auth, done))

	return router, nil
}

func loadTemplate(name, content string) (*template.Template, error) {
	t, err := template.New(name).Parse(content)
	if err != nil {
		return nil, err
	}
	return t, nil
}
