#!/bin/bash

set -ex
cd $(dirname $0)

myuser=root
mypass=root

# Isuda
isuda_mydb=isuda
mysql -u${myuser} -p${mypass} -e "DROP DATABASE IF EXISTS ${isuda_mydb}; CREATE DATABASE ${isuda_mydb}"
mysql -u${myuser} -p${mypass} ${isuda_mydb} < ./isuda.sql
mysql -u${myuser} -p${mypass} ${isuda_mydb} < ./isuda_user.sql
mysql -u${myuser} -p${mypass} ${isuda_mydb} < ./isuda_entry.sql

# Isutar
#isutar_mydb=isutar
#mysql -u${myuser} -p${mypass} -e "DROP DATABASE IF EXISTS ${isutar_mydb}; CREATE DATABASE ${isutar_mydb}"
#mysql -u${myuser} -p${mypass} ${isutar_mydb} < ./isutar.sql
