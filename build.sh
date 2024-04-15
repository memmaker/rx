#!/usr/bin/env bash
rm -rf builds
mkdir -p builds/mac
mkdir -p builds/win
mkdir -p builds/linux
#mkdir -p builds/wasi

GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o ./builds/mac/rx .

GOOS=windows GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o ./builds/win/rx.exe .

GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o ./builds/linux/rx .

upx ./builds/mac/rx
upx ./builds/win/rx.exe
upx ./builds/linux/rx


#GOOS=js GOARCH=wasm go build -tags wasm -trimpath -ldflags '-s -w' -o ./builds/wasi/rx.wasm .