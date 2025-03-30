FROM golang:1.23

WORKDIR /app

# Copy the mod and sum and install the dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the used code.
COPY /fetcher ./fetcher
COPY /revalidator ./revalidator
COPY /pkg ./pkg

# Run the build.
RUN go build -o revalidate ./revalidator

CMD ["./revalidate"]
