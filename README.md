# aws-oidc-login

OpenIDプロバイダーにログインしてIDトークンを取得し、さらにSTSの`AssumeRoleWithWebIdentity`を実行して得られた一時的な認証情報を`$HOME/.aws/credentials`にセットするまでを自動化する。

## 使い方

### 準備

- OpenIDプロバイダーを用意し(Auth0やCognito, Keycloak等)、Callback URLとして`http://localhost:3000/callback`を許可しておく

- 連携するAWSアカウントでIAM IDプロバイダーをセットアップする。またスイッチ先ロールの信頼関係ポリシーでIDプロバイダーからの`sts:AssumeRoleWithWebIdentity`を許可しておく

- 実行ファイルと同じディレクトリにある`.env` ファイルを環境に合わせて編集する
(直接環境変数としてセットしてもOK)

以下`.env`ファイルの例
```.env
# The URL of your OIDC Domain. (Issuer claim without https://)
OIDC_DOMAIN='~.jp.auth0.com/'

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

### Process Credential Providerを使った方法

`$HOME/.aws/config`に以下のように設定する

```ini
[oidc]
region = ap-northeast-1
output = json
credential_process = path/to/aws-oidc-login -p
```

通常どおりAWS CLIを実行すると、ブラウザが開くので同様にOIDCプロバイダーにログインする。  
ログインに成功するとシームレスにCLI実行が可能となる。  
取得した認証情報は`$HOME/.aws/cache/<Profile name>`に保存され、セッションの有効期限内であれば再利用される。

### オプション

```
$ aws-oidc-login -h
usage: aws-oidc-login [flags] [env]
  -d, --envdir string   directory where env file exists
  -p, --provider        work as process credential provider
```

\-d, --envdir (option) : `.env`ファイルが存在するディレクトリを指定する (デフォルトは実行ファイルと同じディレクトリ)  
\-p, --provider (option) : このフラグをつけた場合、AWS CLIが[credential_process](https://docs.aws.amazon.com/ja_jp/cli/latest/userguide/cli-configure-sourcing-external.html)で受け取れる形式のJSONを出力する  
\[env\] (option) : `.env`ファイルの名前を指定する (デフォルトは`.env`) (例)auth0 -> `auth0.env`を参照しに行く.

## 参考
[Auth0 Go SDK Quickstarts: Login](https://auth0.com/docs/quickstart/webapp/golang/01-login)  
https://github.com/auth0-samples/auth0-golang-web-app/tree/master/01-Login
