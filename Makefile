
build:
	go build -o bin/ytkiosk ./cmd/ytkiosk

run:
	go run ./cmd/ytkiosk/main.go

pkg: build
	cp bin/ytkiosk dpkg/xenial/usr/bin
	IAN_DIR=dpkg/xenial ian pkg
