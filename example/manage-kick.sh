#!/usr/bin/env bash

SUBSCRIBER_ID="$1" # run ./manage-list.sh, select some subscriber's id (UUID), and then pass it as $1

wget -O- -q --method=POST "http://localhost:8080/manage/kick/$SUBSCRIBER_ID"
