# Server_HTTP
Server HTTP in Go, REST API, CRUD

## Instalation

```bash
git clone https://github.com/MonikaBer/base64-Monika-Berlinska-.git
cd base64-Monika-Berlinska-
PROJEKT_DIR=`pwd`
mkdir bin
mkdir libs
cd libs
go get -u github.com/go-chi/chi
cd ..
export GOPATH=$PROJEKT_DIR/libs:$PROJEKT_DIR
```

## Execution Server

```bash
cd bin
go build main
./main
```

## Execution Tests

```bash
cd src/main
go test *.go
```
