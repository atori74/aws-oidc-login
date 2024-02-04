package credential

import (
	"context"

	"gopkg.in/ini.v1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Credential struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func AssumeRoleWithWebIdentity(roleArn, roleSessionName, idToken string) (*Credential, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	stsSvc := sts.NewFromConfig(cfg)

	result, err := stsSvc.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleArn),
		RoleSessionName:  aws.String(roleSessionName),
		WebIdentityToken: aws.String(idToken),
	})
	if err != nil {
		return nil, err
	}

	return &Credential{
		AccessKeyID:     *result.Credentials.AccessKeyId,
		SecretAccessKey: *result.Credentials.SecretAccessKey,
		SessionToken:    *result.Credentials.SessionToken,
	}, nil
}

func (c Credential) SetCredentialFile(path string) error {
	config, err := ini.Load(path)
	if err != nil {
		return err
	}

	config.Section("oidc").Key("aws_access_key_id").SetValue(c.AccessKeyID)
	config.Section("oidc").Key("aws_secret_access_key").SetValue(c.SecretAccessKey)
	config.Section("oidc").Key("aws_session_token").SetValue(c.SessionToken)
	config.SaveTo(path)

	return nil
}
