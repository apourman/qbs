#!/bin/sh
set -eu

version=${VERSION:-dev}
dist_dir=${DIST_DIR:-dist}
module='github.com/trues/qbs/internal/cli'
ldflags="-s -w -X ${module}.Version=${version}"

mkdir -p "$dist_dir"

for target in \
  linux-amd64 linux-arm64 \
  darwin-amd64 darwin-arm64 \
  windows-amd64 windows-arm64
do
  os=${target%-*}
  arch=${target#*-}
  suffix=
  [ "$os" = windows ] && suffix=.exe
  output="$dist_dir/qbs_${version}_${os}_${arch}${suffix}"
  printf 'building %s\n' "$output"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 \
    go build -trimpath -ldflags "$ldflags" -o "$output" ./cmd/qbs
done

if command -v sha256sum >/dev/null 2>&1; then
  (cd "$dist_dir" && sha256sum qbs_*) > "$dist_dir/SHA256SUMS"
else
  (cd "$dist_dir" && shasum -a 256 qbs_*) > "$dist_dir/SHA256SUMS"
fi
