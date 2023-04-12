# Build the hanko binary
FROM --platform=$BUILDPLATFORM golang:1.19 as builder

ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd cmd/
COPY config config/
COPY persistence persistence/
COPY server server/
COPY handler handler/
COPY crypto crypto/
COPY dto dto/
COPY session session/
COPY mail mail/
COPY audit_log audit_log/
COPY pagination pagination/
COPY rate_limiter rate_limiter/
COPY thirdparty thirdparty/
COPY build_info build_info/
COPY middleware middleware/

# Build
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -o hanko main.go

# Use distroless as minimal base image to package hanko binary
# See https://github.com/GoogleContainerTools/distroless for details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/hanko .
USER 65532:65532

ENTRYPOINT ["/hanko"]
