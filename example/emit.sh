#!/usr/bin/env bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")"

#EMIT_PATH="$1"
EMIT_URL='foo/bar/baz?one=1'

wget -O- -q -S \
    --header='X-Test: 1' \
    --header='Content-Type: application/json' \
    --post-file=test-body.json \
    "http://localhost:8080/emit/$EMIT_URL"
