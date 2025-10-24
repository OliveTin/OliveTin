FROM fedora
RUN dnf install -v go -y --exclude=iputils
RUN pwd
RUN ls -l
COPY go.mod /workspace/
COPY go.sum /workspace/
WORKDIR /workspace
RUN go install -v "github.com/bufbuild/buf/cmd/buf"
RUN go install -v "github.com/fzipp/gocyclo/cmd/gocyclo"
RUN go install -v "github.com/go-critic/go-critic/cmd/gocritic"
RUN go install -v "google.golang.org/protobuf/cmd/protoc-gen-go"
