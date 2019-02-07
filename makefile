VERSION := $(shell git describe --tags)

linux:
	GOOS=darwin GOARCH=386 go build -o ./dist/mgphoto-linux -ldflags="-X main.version=${VERSION}" ./*.go

mac:
	GOOS=darwin GOARCH=amd64 go build -o ./dist/mgphoto-mac -ldflags="-X main.version=${VERSION}" ./*.go
	
windows:
	GOOS=windows GOARCH=386 go build -o ./dist/mgphoto-windows.exe -ldflags="-X main.version=${VERSION}" ./*.go

clean:
	rm -rf ./dist

all: linux mac windows
