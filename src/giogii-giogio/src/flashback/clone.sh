#!/bin/bash
username=$1
password=$2
ip=$3
port=$4
/data/app/mysql-8.0.26/bin/mysql -uroot -S /data/mysqldata/clonebackup/socket/mysql.sock -e "set global super_read_only=0;INSTALL PLUGIN clone SONAME 'mysql_clone.so';set global clone_autotune_concurrency = off;set global clone_buffer_size=33554432;set global clone_max_concurrency=32;"
echo "INSTALL"
sleep 5s
nohup /data/app/mysql-8.0.26/bin/mysql -uroot -S /data/mysqldata/clonebackup/socket/mysql.sock -e "SET GLOBAL clone_valid_donor_list = '${ip}:${port}';CLONE INSTANCE FROM '${username}'@'${ip}':${port} IDENTIFIED BY '${password}';" > out.log 2>&1 &
echo "CLONE"
sleep 5s