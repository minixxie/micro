FROM minixxie/golang:1.21.0 as golang

ARG APP

WORKDIR /go/src/app

# Prepare 3rd party library files
COPY ./examples/golang_grpc/go.mod .
COPY ./examples/golang_grpc/go.sum .
COPY . /micro
RUN go mod edit -replace github.com/minixxie/micro=/micro
RUN go mod download

# Generate golang code from proto files
COPY ./examples/golang_grpc/pb ./pb
COPY ./examples/golang_grpc/scripts/protoc.sh ./scripts/protoc.sh
RUN ./scripts/protoc.sh
RUN cp `find /go/src/app -name "*swagger.json" | head -n 1` /swagger.json

# Build golang code
ADD ./examples/golang_grpc /go/src/app
WORKDIR /go/src/app
RUN diffFiles=$(gofmt -l `find . -name "*.go" -not -path "./vendor/*"` 2>/dev/null) ; diffFileCount=$(echo -n "$diffFiles" | wc -l); if [ "$diffFileCount" -gt 0 ]; then echo >&2 "Error: files need to be gofmt'ed:"; echo >&2 "$diffFiles"; exit 1; fi
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /tmp/server ./cmd/server


FROM gcr.io/distroless/base
COPY --from=golang /tmp/server /
COPY --from=golang /go/bin/swagger /go/bin/swagger
COPY --from=golang /swagger.json /swagger.json

ENTRYPOINT ["/server"]
