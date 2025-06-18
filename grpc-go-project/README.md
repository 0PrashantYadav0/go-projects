# gRPC Go Project

This project demonstrates different types of gRPC communication patterns using Go:

- Unary RPC
- Server streaming RPC
- Client streaming RPC
- Bidirectional streaming RPC

## Project Structure

```plaintext
grpc-go-project/
├── proto/           # Protocol buffer definitions
│   └── greet.proto  # Service definitions
├── server/          # Server implementation
│   ├── main.go
│   ├── unary.go
│   ├── server_stream.go
│   ├── client_stream.go
│   └── bi_stream.go
└── client/          # Client implementation
    ├── main.go
    ├── unary.go
    ├── server_stream.go
    ├── client_stream.go
    └── bi_stream.go
```

## Prerequisites

- Go 1.20+
- Protocol Buffers compiler (protoc)
- Go plugins for Protocol Buffers

## Setup

### Install Protocol Buffer Compiler

```bash
# macOS
brew install protobuf

# Linux
apt install -y protobuf-compiler

# Verify installation
protoc --version
```

### Install Go plugins for Protocol Buffers

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Ensure the binary installation directory is in your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Generate Go code from Protocol Buffers

```bash
# From project root directory
protoc --go_out=. --go-grpc_out=. proto/greet.proto
```

## Running the Application

### Start the server

```bash
cd server
go run *.go
```

### Run the client

```bash
cd client
go run *.go
```

## Features

1. **Unary RPC** - Simple request/response
2. **Server Streaming** - Server sends multiple responses
3. **Client Streaming** - Client sends multiple requests
4. **Bidirectional Streaming** - Both client and server send multiple messages

To test different communication patterns, uncomment the respective function calls in `client/main.go`.

## License

MIT
