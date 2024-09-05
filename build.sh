#!/usr/bin/env bash

find . -name ".DS_Store" -delete

rm -rf builds
mkdir -p builds/contractor-mac
mkdir -p builds/contractor-win

pushd .
cd /Users/felix/Projects/TextPainter || exit
./build.sh
popd || exit

cp /Users/felix/Projects/TextPainter/builds/mac/mapedit ./builds/contractor-mac/mapedit

cp /Users/felix/Projects/TextPainter/builds/win/mapedit.exe ./builds/contractor-win/mapedit.exe


cp ./config.rec ./builds/contractor-mac/config.rec
cp ./config.rec ./builds/contractor-win/config.rec

#mkdir -p builds/linux
#mkdir -p builds/wasi

GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o ./builds/contractor-mac/contractor .

GOOS=windows GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o ./builds/contractor-win/contractor.exe .

rsync -av ./data_atom ./builds/contractor-mac --exclude audio
#cp -R ./data_atom ./builds/mac/data_atom

rsync -av ./data_atom ./builds/contractor-win --exclude audio
#cp -R ./data_atom ./builds/win/data_atom

cp winRun.cmd winEdit.cmd ./builds/contractor-win

#GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o ./builds/linux/rx .

#upx ./builds/mac/rx
#upx ./builds/win/rx.exe
#upx ./builds/linux/rx

#GOOS=js GOARCH=wasm go build -tags wasm -trimpath -ldflags '-s -w' -o ./builds/wasi/rx.wasm .

# zip

pushd .
cd builds || exit

zip -r ./contractor-mac.zip ./contractor-mac
zip -r ./contractor-win.zip ./contractor-win

popd || exit