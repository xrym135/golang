#!/bin/bash
username=$1
password=$2
string=`ls /data/mysqldata/`
array=(${string// /})
conf=`find /data/mysqldata/16*/ -name *.conf`
/data/app/mysql-8.0.26/bin/mysql -u${username} -p${password} -S /data/mysqldata/clonebackup/socket/mysql.sock -e "shutdown;"
sleep 2s
/data/app/mysql-8.0.26/bin/mysql -u${username} -p${password} -S /data/mysqldata/${array}/socket/mysql.sock -e "shutdown;"
sleep 3s
rm -rf /data/mysqldata/${array}/dbdata_bak
mv /data/mysqldata/${array}/dbdata /data/mysqldata/${array}/dbdata_bak
mv /data/mysqldata/clonebackup/dbdata /data/mysqldata/${array}/

nohup /bin/sh /data/app/mysql-8.0.26/bin/mysqld_safe --defaults-file=${conf} --user=mysql --datadir=/data/mysqldata/${array}/dbdata > /data/mysqldata/${array}/logfile/out.log 2>&1 &

sleep 10s
args=3
while [ $args -gt 0 ]
do
  echo -n "${args}"
  CMD=`timeout 4 /data/app/mysql-8.0.26/bin/mysql -p${password} -u${username} -S /data/mysqldata/${array}/socket/mysql.sock --connect-timeout=3 -A -e 'select 1;'`
  if [ -n "${CMD}" ]; then
    break
  fi
  nohup /bin/sh /data/app/mysql-8.0.26/bin/mysqld_safe --defaults-file=${conf} --user=mysql --datadir=/data/mysqldata/${array}/dbdata > /data/mysqldata/${array}/logfile/out.log 2>&1 &
  let args-=1
  sleep 20s
done

rm -rf /data/mysqldata/clonebackup