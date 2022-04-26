compile: daemon-compile-x64-lin

daemon-compile-armhf: 
	GOARCH=arm GOARM=6 go build -o OliveTin.armhf github.com/OliveTin/OliveTin/cmd/OliveTin

daemon-compile-x64-lin: 
	GOOS=linux go build -o OliveTin github.com/OliveTin/OliveTin/cmd/OliveTin

daemon-compile-x64-win:
	GOOS=windows GOARCH=amd64 go build -o OliveTin.exe github.com/OliveTin/OliveTin/cmd/OliveTin

daemon-compile: daemon-compile-armhf daemon-compile-x64-lin daemon-compile-x64-win

daemon-codestyle:
	go fmt ./...
	go vet ./...
	gocyclo -over 4 cmd internal 
	gocritic check ./...

daemon-unittests:
	mkdir -p reports
	go test ./... -coverprofile reports/unittests.out
	go tool cover -html=reports/unittests.out -o reports/unittests.html

githooks:
	cp -v .githooks/* .git/hooks/
	
go-tools:
	go install "github.com/bufbuild/buf/cmd/buf"
	go install "github.com/fzipp/gocyclo/cmd/gocyclo"
	go install "github.com/go-critic/go-critic/cmd/gocritic"
	go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	go install "google.golang.org/protobuf/cmd/protoc-gen-go"

grpc: githooks go-tools
	buf generate

podman-image:
	buildah bud -t olivetin

podman-container:
	podman kill olivetin
	podman rm olivetin
	podman create --name olivetin -p 1337:1337 -v /etc/OliveTin/:/config:ro olivetin
	podman start olivetin

integration-tests-docker-image:
	docker rm -f olivetin && docker rmi -f olivetin
	docker build -t olivetin:latest .
	docker create --name olivetin -p 1337:1337 -v `pwd`/integration-tests/configs/:/config/ olivetin 

devrun: compile
	killall OliveTin || true
	./OliveTin &

devcontainer: compile podman-image podman-container

webui-codestyle:
	cd webui && npm install
	cd webui && ./node_modules/.bin/eslint main.js js/*
	cd webui && ./node_modules/.bin/stylelint style.css

clean:
	rm -rf dist OliveTin OliveTin.armhf OliveTin.exe reports gen

.PHONY: grpc 
