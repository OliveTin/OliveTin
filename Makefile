compile: 
	go build -o OliveTin github.com/jamesread/OliveTin/cmd/OliveTin

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

.PHONY: grpc
