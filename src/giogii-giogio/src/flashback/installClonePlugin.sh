#!/bin/bash
socketDir=`find /data/mysqldata/16* -name mysql.sock`
/data/app/mysql-8.0.26/bin/mysql -uadmin -p'!QAZ2wsx' -S ${socketDir} -e "set global super_read_only=0;INSTALL PLUGIN clone SONAME 'mysql_clone.so';"
/data/app/mysql-8.0.26/bin/mysql -uadmin -p'!QAZ2wsx' -S ${socketDir} -e "set global clone_autotune_concurrency = off;set global clone_buffer_size=33554432;set global clone_max_concurrency=32;set global super_read_only=1;"