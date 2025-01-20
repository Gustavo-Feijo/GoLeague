FROM golang:1.23

WORKDIR /app

# Copy the mod and sum and install the dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the used code.
COPY /api ./api
COPY /pkg ./pkg

# Run the build.
RUN go build -o api ./api

CMD ["./api/api"]
