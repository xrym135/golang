package lock

/**
mysql> dbscale show session id with dataserver = normal_0_1 connection = 3689080;
+------------+------------+
| Cluster id | Session_id |
+------------+------------+
| 2          | 1835715    |
+------------+------------+
1 row in set (0.01 sec)

mysql> dbscale show user status 1835715;
+---------+--------------------+----------------+------------+------------------------------+------------+
| User_id | Cur_schema         | Working State  | Extra Info | Kept Conn List               | Cluster id |
+---------+--------------------+----------------+------------+------------------------------+------------+
| 1835715 | information_schema | in-transaction | testdb     | 172.17.139.27-16315-3689080; | 2          |
+---------+--------------------+----------------+------------+------------------------------+------------+
1 row in set (0.01 sec)

mysql> dbscale show innodb_lock_waiting status;
*/
