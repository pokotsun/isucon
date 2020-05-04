#!/bin/bash

git reset --hard HEAD && git pull origin dev
#sudo ./isubata/db/init.sh
#zcat ~/isubata/bench/isucon7q-initial-dataset.sql.gz | sudo mysql isubata
cd webapp/go && make
cd ../..
## logからデータを削除
sudo sh -c 'echo "" > /var/log/nginx/access.log'
sudo sh -c 'echo "" > /var/log/mysql/slow.log'
sudo cp conf/my.cnf /etc/mysql/my.cnf
#sudo sh -c 'echo "" > /var/log/mariadb/slow.sql'
#sudo sh -c 'echo "" > /var/log/mariadb/general.sql'
sudo systemctl restart isuda.go
echo 'Finished to restart isuda!!'
sudo systemctl restart nginx.service
echo 'Finished to restart nginx!!'
sudo systemctl restart mysql.service
echo 'Finished to restart mysql!!'
#cd isubata/bench && ./bin/bench -remotes=127.0.0.1 -output result.json
#jq . < result.json
