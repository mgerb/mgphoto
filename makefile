VERSION := $(shell git describe --tags)

linux32:
	GOOS=linux GOARCH=386 go build -o ./dist/linux-32/mgphoto -ldflags="-X main.version=${VERSION}" ./*.go

linux64:
	GOOS=linux GOARCH=amd64 go build -o ./dist/linux-64/mgphoto -ldflags="-X main.version=${VERSION}" ./*.go
	
mac:
	GOOS=darwin GOARCH=amd64 go build -o ./dist/mac/mgphoto -ldflags="-X main.version=${VERSION}" ./*.go
	
windows:
	GOOS=windows GOARCH=386 go build -o ./dist/windows/mgphoto.exe -ldflags="-X main.version=${VERSION}" ./*.go

clean:
	rm -rf ./dist

all: linux32 linux64 mac windows
