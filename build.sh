#!/bin/bash
pub global activate webdev

DEST=$1

echo "build to $DEST"

webdev build --output $DEST

mkdir -p $DEST
mkdir -p $DEST/static
go build -o $DEST
