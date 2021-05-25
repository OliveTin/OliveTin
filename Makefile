compile: daemon-compile-x64-lin

daemon-compile-armhf: 
	GOARCH=arm GOARM=6 go build -o OliveTin.armhf github.com/jamesread/OliveTin/cmd/OliveTin

daemon-compile-x64-lin: 
	GOOS=linux go build -o OliveTin github.com/jamesread/OliveTin/cmd/OliveTin 

daemon-compile-x64-win:
	GOOS=windows GOARCH=amd64 go build -o OliveTin.exe github.com/jamesread/OliveTin/cmd/OliveTin

daemon-compile: daemon-compile-armhf daemon-compile-x64-lin daemon-compile-x64-win

daemon-codestyle:
	go fmt ./...
	go vet ./...
	golint ./...
	gocyclo -over 4 cmd internal 

daemon-unittests:
	mkdir -p reports
	go test ./... -coverprofile reports/unittests.out
	go tool cover -html=reports/unittests.out -o reports/unittests.html

grpc:
	protoc -I.:$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/ --go_out=. --go-grpc_out=. --grpc-gateway_out=. OliveTin.proto 

podman-image:
	buildah bud -t olivetin

podman-container:
	podman kill olivetin
	podman rm olivetin
	podman create --name olivetin -p 1337:1337 -v /etc/OliveTin/:/config:ro olivetin
	podman start olivetin

devrun: compile
	killall OliveTin || true
	./OliveTin &

devcontainer: compile podman-image podman-container

webui-codestyle:
	cd webui && eslint main.js js/*
	cd webui && stylelint style.css

.PHONY: grpc 
