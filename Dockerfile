FROM golang:1.24-alpine

# Set maintainer label
LABEL maintainer=your.email@example.com

# Set working directory
WORKDIR /src

# Copy files to the working directory
COPY . .

# List items in the working directory (ls)
RUN ls -l

COPY go.mod go.sum ./
RUN go mod download

# Build the Go app as myapp binary and move it to /usr/
RUN go build -o myapp main.go && mv myapp /usr/

# Expose port 8888
EXPOSE 8888

# Run the service myapp when a container of this image is launched
CMD ["/usr/myapp"]
