FROM golang:1.14

VOLUME /var/log/
ENV LOG_PATH=/var/log/
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

COPY . .

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

CMD  ["./main"]