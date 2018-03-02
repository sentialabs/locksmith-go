VERSION=0.0.1
GOPATH=$(CURDIR)

default: locksmith/main.go
	# darwin
	GOOS=darwin GOARCH=amd64 go build -o bin/locksmith-darwin-amd64-$(VERSION) ./locksmith
	# windows
	GOOS=windows GOARCH=386 go build -o bin/locksmith-winx86-$(VERSION).exe ./locksmith
	GOOS=windows GOARCH=amd64 go build -o bin/locksmith-winamd64-$(VERSION).exe ./locksmith
