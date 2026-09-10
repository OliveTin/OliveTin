FROM fedora
RUN dnf install -v go -y --exclude=iputils
RUN pwd
RUN ls -l
COPY go.mod /workspace/
COPY go.sum /workspace/
WORKDIR /workspace
RUN go install -v "github.com/bufbuild/buf/cmd/buf"
RUN go install -v "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2"
RUN go install -v "google.golang.org/protobuf/cmd/protoc-gen-go"
