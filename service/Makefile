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

compile-x64-win: windows-resources
	GOOS=windows GOARCH=amd64 go build -o OliveTin.exe

windows-resources:
	$(MAKE) -wC .. windows-resources

compile: compile-armhf compile-x64-lin compile-x64-win

codestyle: go-tools
	go fmt ./...
	go vet ./...
	gocyclo -over 4 internal
	gocritic check ./...

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
	go install "github.com/fzipp/gocyclo/cmd/gocyclo"
	go install "github.com/go-critic/go-critic/cmd/gocritic"

.PHONY: unittests unittests-fast find-flakey-tests find-flakey-tests-inf

go-tools-all:
	go install "github.com/bufbuild/buf/cmd/buf"
	go install "github.com/fzipp/gocyclo/cmd/gocyclo"
	go install "github.com/go-critic/go-critic/cmd/gocritic"
	go install "google.golang.org/protobuf/cmd/protoc-gen-go"
