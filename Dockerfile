FROM golang:1.24-alpine as builder

WORKDIR /src

# Copy go.mod and go.sum to cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your code
COPY . .

# Build your Go app
RUN go build -o myapp .

FROM alpine:3.21

WORKDIR /usr

# Copy ONLY the binary from the builder stage
COPY --from=builder /src/myapp .

# Expose your app port
EXPOSE 8888

# Start your app
CMD ["./myapp"]
