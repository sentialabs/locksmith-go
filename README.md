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

# Install:

Download one of the binaries, or install directly.

## Install a binary
From the
[releases](https://github.com/sentialabs/locksmith-go/releases)
page, download the latest zip archive for your platform. Extract this file
within a directory that is included in your `$PATH`, and (optionally) rename
the file to `locksmith`.

For Mac OS X, you would most probably like to get the `darwin-amd64` archive.

## Install using `go get`
```
export GOPATH=~/go
export GOBIN=$GOPATH/bin

go get -u github.com/sentialabs/locksmith-go/locksmith
```

Now, add `~/go/bin` to your path. And start `locksmith`!

# Building a Release
```
make depend # exits with an error until I figure out a better way
make
```
