#!/usr/bin/env sh

set -e


# wget -q -O /tmp/musl.tar.gz http://www.musl-libc.org/releases/musl-1.1.20.tar.gz
# echo '44be8771d0e6c6b5f82dd15662eb2957c9a3173a19a8b49966ac0542bbd40d61 */tmp/musl.tar.gz' | shasum -a 256 -c -
wget -q -O /tmp/librdkafka.tar.gz https://github.com/edenhill/librdkafka/archive/v0.11.6.tar.gz
echo '9c0afb8b53779d968225edf1e79da48a162895ad557900f75e7978f65e642032 */tmp/librdkafka.tar.gz' | shasum -a 256 -c -
# (
#   TMPDIR="$(mktemp -d)"
#   tar --strip-components=1 -xzf /tmp/musl.tar.gz -C "$TMPDIR"
#   CDPATH='' cd -- "$TMPDIR"
#   ./configure
#   make
#   sudo make install
# )
(
  TMPDIR="$(mktemp -d)"
  tar --strip-components=1 -xzf /tmp/librdkafka.tar.gz -C "$TMPDIR"
  CDPATH='' cd -- "$TMPDIR"
  ./configure --prefix=/usr/local
  make -j
  make check
  sudo make install
)

go build -a --ldflags '-linkmode external -extldflags "-static"' -tags static cmd/kafro2go.go
