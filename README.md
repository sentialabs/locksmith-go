# Locksmith in Go

Configure your `~/.aws/credentials` with your AWS and Beagle credentials:

```
[locksmith]
aws_access_key_id = AKIAXXXXXXXXXXXXXXXX
aws_secret_access_key = XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
mfa_serial = XXXXXXXXXXXXXXXXXXXXXXX
beagle_url = https://beagle.sentiampc.com/api/v1/bookmarks
beagle_pass = XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

Install:

```
export GOPATH=~/go
export GOBIN=$GOPATH/bin

go get -u github.com/sentialabs/locksmith-go/locksmith
```

Now, add `~/go/bin` to your path. And start `locksmith`!
