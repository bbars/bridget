#!/usr/bin/env bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")"

go run github.com/rakyll/hey@latest \
    -n 10000 -c 200 \
    -z 1m \
    -H 'X-Test: 1' \
    -H 'X-Test: 2' \
    -H 'X-Test: 3' \
    -T 'application/json' \
    -D test-body.json \
    -m POST \
    http://localhost:8080/emit/foo/35gmEIc2mP/1/2/3/ddd?bar=baz
