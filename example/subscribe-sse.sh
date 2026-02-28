#!/usr/bin/env bash

#PATTERN="$1"
PATTERN='foo/**'

wget -q -S -O- --header='Accept: text/event-stream' "http://localhost:8080/subscribe/$PATTERN"
