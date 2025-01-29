FROM golang:1.23

WORKDIR /app

# Copy the mod and sum and install the dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the used code.
COPY /fetcher ./fetcher
COPY /pkg ./pkg

RUN wget -O /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.37/grpc_health_probe-linux-amd64 && \
  chmod +x /bin/grpc_health_probe

# Run the build.
RUN go build -o fetcher ./fetcher

CMD ["./fetcher/fetcher"]
