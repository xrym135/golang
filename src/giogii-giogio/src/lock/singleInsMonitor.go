package lock

import (
	"fmt"
	"giogii/src/mapper"
	"log"
	"strconv"
)

var SourceSqlMapper mapper.SqlScaleOperator
var TargetSocket string

func InitConf(sourceUserInfo string, sourceSocket string, sourceDatabase string) {
	s := mapper.InitSourceConn(sourceUserInfo, sourceSocket, sourceDatabase)
	SourceSqlMapper = &s
}

/*
	1) top –Hp 进程
	可以找到打满CPU的线程是哪个，对应的pid就是performance_schema.threads表中的
	thread_os_id，也就能找processlist_id，通过 show processlist 找到对应的执行用户窗
	口，找到相应用户信息，还可以通过performance_schema.events_statements_history 表找
	到对应执行的语句。

	2) %Cpu(s): sy us wa
	CPU这3个状态能简单判断锁的情况，只是wait高一般跟IO有关，如果是us高的话，更可能是慢查
	询导致，如果是sys和wait都高，大几率锁定是锁的问题。

	1.元数据锁排查
	1.1 show processlist; 查看自己的锁类型，案例情况下，当为 Waiting for global read lock

	1.2 select * from performance_schema.metadata_locks;
	—> PENDING 对象是正在等待元数据锁，在通过库表信息，看谁锁了自己想要的元数据
	—> GRANTED 对象是已经获取了元数据锁，对应上自己想要的元数据，确定OWNER_THREAD_ID 的值
	—> select * from performance_schema.threads where THREAD_ID = num; 确定 PROCESSLIST_ID 的值

	1.3 show processlist; 比对 PROCESSLIST_ID 看对方是谁，正在做什么，在确认可以杀死的话，kill Id

	2. 行锁排查
	2.1 确认有没有锁等待:
	show status like 'innodb_row_lock%';
	mysql> show status like 'innodb_row_lock%';
	+-------------------------------+--------+
	| Variable_name                 | Value  |
	+-------------------------------+--------+
	| Innodb_row_lock_current_waits | 0      | 值等多少，就有多少个正在锁等待
	| Innodb_row_lock_time          | 859482 |
	| Innodb_row_lock_time_avg      | 2933   |
	| Innodb_row_lock_time_max      | 601241 |
	| Innodb_row_lock_waits         | 293    |
	+-------------------------------+--------+
	select * from information_schema.innodb_trx;

	2.2查询锁等待详细信息
	select * from sys.innodb_lock_waits; —> blocking_pid(锁源的连接线程，假设等于30)

	2.3 通过连接线程找SQL线程
	select * from performance_schema.threads where processlist_id=30; -–> thread_id（假设等于67）

	2.4 通过SQL线程找到 SQL语句
	select thread_id,SQL_TEXT from performance_schema.events_statements_history where thread_id=67;
	show processlist; —> 找到id=30的用户是谁，拿着语句向对方确认是否可以杀死，释放锁。


	3. 解决死锁
	1.一种策略是，直接进入等待，直到超时。这个超时时间可以通过参数 innodb_lock_wait_timeout 来设置。(消耗时间，并发弱)
	2.另一种策略是，发起死锁检测，发现死锁后，主动回滚死锁链条中的某一个事务，让其他事务得以继续执行。将参数 innodb_deadlock_detect 设置为 on，表示开启这个逻辑。(大量并发，死锁检测消耗大量的cpu)
*/

/*
flush table with read lock;  在16310里不好用
unlock tables;

*/
var strSql string
var BaseSqlScaleOperator mapper.SqlScaleOperator

