#!/usr/bin/env bash

#PATTERN="$1"
PATTERN='foo/**'

wget -q -S -O- "http://localhost:8080/subscribe/$PATTERN"
