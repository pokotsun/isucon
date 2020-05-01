#!/bin/sh

HOSTS="isucon81 isucon82 isucon83"
echo "Deploy Start"
for host in ${HOSTS}; do
ssh -t $host <<EOC
cd torb
git reset --hard HEAD
git pull origin master
cd webapp/go
make
sudo systemctl enable torb.go
sudo systemctl restart torb.go
EOC
done

#ssh -t isucon1 <<EOC
#sudo systemctl disable mariadb
#sudo systemctl restart mariadb
#EOC

ssh -t isucon83 <<EOC
cd torb/conf
sudo cp nginx.conf /etc/nginx/nginx.conf
sudo systemctl enable nginx 
sudo systemctl restart nginx 
EOC
echo "Deploy Ended"

