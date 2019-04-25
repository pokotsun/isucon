#!/bin/bash

git reset --hard HEAD && git pull origin master
cd webapp/go && make
cd ../..
sudo systemctl restart torb.go
cd bench && bin/bench -remotes=127.0.0.1 -output result.json
