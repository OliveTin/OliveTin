FROM fedora
RUN dnf install go -y --exclude=iputils
RUN echo "Docker is stupid"
COPY go.mod ./
COPY go.sum ./
RUN go install "github.com/bufbuild/buf/cmd/buf"
RUN go install "github.com/fzipp/gocyclo/cmd/gocyclo"
RUN go install "github.com/go-critic/go-critic/cmd/gocritic"
RUN go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
RUN go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
RUN go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
RUN go install "google.golang.org/protobuf/cmd/protoc-gen-go"
