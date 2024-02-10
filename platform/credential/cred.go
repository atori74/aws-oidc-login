package credential

import (
	"context"
	"os"
	"time"

	"gopkg.in/ini.v1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	SessionDuration = 3600
)

type Credential struct {
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

func AssumeRoleWithWebIdentity(roleArn, roleSessionName, idToken string) (*Credential, error) {
	stsSvc := sts.New(sts.Options{Region: "ap-northeast-1"})

	result, err := stsSvc.AssumeRoleWithWebIdentity(context.Background(), &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleArn),
		RoleSessionName:  aws.String(roleSessionName),
		WebIdentityToken: aws.String(idToken),
		DurationSeconds:  aws.Int32(int32(SessionDuration)),
	})
	if err != nil {
		return nil, err
	}

	return &Credential{
		AccessKeyID:     *result.Credentials.AccessKeyId,
		SecretAccessKey: *result.Credentials.SecretAccessKey,
		SessionToken:    *result.Credentials.SessionToken,
		Expiration:      time.Now().Add(time.Duration(SessionDuration) * time.Second).Format(time.RFC3339),
	}, nil
}

func (c Credential) SetCredentialFile(path string) error {
	config, err := ini.Load(path)
	if err != nil {
		return err
	}

	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "oidc"
	}

	config.Section(profile).Key("aws_access_key_id").SetValue(c.AccessKeyID)
	config.Section(profile).Key("aws_secret_access_key").SetValue(c.SecretAccessKey)
	config.Section(profile).Key("aws_session_token").SetValue(c.SessionToken)
	config.SaveTo(path)

	return nil
}

type AWSProcessCredential struct {
	Version int `json:"Version"`
	Credential
}

func NewAWSProcessCredential(c *Credential) *AWSProcessCredential {
	return &AWSProcessCredential{
		Version:    1,
		Credential: *c,
	}
}
