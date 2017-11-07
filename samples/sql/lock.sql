use performance_schema;
#select * from data_locks;

# use information_schema;
# select * from information_schema.INNODB_TRX;

# select * from performance_schema.data_locks d INNER JOIN information_schema.innodb_trx i where d.ENGINE_TRANSACTION_ID = i.trx_id;

select version();

select PARTITION_NAME,INDEX_NAME,LOCK_TYPE,LOCK_MODE,LOCK_STATUS,LOCK_DATA,trx_id,trx_state,trx_started,trx_requested_lock_id,trx_query,trx_operation_state,trx_tables_in_use,trx_tables_locked,trx_lock_structs,trx_rows_locked,trx_rows_modified,trx_adaptive_hash_latched,trx_autocommit_non_locking from performance_schema.data_locks d INNER JOIN information_schema.innodb_trx i where d.ENGINE_TRANSACTION_ID = i.trx_id;