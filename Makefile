define delete-files
	python -c "import shutil;shutil.rmtree('$(1)', ignore_errors=True)"
endef

compile: daemon-compile-currentenv

daemon-compile-currentenv:
	go build github.com/OliveTin/OliveTin/cmd/OliveTin

daemon-compile-armhf:
	go env -w GOARCH=arm GOARM=6
	go build -o OliveTin.armhf github.com/OliveTin/OliveTin/cmd/OliveTin
	go env -u GOARCH GOARM

daemon-compile-x64-lin:
	go env -w GOOS=linux
	go build -o OliveTin github.com/OliveTin/OliveTin/cmd/OliveTin
	go env -u GOOS

daemon-compile-x64-win:
	go env -w GOOS=windows GOARCH=amd64
	go build -o OliveTin.exe github.com/OliveTin/OliveTin/cmd/OliveTin
	go env -u GOOS GOARCH

daemon-compile: daemon-compile-armhf daemon-compile-x64-lin daemon-compile-x64-win

daemon-codestyle:
	go fmt ./...
	go vet ./...
	gocyclo -over 4 cmd internal
	gocritic check ./...

daemon-unittests:
	$(call delete-files,reports)
	mkdir reports
	go test ./... -coverprofile reports/unittests.out
	go tool cover -html=reports/unittests.out -o reports/unittests.html


it:
	cd integration-tests && make

go-tools:
	go install "github.com/bufbuild/buf/cmd/buf"
	go install "github.com/fzipp/gocyclo/cmd/gocyclo"
	go install "github.com/go-critic/go-critic/cmd/gocritic"
	go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	go install "google.golang.org/protobuf/cmd/protoc-gen-go"

grpc: go-tools
	buf generate

dist: protoc

protoc:
	protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. -I .:/usr/include/ OliveTin.proto

podman-image:
	buildah bud -t olivetin

podman-container:
	podman kill olivetin || true
	podman rm olivetin || true
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
	cd webui.dev && npm install
	cd webui.dev && npx eslint main.js js/*
	cd webui.dev && npx stylelint style.css

webui-dist:
	$(call delete-files,webui)
	$(call delete-files,webui.dev/dist)
	cd webui.dev && npm install
	cd webui.dev && npx parcel build --public-url "."
	python -c "import shutil;shutil.move('webui.dev/dist', 'webui')"
	python -c "import shutil;import glob;[shutil.copy(f, 'webui') for f in glob.glob('webui.dev/*.png')]"

clean:
	$(call delete-files,dist)
	$(call delete-files,OliveTin)
	$(call delete-files,OliveTin.armhf)
	$(call delete-files,OliveTin.exe)
	$(call delete-files,reports)
	$(call delete-files,gen)

.PHONY: grpc
