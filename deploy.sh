#!/bin/bash

if [ $# == 1 ]; then
	echo "execute checkout and pull!!"
	git reset --hard HEAD && git fetch && git checkout $1 && git pull origin $1
fi

## Database data reset
sudo ./db/db_setup.sh

## application build
(
cd go
make
)

## replaces log data 
LOGPATH=/var/log
NOW=`date +'%H-%M-%S'`

#sudo cp $LOGPATH/nginx/access.log $LOGPATH/nginx/access-$NOW.log
#sudo sh -c 'echo "" > /var/log/nginx/access.log'
#
#sudo cp $LOGPATH/mariadb/slow.log $LOGPATH/mariadb/slow-$NOW.log
#sudo sh -c 'echo "" > /var/log/mariadb/slow.log'

## replace mysql conf
#sudo cp conf/mysqld.cnf /etc/mysql/mysql.conf.d/mysqld.cnf

## restart application services
## db, app, nginx, redis
echo 'systemctl are restarting...'
sudo systemctl restart mysql.service
sudo systemctl restart isuda.go.service
sudo systemctl restart isutar.go.service
sudo systemctl restart nginx.service
sudo systemctl restart redis.service
echo 'Finished to restart!!'

## execute bench marker and analysis tools
(
cd ../isucon6q/
./isucon6q-bench -target http://127.0.0.1 > result.json
jq . < result.json
)
#sudo /usr/local/bin/alp -f /var/log/nginx/access.log -r --sum | head -n 30
#sudo mysqldumpslow -s t /var/log/mariadb/slow.log | head -n 30
