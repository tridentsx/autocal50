.PHONY: dev build build-all build-linux build-windows build-mac clean

APPNAME := autocal50
OUTDIR  := build/bin

dev:
	wails dev -tags webkit2_41

build:
	wails build -tags webkit2_41

build-linux:
	GOOS=linux GOARCH=amd64 wails build -tags webkit2_41 -platform linux/amd64 -o $(OUTDIR)/$(APPNAME)-linux-amd64

build-windows:
	GOOS=windows GOARCH=amd64 wails build -platform windows/amd64 -o $(OUTDIR)/$(APPNAME)-windows-amd64.exe

build-mac:
	wails build -platform darwin/universal -o $(OUTDIR)/$(APPNAME)-darwin-universal

build-all: build-linux build-windows build-mac

clean:
	rm -rf $(OUTDIR)
