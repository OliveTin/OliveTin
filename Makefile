compile: daemon-compile

daemon-compile: 
	go build -o OliveTin github.com/jamesread/OliveTin/cmd/OliveTin

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
	protoc -I.:/usr/share/gocode/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/ --go_out=plugins=grpc:gen/grpc/ OliveTin.proto 
	protoc -I.:/usr/share/gocode/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/ --grpc-gateway_out=gen/grpc --grpc-gateway_opt paths=source_relative OliveTin.proto

podman-image:
	buildah bud -t olivetin

podman-container:
	podman kill olivetin
	podman rm olivetin
	podman create --name olivetin -p 1337:1337 -p 1338:1338 -p 1339:1339 -v /etc/OliveTin/:/config:ro olivetin
	podman start olivetin

devrun: compile
	killall OliveTin || true
	./OliveTin &

devcontainer: compile podman-image podman-container

webui-codestyle:
	cd webui && eslint main.js js/*
	cd webui && stylelint style.css

.PHONY: grpc
