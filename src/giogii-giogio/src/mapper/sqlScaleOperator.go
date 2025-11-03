package mapper

import (
	"giogii/src/entity"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type SqlScaleOperator interface {
	DoClose()
	DoQueryParseMaster(sqlStr string) entity.MasterStatus
	DoQueryParseSlave(sqlStr string) entity.SlaveStatus
	DoQueryParseString(sqlStr string) string
	DoQueryParseParameter(sqlStr string, args string) (c []entity.Configuration)
	DoQueryParseConsumers(sqlStr string, args string) entity.Consumers
	DoQueryParseValue(sqlStr string) string
	DoQueryParseSingleValue(sqlStr string) string
	DoQueryParseToBigTransaction(sqlStr string) (bt []entity.BigTransaction)
	DoQueryParseToMetadataLocks(sqlStr string) (ml []entity.MetadataLocks)
	DoQueryParseToSysInnodbLockWaits(sqlStr string) (ml []entity.SysInnodbLockWaits)
	DoQueryParseMap(sqlStr string) (m map[string]string)
	DoQueryParseToDataServers(sqlStr string) (d []entity.DataServers)
	DoQueryWithoutRes(sqlStr string)
	DoQueryParseToClusterInfo(sqlStr string) (c []entity.ClusterInfo)
	DoInsertValues(sqlStr string, id int64, args string, args2 string) (count int64)
}

func (sqlScaleStruct *SqlStruct) DoClose() {
	sqlScaleStruct.Connection.Close()
}

func (sqlScaleStruct *SqlStruct) DoQueryParseMaster(sqlStr string) entity.MasterStatus {
	rows := sqlScaleStruct.doQuery(sqlStr)
	var masterStatus entity.MasterStatus
	for rows.Next() {
		rows.Scan(&masterStatus.File, &masterStatus.Position, &masterStatus.BinlogDoDB, &masterStatus.BinlogIgnoreDB, &masterStatus.ExecutedGtidSet)
	}
	return masterStatus
}

func (sqlScaleStruct *SqlStruct) DoQueryParseSlave(sqlStr string) entity.SlaveStatus {
	rows := sqlScaleStruct.doQuery(sqlStr)
	var slaveStatus entity.SlaveStatus
	for rows.Next() {
		err := rows.Scan(&slaveStatus.SlaveIOState, &slaveStatus.MasterHost, &slaveStatus.MasterUser, &slaveStatus.MasterPort, &slaveStatus.ConnectRetry,
			&slaveStatus.MasterLogFile, &slaveStatus.ReadMasterLogPos, &slaveStatus.RelayLogFile, &slaveStatus.RelayLogPos, &slaveStatus.RelayMasterLogFile,
			&slaveStatus.SlaveIORunning, &slaveStatus.SlaveSQLRunning, &slaveStatus.ReplicateDoDB, &slaveStatus.ReplicateIgnoreDB, &slaveStatus.ReplicateDoTable,
			&slaveStatus.ReplicateIgnoreTable, &slaveStatus.ReplicateWildDoTable, &slaveStatus.ReplicateWildIgnoreTable, &slaveStatus.LastErrno,
			&slaveStatus.LastError, &slaveStatus.SkipCounter, &slaveStatus.ExecMasterLogPos, &slaveStatus.RelayLogSpace, &slaveStatus.UntilCondition,
			&slaveStatus.UntilLogFile, &slaveStatus.UntilLogPos, &slaveStatus.MasterSSLAllowed, &slaveStatus.MasterSSLCAFile, &slaveStatus.MasterSSLCAPath,
			&slaveStatus.MasterSSLCert, &slaveStatus.MasterSSLCipher, &slaveStatus.MasterSSLKey, &slaveStatus.SecondsBehindMaster, &slaveStatus.MasterSSLVerifyServerCert,
			&slaveStatus.LastIOErrno, &slaveStatus.LastIOError, &slaveStatus.LastSQLErrno, &slaveStatus.LastSQLError, &slaveStatus.ReplicateIgnoreServerIds, &slaveStatus.MasterServerId,
			&slaveStatus.MasterUUID, &slaveStatus.MasterInfoFile, &slaveStatus.SQLDelay, &slaveStatus.SQLRemainingDelay, &slaveStatus.SlaveSQLRunningState,
			&slaveStatus.MasterRetryCount, &slaveStatus.MasterBind, &slaveStatus.LastIOErrorTimestamp, &slaveStatus.LastSQLErrorTimestamp, &slaveStatus.MasterSSLCrl,
			&slaveStatus.MasterSSLCrlpath, &slaveStatus.RetrievedGtidSet, &slaveStatus.ExecutedGtidSet, &slaveStatus.AutoPosition, &slaveStatus.ReplicateRewriteDB,
			&slaveStatus.ChannelName, &slaveStatus.MasterTLSVersion, &slaveStatus.Masterpublickeypath, &slaveStatus.Getmasterpublickey, &slaveStatus.NetworkNamespace,
		)
		if err != nil {
			log.Println(err)
		}
	}
	return slaveStatus
}

func (sqlScaleStruct *SqlStruct) DoQueryParseString(sqlStr string) string {
	rows := sqlScaleStruct.doQuery(sqlStr)
	var rst string
	var value string
	for rows.Next() {
		err := rows.Scan(&rst, &value)
		if err != nil {
			log.Println(err)
		}
	}
	return value
}

func (sqlScaleStruct *SqlStruct) DoQueryParseParameter(sqlStr string, args string) (c []entity.Configuration) {
	rows := sqlScaleStruct.doPrepareQuery(sqlStr, args)
	for rows.Next() {
		var configuration entity.Configuration
		rows.Scan(&configuration.Name, &configuration.Value, &configuration.Type)
		c = append(c, configuration)
	}
	return
}

func (sqlScaleStruct *SqlStruct) DoQueryParseConsumers(sqlStr string, args string) entity.Consumers {
	rows := sqlScaleStruct.doPrepareQuery(sqlStr, args)
	var consumer entity.Consumers
	for rows.Next() {
		rows.Scan(&consumer.Name, &consumer.Enabled)
	}
	return consumer
}

func (sqlScaleStruct *SqlStruct) DoQueryParseValue(sqlStr string) string {
	rows := sqlScaleStruct.doQuery(sqlStr)
	var optionName string
	var value string
	for rows.Next() {
		err := rows.Scan(&optionName, &value)
		if err != nil {
			log.Println(err)
		}
	}
	return value
}

func (sqlScaleStruct *SqlStruct) DoQueryParseSingleValue(sqlStr string) string {
	rows := sqlScaleStruct.doQuery(sqlStr)
	var value string
	for rows.Next() {
		err := rows.Scan(&value)
		if err != nil {
			log.Println(err)
		}
	}
	return value
}

func (sqlScaleStruct *SqlStruct) DoQueryParseToBigTransaction(sqlStr string) (b []entity.BigTransaction) {
	rows := sqlScaleStruct.doQuery(sqlStr)
	for rows.Next() {
		var bt entity.BigTransaction
		rows.Scan(&bt.ThreadId, &bt.LockCount, &bt.ProcesslistId, &bt.ProcesslistUser, &bt.ProcesslistHost, &bt.SqlText)
		b = append(b, bt)
	}
	return b
}

func (sqlScaleStruct *SqlStruct) DoQueryParseToMetadataLocks(sqlStr string) (ml []entity.MetadataLocks) {
	rows := sqlScaleStruct.doQuery(sqlStr)
	for rows.Next() {
		var m entity.MetadataLocks
		rows.Scan(&m.ObjectType, &m.LockType, &m.LockStatus, &m.ProcesslistId, &m.ProcesslistTime, &m.ProcesslistInfo)
		ml = append(ml, m)
	}
	return ml
}

func (sqlScaleStruct *SqlStruct) DoQueryParseToSysInnodbLockWaits(sqlStr string) (lw []entity.SysInnodbLockWaits) {
	rows := sqlScaleStruct.doQuery(sqlStr)
	for rows.Next() {
		var l entity.SysInnodbLockWaits
		rows.Scan(&l.WaitStarted, &l.WaitAge, &l.WaitAgeSecs, &l.LockedTable, &l.LockedTableSchema,
			&l.LockedTableName, &l.LockedTablePartition, &l.LockedTableSubpartition, &l.LockedIndex, &l.LockedType,
			&l.WaitingTrxId, &l.WaitingTrxStarted, &l.WaitingTrxAge, &l.WaitingTrxRowsLocked, &l.WaitingTrxRowsModified,
			&l.WaitingPid, &l.WaitingQuery, &l.WaitingLockId, &l.WaitingLockMode, &l.BlockingTrxId,
			&l.BlockingPid, &l.BlockingQuery, &l.BlockingLockId, &l.BlockingLockMode, &l.BlockingTrxStarted,
			&l.BlockingTrxAge, &l.BlockingTrxRowsLocked, &l.BlockingTrxRowsModified, &l.SqlKillBlockingQuery, &l.SqlKillBlockingConnection)
		lw = append(lw, l)
	}
	return lw
}

func (sqlScaleStruct *SqlStruct) DoQueryParseMap(sqlStr string) (m map[string]string) {
	rows := sqlScaleStruct.doQuery(sqlStr)
	m = make(map[string]string)
	for rows.Next() {
		var key string
		var value string
		rows.Scan(key, value)
		m[key] = value
	}
	return m
}

func (sqlScaleStruct *SqlStruct) DoQueryParseToDataServers(sqlStr string) (d []entity.DataServers) {
	rows := sqlScaleStruct.doQuery(sqlStr)
	for rows.Next() {
		var ds entity.DataServers
		rows.Scan(&ds.Servername, &ds.Host, &ds.Port, &ds.Username, &ds.Status, &ds.MasterOnlineStatus,
			&ds.MasterBackup, &ds.RemoteUser, &ds.RemotePort, &ds.MaxNeededConn, &ds.MasterPriority)
		d = append(d, ds)
	}
	return d
}

func (sqlScaleStruct *SqlStruct) DoQueryWithoutRes(sqlStr string) {
	sqlScaleStruct.doQuery(sqlStr)
}

func (sqlScaleStruct *SqlStruct) DoQueryParseToClusterInfo(sqlStr string) (c []entity.ClusterInfo) {
	rows := sqlScaleStruct.doQuery(sqlStr)
	for rows.Next() {
		var ds entity.ClusterInfo
		rows.Scan(&ds.MasterDbscale, &ds.ClusterServerId, &ds.Host, &ds.JoinTime, &ds.KaInitVersion, &ds.KaUpdateVersion,
			&ds.DynamicNodeVersion, &ds.DynamicSpaceVersion, &ds.MasterReScrambleDelay, &ds.DbscaleVersion)
		c = append(c, ds)
	}
	return
}

func (sqlScaleStruct *SqlStruct) DoInsertValues(sqlStr string, id int64, args string, args2 string) (count int64) {
	result := sqlScaleStruct.doPrepareInsert(sqlStr, id, args, args2)
	count, err := result.RowsAffected()
	if err != nil {
		log.Println("获取结果失败", err)
	}
	return count
}
