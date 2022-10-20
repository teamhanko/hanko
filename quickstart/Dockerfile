# Build the quickstart binary
FROM golang:1.17 as builder

WORKDIR /workspace
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o quickstart main.go

# Use distroless as minimal base image to package quickstart binary
# See https://github.com/GoogleContainerTools/distroless for details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/quickstart .
COPY /public /public
USER 65532:65532

ENTRYPOINT ["/quickstart"]
