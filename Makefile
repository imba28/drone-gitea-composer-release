.PHONY: clean test image

gitea-composer-release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gitea-composer-release -tags netgo -ldflags '-w' main.go

clean:
	rm -f gitea-composer-release

test:
	go test ./... -v

image:
	docker build -t imba28/drone-gitea-composer-release .