FROM golang:1.24 AS builder

WORKDIR /app

# Copy the mod and sum and install the dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the used code.
COPY api ./api
COPY scheduler ./scheduler
COPY pkg ./pkg

# Run the build.
RUN CGO_ENABLED=0 GOOS=linux  go build -o api ./api

FROM alpine:3.21.3

RUN adduser -D appuser
USER appuser

WORKDIR /app
COPY --from=builder /app/api/api /app/api

CMD ["./api"]
