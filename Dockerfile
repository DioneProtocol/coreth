# ============= Compilation Stage ================
FROM golang:1.20.1-buster AS builder

RUN apt-get update && apt-get install -y --no-install-recommends bash=5.0-4 make=4.2.1-1.2 gcc=4:8.3.0-1 musl-dev=1.1.21-2 ca-certificates=20200601~deb10u2 linux-headers-amd64

ARG ODYSSEY_VERSION

RUN mkdir -p $GOPATH/src/github.com/DioneProtocol
WORKDIR $GOPATH/src/github.com/DioneProtocol

RUN git clone -b $ODYSSEY_VERSION --single-branch https://github.com/DioneProtocol/odysseygo.git

# Copy coreth repo into desired location
COPY . coreth

# Set the workdir to OdysseyGo and update coreth dependency to local version
WORKDIR $GOPATH/src/github.com/DioneProtocol/odysseygo
# Run go mod download here to improve caching of OdysseyGo specific depednencies
RUN go mod download
# Replace the coreth dependency
RUN go mod edit -replace github.com/DioneProtocol/coreth=../coreth
RUN go mod download && go mod tidy -compat=1.19

# Build the OdysseyGo binary with local version of coreth.
RUN ./scripts/build_odyssey.sh
# Create the plugins directory in the standard location so the build directory will be recognized
# as valid.
RUN mkdir build/plugins

# ============= Cleanup Stage ================
FROM debian:11-slim AS execution

# Maintain compatibility with previous images
RUN mkdir -p /odysseygo/build
WORKDIR /odysseygo/build

# Copy the executables into the container
COPY --from=builder /go/src/github.com/DioneProtocol/odysseygo/build .

CMD [ "./odysseygo" ]
