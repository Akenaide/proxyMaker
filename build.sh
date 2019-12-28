#!/bin/bash
pub global activate webdev

DEST=$1

echo "build to $DEST"
mkdir -p $DEST
mkdir -p $DEST/static

webdev build --output $DEST

go build -o $DEST