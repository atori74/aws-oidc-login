package credential

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/ini.v1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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

func AssumeRoleWithMfa(profile, roleArn, mfaSerial, mfaToken string) (*Credential, error) {
	ctx := context.Background()
	config, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, err
	}
	stsSvc := sts.NewFromConfig(config)

	cred, err := config.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	result, err := stsSvc.AssumeRole(context.Background(), &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(cred.AccessKeyID),
		SerialNumber:    aws.String(mfaSerial),
		TokenCode:       aws.String(mfaToken),
		DurationSeconds: aws.Int32(int32(SessionDuration)),
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

func (c Credential) GetSiginToken() (string, error) {
	session, err := json.Marshal(map[string]string{
		"sessionId":    c.AccessKeyID,
		"sessionKey":   c.SecretAccessKey,
		"sessionToken": c.SessionToken,
	})
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to create json session: %s", err))
	}

	u, _ := url.Parse("https://signin.aws.amazon.com/federation")
	params := url.Values{}
	params.Add("Action", "getSigninToken")
	params.Add("Session", string(session))
	u.RawQuery = params.Encode()

	res, err := http.Get(u.String())
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to get signin token: %s", err))
	} else if res.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Failed to get signin token. Status code is %d", res.StatusCode))
	}
	defer res.Body.Close()

	var dec map[string]string
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&dec)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to decode signin token: %s", err))
	}
	return dec["SigninToken"], nil
}

func GetFederatedSigninURL(signinToken string) string {
	u, _ := url.Parse("https://signin.aws.amazon.com/federation")
	params := url.Values{}
	params.Add("Action", "login")
	params.Add("Issuer", "")
	params.Add("Destination", "https://console.aws.amazon.com/console/home?region=ap-northeast-1")
	params.Add("SigninToken", signinToken)
	u.RawQuery = params.Encode()
	return u.String()
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

func (c AWSProcessCredential) Cache() error {
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "oidc"
	}
	dir := filepath.Join(filepath.Dir(config.DefaultSharedCredentialsFilename()), "cache")
	os.MkdirAll(dir, os.ModePerm)
	f, err := os.Create(filepath.Join(dir, profile))
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	err = encoder.Encode(&c)
	return err
}

func GetCache() (string, error) {
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "oidc"
	}
	path := filepath.Join(filepath.Dir(config.DefaultSharedCredentialsFilename()), "cache", profile)
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	plainJsonBytes, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	var cred AWSProcessCredential
	err = json.Unmarshal(plainJsonBytes, &cred)
	if err != nil {
		return "", err
	}
	exp, err := time.Parse(time.RFC3339, cred.Expiration)
	if err != nil {
		return "", err
	}
	if exp.Before(time.Now()) {
		return "", errors.New("Cached credential is expired.")
	}
	return string(plainJsonBytes), nil
}

func GetAWSCredentialWithMfa() (string, error) {
	mfaUserProfile := os.Getenv("AWS_MFA_USER_PROFILE")
	roleArn := os.Getenv("AWS_ROLE_ARN")
	mfaSerial := os.Getenv("AWS_MFA_SERIAL")
	if du, err := strconv.Atoi(os.Getenv("SESSION_DURATION")); err == nil {
		SessionDuration = du
	}

	buf := bufio.NewReader(os.Stdin)
	input, _, err := buf.ReadLine()
	if err != nil {
		return "", err
	}
	mfaToken := string(input)

	credential, err := AssumeRoleWithMfa(mfaUserProfile, roleArn, mfaSerial, mfaToken)
	if err != nil {
		return "", err
	}
	awsCred := NewAWSProcessCredential(credential)

	result, err := json.Marshal(awsCred)
	if err != nil {
		return "", err
	}

	// Cache credential for next execution.
	err = awsCred.Cache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to cache credentials. %s\n", err)
	}

	return string(result), nil
}
