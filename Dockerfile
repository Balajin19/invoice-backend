FROM golang:1.25

WORKDIR /app

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy all code
COPY . .

# Build from cmd/main.go
RUN go build -o main ./cmd/main.go

EXPOSE 8080

CMD ["./main"]