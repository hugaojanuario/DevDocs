#!/bin/sh
set -eu
mkdir -p bin/release
for platform in darwin/arm64 darwin/amd64 linux/arm64 linux/amd64 windows/amd64 windows/arm64; do
  os=${platform%/*}
  arch=${platform#*/}
  extension=''
  if [ "$os" = windows ]; then extension=.exe; fi
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags='-s -w' -o "bin/release/devdocs-$os-$arch$extension" ./cmd/devdocs
done
