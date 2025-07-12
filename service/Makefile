define delete-files
	python -c "import shutil;shutil.rmtree('$(1)', ignore_errors=True)"
endef

compile-currentenv:
	go build

prep:
	go mod download
	go generate ./...

compile-armhf:
	go env -w GOARCH=arm GOARM=6
	go build -o OliveTin.armhf
	go env -u GOARCH GOARM

compile-x64-lin:
	go env -w GOOS=linux
	go build -o OliveTin
	go env -u GOOS

compile-x64-win:
	go env -w GOOS=windows GOARCH=amd64
	go build -o OliveTin.exe
	go env -u GOOS GOARCH

compile: compile-armhf compile-x64-lin compile-x64-win

codestyle: go-tools
	go fmt ./...
	go vet ./...
	gocyclo -over 4 internal
	gocritic check ./...

unittests:
	$(call delete-files,reports)
	mkdir reports
	go test ./... -coverprofile reports/unittests.out
	go tool cover -html=reports/unittests.out -o reports/unittests.html

go-tools:
	go install "github.com/fzipp/gocyclo/cmd/gocyclo"
	go install "github.com/go-critic/go-critic/cmd/gocritic"

go-tools-all:
	go install "github.com/bufbuild/buf/cmd/buf"
	go install "github.com/fzipp/gocyclo/cmd/gocyclo"
	go install "github.com/go-critic/go-critic/cmd/gocritic"
	go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	go install "google.golang.org/protobuf/cmd/protoc-gen-go"
