# Server_HTTP
Server HTTP in Go, REST API, CRUD

## Configuration

```bash
PROJEKT_DIR=`pwd`
mkdir bin
mkdir libs
cd libs
go get -u github.com/go-chi/chi
cd ..
export GOPATH=$PROJEKT_DIR/libs:$PROJEKT_DIR
```

## Server execution

```bash
cd bin
go build main
./main
```

## Test execution

```bash
cd src/main
go test *.go
```
