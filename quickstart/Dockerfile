# Build the quickstart binary
FROM --platform=$BUILDPLATFORM golang:1.20 as builder

ARG TARGETARCH

WORKDIR /workspace
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -o quickstart main.go

# Use distroless as minimal base image to package quickstart binary
# See https://github.com/GoogleContainerTools/distroless for details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/quickstart .
COPY /public /public
USER 65532:65532

ENTRYPOINT ["/quickstart"]
