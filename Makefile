default: 
	go build -o OliveTin github.com/jamesread/OliveTin/cmd/OliveTin

grpc:
	protoc -I.:/usr/share/gocode/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/ --go_out=plugins=grpc:gen/grpc/ OliveTin.proto 
	protoc -I.:/usr/share/gocode/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/ --grpc-gateway_out=gen/grpc --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true OliveTin.proto

.PHONY: grpc