func DoMonitorLock() {

	defer func() {
		SourceSqlMapper.DoClose()
	}()

	/**
	1） 判断是否有长时间运行的事务  -L l
	*/
	// 1.1 超过60秒的事务有几个
	strSql = fmt.Sprint("select count(*) from information_schema.INNODB_TRX i inner join information_schema.PROCESSLIST p on i.trx_mysql_thread_id = p.ID where p.TIME > 60")
	trxLongerThanSixtySecondCount := SourceSqlMapper.DoQueryParseSingleValue(strSql)
	count, _ := strconv.Atoi(trxLongerThanSixtySecondCount)
	if count > 0 {
		log.Printf("超过60秒的事务>  共计: %s 个", trxLongerThanSixtySecondCount)
		//strSql = fmt.Sprint("select p.ID, p.USER, p.HOST, p.DB, p.TIME, s.SQL_TEXT from information_schema.INNODB_TRX i inner join information_schema.PROCESSLIST p on i.trx_mysql_thread_id = p.ID inner join performance_schema.threads t on i.trx_mysql_thread_id = t.PROCESSLIST_ID inner join performance_schema.events_statements_current s on t.THREAD_ID = s.THREAD_ID")
	}

	/**
	2) 判断当前环境是否有大事务锁了多行  -L t
	select count(*) from data_locks where LOCK_MODE <> 'IX';
	select THREAD_ID,count(THREAD_ID) from performance_schema.data_locks where LOCK_MODE <> 'IX' group by THREAD_ID;
	*/
	strSql = fmt.Sprint("select l.THREAD_ID,l.LOCK_COUNT ,t.PROCESSLIST_ID,t.PROCESSLIST_USER,t.PROCESSLIST_HOST ,p.SQL_TEXT from (select THREAD_ID,count(THREAD_ID) as LOCK_COUNT from performance_schema.data_locks where LOCK_MODE <> 'IX' and LOCK_TYPE <> 'TABLE' group by THREAD_ID) l left join performance_schema.threads t on l.THREAD_ID = t.THREAD_ID left join performance_schema.events_statements_current p  on l.THREAD_ID = p.THREAD_ID;")
	bt := SourceSqlMapper.DoQueryParseToBigTransaction(strSql)
	if len(bt) > 0 {
		for i := 0; i < len(bt); i++ {
			b := bt[i]
			if *b.LockCount > 0 {
				log.Print("大事务行锁检查> ", " 锁定行数: ", *b.LockCount, "; PROCESS_ID: ", *b.ProcesslistId, "; 连接主机: ", b.ProcesslistHost, "; 连接用户: ", b.ProcesslistUser, "; 执行SQL: ", b.SqlText)
			}
		}
	}

	/**
	3）判断当前环境是否存在锁等待   -L w
	等待时长要记录
	*/
	strSql = fmt.Sprint("show status like 'Innodb_row_lock_current_waits'")
	innodbRowLockCurrentWaits := SourceSqlMapper.DoQueryParseString(strSql)
	value, _ := strconv.Atoi(innodbRowLockCurrentWaits) //strconv.ParseInt(innodbRowLockCurrentWaits, 10, 64)
	if value > 0 {                                      // 如果正在等待锁的值大于0，
		log.Print("行锁锁等待检查> ", " 当前环境至少存在", value, "个锁等待")
		strSql = fmt.Sprint("select * from sys.innodb_lock_waits")
		lw := SourceSqlMapper.DoQueryParseToSysInnodbLockWaits(strSql)
		if len(lw) > 0 {
			for i := 0; i < len(lw); i++ {
				l := lw[i]
				log.Print("(", i+1, ") 语句> ", " ", l.WaitingQuery.String, " ; 被PROCESS_ID : ", *l.BlockingPid, " 阻塞;", " 可执行: ", l.SqlKillBlockingQuery.String, " 解除; ")
			}
		}
	}

	/**
	4) 判断当前环境是否存在MDL锁等待且阻塞现象
	*/

	strSql = fmt.Sprint("select m.OBJECT_TYPE,m.LOCK_TYPE,m.LOCK_STATUS, t.PROCESSLIST_ID,t.PROCESSLIST_TIME,t.PROCESSLIST_INFO from performance_schema.metadata_locks m inner join performance_schema.threads t on m.OWNER_THREAD_ID = t.THREAD_ID where m.LOCK_STATUS = 'PENDING' order by t.PROCESSLIST_TIME DESC ")
	ml := SourceSqlMapper.DoQueryParseToMetadataLocks(strSql)
	if len(ml) > 0 {
		for i := 0; i < len(ml); i++ {
			m := ml[i]
			log.Print("MDL锁检查> ", " 锁对象类型: ", m.ObjectType, "; 锁状态: ", m.LockStatus, "; PROCESS_ID: ", *m.ProcesslistId, "; 执行时间: ", m.ProcesslistTime, "; 执行SQL: ", m.ProcesslistInfo)
		}
	}

}
