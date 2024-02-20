# Go TCP Chat

This is a command line chat application that consists of a server and a client which communicate over network using TCP protocol.

## Run server

`go run ./cmd/server`

## Run client

### Connect to localhost server

`go run ./cmd/client`

### Explicit server connection params

`go run ./cmd/client -host <server-host> -port <server-port>`
