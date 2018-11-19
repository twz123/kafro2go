#!/usr/bin/env sh

set -e

wget -q -O /tmp/librdkafka.tar.gz https://github.com/edenhill/librdkafka/archive/v0.11.6.tar.gz
echo '9c0afb8b53779d968225edf1e79da48a162895ad557900f75e7978f65e642032 */tmp/librdkafka.tar.gz' | shasum -a 256 -c -
TMPDIR="$(mktemp -d)"
tar --strip-components=1 -xzf /tmp/librdkafka.tar.gz -C "$TMPDIR"
(
  CDPATH='' cd -- "$TMPDIR"
  ./configure --enable-static --prefix=/usr
  make -j
  make check
  make install
)

go build -a --ldflags '-linkmode external -extldflags "-static"' -tags static_all cmd/kafro2go.go
