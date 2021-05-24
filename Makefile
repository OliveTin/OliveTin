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
#	protoc --go-grpc_out=grpc:gen/grpc/ OliveTin.proto 
#	protoc -I.:$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/ --grpc-gateway_out=gen/grpc --grpc-gateway_opt paths=source_relative OliveTin.proto

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

release-common: 
	rm -rf webui/node_modules/
	rm -rf releases/
	mkdir -p releases/common/
	cp -r webui releases/common/
	cp -r var/config.yaml releases/common/
	cp OliveTin.service releases/common/
	cp README.md releases/common/
	cp Dockerfile releases/common/

release-bin-rpi: daemon-compile-armhf release-common
	mkdir -p releases/rpi/
	cd releases/common && cp -r * ../rpi/
	cp OliveTin.armhf releases/rpi/OliveTin
	cd releases/rpi && tar cavf "../OliveTin-armhf-`date +'%Y-%m-%d'`.`git rev-parse --short HEAD`.tgz" .

release-bin-x64-lin: daemon-compile-x64-lin release-common
	mkdir -p releases/x64-lin/
	cd releases/common && cp -r * ../x64-lin/
	cp OliveTin releases/x64-lin/
	cd releases/x64-lin && tar cavf "../OliveTin-x64-linux-`date +'%Y-%m-%d'`.`git rev-parse --short HEAD`.tgz" .

release-bin-x64-win: daemon-compile-x64-win release-common
	mkdir -p releases/x64-win/
	cd releases/common && cp -r * ../x64-win/
	cp OliveTin.exe releases/x64-win/
	cd releases/x64-lin && zip -r "../OliveTin-x64-windows-`date +'%Y-%m-%d'`.`git rev-parse --short HEAD`.zip" .

releases: release-bin-rpi release-bin-x64-lin release-bin-x64-win


.PHONY: grpc release-common
