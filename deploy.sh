#!/bin/bash

git reset --hard HEAD && git pull origin master
cd webapp/go && make
cd ../..
bash ./db/init.sh
## logからデータを削除
sudo sh -c 'echo "" > /var/log/nginx/access.log'
sudo sh -c 'echo "" > /var/log/mariadb/slow.sql'
sudo sh -c 'echo "" > /var/log/mariadb/general.sql'
# sudo systemctl restart mariadb.service
sudo systemctl restart torb.go
cd bench && bin/bench -remotes=127.0.0.1 -output result.json
