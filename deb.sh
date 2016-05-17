#!/bin/sh

VERSION="$(./go-rsyslog-pstats -version 2>&1 | cut -d' ' -f2)"
BUILD="slack1"
set -e -x

DIRNAME="$(cd "$(dirname "$0")" && pwd)"
OLDESTPWD="$PWD"

cd "$(mktemp -d)"
trap "rm -rf \"$PWD\"" EXIT INT QUIT TERM

mkdir -p "$PWD/rootfs/usr/local/bin"
cp "$OLDESTPWD/go-rsyslog-pstats" "$PWD/rootfs/usr/local/bin"

fakeroot fpm -C "$PWD/rootfs" \
    -m "Nate Brown <nate@slack-corp.com>" \
    -n "go-rsyslog-pstats" -v "$VERSION-$BUILD" \
    -p "$OLDESTPWD/go-rsyslog-pstats_${VERSION}-${BUILD}_amd64.deb" \
    --license "MIT" --vendor "" \
    --url "https://github.com/slackhq/go-rsyslog-pstats" \
    --description "Parses and forwards rsyslog process stats to a local statsite or statsd process" \
    -s "dir" -t "deb" \
    "usr"
