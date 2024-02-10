package options

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

type Options struct {
	EnvDir              string
	EnvFilename         string
	IsCredentialProcess bool
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: aws-oidc-login [flags] [env]\n")
	flag.PrintDefaults()
}

func Parse() *Options {
	opts := Options{EnvFilename: ".env"}
	flag.StringVarP(&opts.EnvDir, "envdir", "d", "", "directory where env file exists")
	flag.BoolVarP(&opts.IsCredentialProcess, "provider", "p", false, "work as process credential provider")

	flag.Usage = usage
	flag.Parse()
	if flag.Arg(0) != "" {
		opts.EnvFilename = flag.Arg(0) + ".env"
	}

	return &opts
}
