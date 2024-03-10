package options

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
)

type Options struct {
	EnvDir              string
	EnvFilename         string
	IsCredentialProcess bool
	IsConsoleSignin     bool
}

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: aws-oidc-login [flags] [env]\n")
	flag.PrintDefaults()
}

func Parse() *Options {
	opts := Options{EnvFilename: ".env"}
	execPath, _ := os.Executable()
	flag.StringVarP(&opts.EnvDir, "envdir", "d", filepath.Dir(execPath), "directory where env file exists")
	flag.BoolVarP(&opts.IsCredentialProcess, "provider", "p", false, "work as process credential provider")
	flag.BoolVarP(&opts.IsConsoleSignin, "console", "c", false, "signin on AWS Management Console")

	flag.Usage = Usage
	flag.Parse()
	if flag.Arg(0) != "" {
		opts.EnvFilename = flag.Arg(0) + ".env"
	}

	return &opts
}

func (opt Options) Validate() error {
	if opt.IsConsoleSignin && opt.IsCredentialProcess {
		return errors.New("Option error: Do not set both provider and console flags.")
	}
	return nil
}
