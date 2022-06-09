# Build the hanko binary
FROM golang:1.17 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the go source
COPY ../main.go main.go
COPY cmd cmd/
COPY config config/
COPY persistence persistence/
COPY server server/
COPY handler handler/
COPY crypto crypto/
COPY dto dto/
COPY session session/
COPY mail mail/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o hanko main.go

# Use distroless as minimal base image to package hanko binary
# See https://github.com/GoogleContainerTools/distroless for details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/hanko .
USER 65532:65532

ENTRYPOINT ["/hanko"]
