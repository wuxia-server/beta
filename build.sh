#go build -o improving-server improving-server.go

 export GOPATH=$GOPATH:$PWD

 echo $GOPATH

 echo "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ludo-server"

 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ludo-server main.go

 echo "编译完成"