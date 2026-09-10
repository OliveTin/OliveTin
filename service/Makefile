define delete-files
	python3 -c "import shutil;shutil.rmtree('$(1)', ignore_errors=True)"
endef

BUF_VERSION := v1.72.0
GOLANGCI_LINT_VERSION := v2.13.2
PROTOC_GEN_GO_VERSION := v1.36.12

compile-currentenv:
	go build

prep:
	go mod download
	go generate ./...

compile-armhf:
	GOOS=linux GOARCH=arm GOARM=6 go build -o OliveTin.armhf

compile-x64-lin:
	GOOS=linux GOARCH=amd64 go build -o OliveTin

compile-x64-win: windows-resources
	GOOS=windows GOARCH=amd64 go build -o OliveTin.exe

windows-resources:
	$(MAKE) -wC .. windows-resources

compile: compile-armhf compile-x64-lin compile-x64-win

codestyle: go-tools
	go fmt ./...
	golangci-lint run ./...

test: unittests

tests: unittests

unittests:
	$(call delete-files,reports)
	mkdir reports
	go test ./... -coverprofile reports/unittests.out
	go tool cover -html=reports/unittests.out -o reports/unittests.html

unittests-fast:
	go test ./... -count=1

find-flakey-tests:
	echo "Running unittests-fast infinitely"
	sh -c "while $(MAKE) unittests-fast; do :; done"

find-flakey-tests-inf:
	go run ./scripts/find-flakey-tests-inf

go-tools:
	go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)"

proto-tools:
	go install "github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)"
	go install "google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)"

go-tools-all: go-tools proto-tools

.PHONY: codestyle go-tools proto-tools go-tools-all unittests unittests-fast find-flakey-tests find-flakey-tests-inf
