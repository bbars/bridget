#!/usr/bin/env bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")"

go run .. -b localhost:8080
