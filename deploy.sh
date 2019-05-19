#!/bin/bash

git reset --hard HEAD && git pull origin app/divide-app-code
sudo ./isubata/db/init.sh
zcat ~/isubata/bench/isucon7q-initial-dataset.sql.gz | sudo mysql isubata
cd isubata/webapp/go && make
cd ../../..
## logからデータを削除
sudo sh -c 'echo "" > /var/log/nginx/access.log'
#sudo sh -c 'echo "" > /var/log/mariadb/slow.sql'
#sudo sh -c 'echo "" > /var/log/mariadb/general.sql'
sudo systemctl restart nginx.service
sudo systemctl restart mysql.service
sudo systemctl restart isubata.golang
cd isubata/bench && ./bin/bench -remotes=127.0.0.1 -output result.json
jq . < result.json
