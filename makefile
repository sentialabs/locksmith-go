VERSION=0.0.1
GOPATH=$(CURDIR)

default: locksmith/main.go
	# darwin
	GOOS=darwin GOARCH=amd64 go build -o bin/locksmith-darwin-amd64-$(VERSION) ./locksmith

	# windows
	GOOS=windows GOARCH=386 go build -o bin/locksmith-win-x86-$(VERSION).exe ./locksmith
	GOOS=windows GOARCH=amd64 go build -o bin/locksmith-win-amd64-$(VERSION).exe ./locksmith

	# linux
	GOOS=linux GOARCH=amd64 go build -o bin/locksmith-linux-amd64-$(VERSION).exe ./locksmith

	# bsd
	GOOS=freebsd GOARCH=amd64 go build -o bin/locksmith-freebsd-amd64-$(VERSION).exe ./locksmith
