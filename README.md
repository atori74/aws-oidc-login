# aws-oidc-login

OIDCプロバイダーにログインしてIDトークンを取得し、さらにSTSの`AssumeRoleWithWebIdentity`を実行して得られた一時的な認証情報を`$HOME/.aws/credentials`にセットするまでを自動化する。

## 使い方

### 準備

OIDCプロバイダーのCallback URLsとして`http://localhost:3000/callback`を許可しておく

実行ファイルと同じディレクトリにある`.env` ファイルを環境に合わせて編集する
(環境変数としてセットしてもOK)

以下は例
```.env
# The URL of your OIDC Domain.
OIDC_DOMAIN='~.jp.auth0.com'

# OIDC application's Client ID.
OIDC_CLIENT_ID='<Client ID>'

# OIDC application's Client Secret.
OIDC_CLIENT_SECRET='<Client Secret>'

# Scopes to request. (comma separated)
OIDC_CLAIMS='profile,email'

# Federated IAM Role ARN.
AWS_ROLE_ARN='arn:aws:iam::<Account ID>:role/<Role Name>'

# Credential profile.
AWS_PROFILE='oidc'
```

### ログイン

コマンドを実行する

```sh
$ aws-oidc-login
```

ブラウザが開いてOIDCプロバイダーのログイン画面にリダイレクトされるので、ログインする。  
ログインに成功したら画面遷移し"Authenticated"と表示される。

.credentialsファイルに`oidc`というプロファイル名で一時的な認証情報を書き込まれるので、それを使ってAWS CLIを実行可能となる。

```sh
$ aws --profile oidc sts get-caller-identity
{
    "UserId": "<Session ID>:<Session Name>",
    "Account": "<Account ID>",
    "Arn": "arn:aws:sts::<Account ID>:assumed-role/<Role Name>/<Session Name>"
}
```

## 参考
[Auth0 Go SDK Quickstarts: Login](https://auth0.com/docs/quickstart/webapp/golang/01-login)  
https://github.com/auth0-samples/auth0-golang-web-app/tree/master/01-Login
