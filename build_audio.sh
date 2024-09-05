#!/usr/bin/env bash

mkdir -p builds

rsync -av ./data_atom/audio ./builds

# zip
pushd .
cd builds || exit

zip -r ./audio-bundle.zip ./audio

popd || exit