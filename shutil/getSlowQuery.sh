#!/bin/sh

echo "GET Slow Query"
scp -i $1 isucon@118.27.28.210:/var/lib/mysql/slow-query.log .
