# Scanner image for https://github.com/teamhanko/hanko (main branch).
# Hanko is a Go authentication server (backend/) backed by PostgreSQL. This build fetches every Go
# dependency, builds the hanko binary as backend/Dockerfile does, prepares a local PostgreSQL with the
# schema migrated, and runs the unit tests. threat_model.md, next to this file, describes how to use it.

# Go toolchain matching backend/go.mod (go 1.26.1).
FROM golang:1.26-trixie

ENV DEBIAN_FRONTEND=noninteractive \
    GOTOOLCHAIN=local \
    HANKO_VERSION=dev

# PostgreSQL is the dialect the test suite and the default config use.
RUN apt-get update && apt-get install -y --no-install-recommends \
      build-essential ca-certificates curl git jq postgresql postgresql-client procps \
    && rm -rf /var/lib/apt/lists/*

# scanner contract: the checkout lives inside the image, at /src
COPY . /src
WORKDIR /src/backend

# Fetch every dependency the build or the tests need; after the image build there is no network.
RUN go mod download

# Build as CI and backend/Dockerfile do; keep debug info for the scanner.
RUN go generate ./... \
    && go build ./... \
    && go build -o /usr/local/bin/hanko .

# Local database matching backend/config/config.yaml (user/password hanko, database hanko on
# localhost:5432), with the current schema applied. Start it again with `service postgresql start`.
RUN service postgresql start \
    && runuser -u postgres -- psql -v ON_ERROR_STOP=1 \
         -c "CREATE ROLE hanko WITH LOGIN SUPERUSER PASSWORD 'hanko'" \
         -c "CREATE DATABASE hanko OWNER hanko" \
         -c "CREATE DATABASE hanko_test OWNER hanko" \
    && hanko migrate up --config config/config.yaml \
    && service postgresql stop

# Everything is fetched. From here on the go command must not reach a proxy or download a toolchain.
ENV GOPROXY=off \
    GOSUMDB=off

# The database-backed suites (handler, persistence, saml, webhooks, ...) start PostgreSQL through
# dockertest, which needs a Docker daemon the scanner does not have; -short skips them.
RUN timeout 1200 go test -short ./... \
    || echo "WARNING: go test -short reported failures"

CMD ["bash"]
