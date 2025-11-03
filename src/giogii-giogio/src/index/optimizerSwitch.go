package index

/**


show variables like '%optimizer_switch%'\G
dbscale backend server execute "set global optimizer_switch = 'mrr=on,mrr_cost_based=on,batched_key_access=on';";
*/

/**
mysql> show variables like '%slave_rows_search_algorithms%';
+------------------------------+----------------------+
| Variable_name                | Value                |
+------------------------------+----------------------+
| slave_rows_search_algorithms | INDEX_SCAN,HASH_SCAN |
+------------------------------+----------------------+
1 row in set (0.01 sec)
*/
