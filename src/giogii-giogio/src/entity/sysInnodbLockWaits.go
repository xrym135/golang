package entity

import "database/sql"

type SysInnodbLockWaits struct {
	WaitStarted               sql.NullString
	WaitAge                   sql.NullString
	WaitAgeSecs               *int64
	LockedTable               sql.NullString
	LockedTableSchema         sql.NullString
	LockedTableName           sql.NullString
	LockedTablePartition      sql.NullString
	LockedTableSubpartition   sql.NullString
	LockedIndex               sql.NullString
	LockedType                sql.NullString
	WaitingTrxId              *int64
	WaitingTrxStarted         sql.NullString
	WaitingTrxAge             sql.NullString
	WaitingTrxRowsLocked      *int64
	WaitingTrxRowsModified    *int64
	WaitingPid                *int64
	WaitingQuery              sql.NullString
	WaitingLockId             sql.NullString
	WaitingLockMode           sql.NullString
	BlockingTrxId             *int64
	BlockingPid               *int64
	BlockingQuery             sql.NullString
	BlockingLockId            sql.NullString
	BlockingLockMode          sql.NullString
	BlockingTrxStarted        sql.NullString
	BlockingTrxAge            sql.NullString
	BlockingTrxRowsLocked     *int64
	BlockingTrxRowsModified   *int64
	SqlKillBlockingQuery      sql.NullString
	SqlKillBlockingConnection sql.NullString
}
