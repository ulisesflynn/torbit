# Chat Server

Chat Server is a simple chat server that allows for clients to connect via telnet.

## Running server locally

To run locally: 
```bash
go run main.go
```

For help with possible flags:
```bash
go run main.go --help
```
### OSX Binary

An Mach-O executable is included (chat)

### Building

Project available on github

```bash
git clone git@github.com:ulisesflynn/torbit.git
go build main.go -o chat
```

### Build server for docker

```bash
docker build -t chat .
```

### Run server on docker as daemon

```bash
docker run -p 2000:2000 -p 8080:8080 chat --server-address=0.0.0.0
```

### Connecting via telnet

```bash
telnet 127.0.0.1 2000
```

### Posting a message to chat server via HTTP Rest API

```bash
curl -X POST -d "The Matrix is everywhere" http://127.0.0.1:8080/send_msg/morpheus
```


### Run linting/tests

```bash
./validate.sh
```

### TODO
 - unit tests
