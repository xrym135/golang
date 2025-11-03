package flashback

import (
	"fmt"
	"giogii/src/entity"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetPosAndSet() (masterStatus entity.MasterStatus) {
	var strSql string
	strSql = fmt.Sprint("show master status")
	masterStatus = SlaveSqlMapper.DoQueryParseMaster(strSql)
	return
}

func SaveInfo(gtid string) int64 {
	var strSql string
	strSql = fmt.Sprint("create table dbscale_tmp.gtid (id int primary key auto_increment, val varchar(1024))")
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
	strSql = fmt.Sprint("insert into dbscale_tmp.gtid (id,val) values (?,?) on duplicate key update val=?")
	count := SlaveSqlMapper.DoInsertValues(strSql, 1, gtid, gtid)
	return count
}

func DoBeginFlashback(sourceUserInfo string, sourceSocket string, targetUserInfo string, targetSocket string) {
	InitMasterConnection(sourceUserInfo, sourceSocket)
	InitSlaveConnection(targetUserInfo, targetSocket)

	defer func() {
		SlaveSqlMapper.DoClose()
		MasterSqlMapper.DoClose()
	}()

	// 1.1 断开主备集群的复制，主集群踢出、备集群断开
	RemoveSlaveCluster()
	CloseReplication()

	// 1.2 等待binlog回放完成
	/**
	获取灾备集群的GTID，确保灾备集群的数据全部回放完成
	*/
	for {
		GetSlaveGTIDSet()
		log.Println("等待灾备集群回放Binlog")
		if SlaveStatus.SecondsBehindMaster.Int64 == 0 {
			log.Println("灾备集群回放Binlog完成", SlaveStatus.ExecutedGtidSet)
			break
		}
		time.Sleep(3 * time.Second)
	}

	// 1.3 记录备集群GTID和POS位点信息，记录备集群拓扑关系、IP信息
	masterStatus := GetPosAndSet()
	log.Println(": ", masterStatus.File)
	log.Println(": ", *masterStatus.Position)
	log.Println(": ", masterStatus.ExecutedGtidSet)

	// 1.4 关闭备集群只读参数，变为read write
	CloseReadOnly()

	// 1.5 保留gtid到数据库里
	count := SaveInfo(masterStatus.ExecutedGtidSet)
	// TODO
	if !(count > 0) {
		EnableReadOnly()
		log.Println(": ", masterStatus.File)
	}

}

func DoEndFlashback(sourceUserInfo string, sourceSocket string, targetUserInfo string, targetSocket string, sshUser string, sshPass string) {
	InitMasterConnection(sourceUserInfo, sourceSocket)
	InitSlaveConnection(targetUserInfo, targetSocket)

	defer func() {
		SlaveSqlMapper.DoClose()
		MasterSqlMapper.DoClose()
	}()
	// 2.1 打开备集群只读参数，变为read only
	EnableReadOnly()

	// 2.2 记录备集群主节点GTID和POS位点信息，
	masterStatus := GetPosAndSet()

	// 2.3 根据binlog位点信息、GTID信息调用dbscale_binlog_tool执行闪回动作
	var primaryPort string
	var secondaryPort string
	var joinerPort string
	var primaryHost string
	var secondaryHost string
	var joinerHost string
	strSql := fmt.Sprint("dbscale show dataservers")
	m := SlaveSqlMapper.DoQueryParseToDataServers(strSql)
	for i := 0; i < len(m); i++ {
		if m[i].MasterOnlineStatus.String == "Master_Online" {
			primaryPort = m[i].Port.String
			primaryHost = m[i].Host.String
		} else if secondaryPort == "" {
			secondaryPort = m[i].Port.String
			secondaryHost = m[i].Host.String
		} else {
			joinerPort = m[i].Port.String
			joinerHost = m[i].Host.String
		}
	}

	// 从dbscale_tmp.gtid获取gtid信息
	var resSet string
	strSql = fmt.Sprint("select val from dbscale_tmp.gtid where id = 1")
	valueSet := SlaveSqlMapper.DoQueryParseSingleValue(strSql)
	resSet = strings.ReplaceAll(valueSet, "\n", "")

	// 初始化ssh连接
	initSshConnection(primaryHost, secondaryHost, joinerHost, sshUser, sshPass)
	primary, _ := primaryClient.Connect()
	secondary, _ := secondaryClient.Connect()
	joiner, _ := joinerClient.Connect()
	defer func() {
		primary.client.Close()
		secondary.client.Close()
		joiner.client.Close()
	}()

	// 获取mysql路径
	var result string
	wg.Add(1)
	go func() {
		scriptStr := fmt.Sprintf("string=`ls /data/mysqldata/` && array=(${string// /}) && echo ${array}")
		result, _ = primaryClient.Run(scriptStr)
		result = strings.TrimSpace(result)
		wg.Done()
	}()
	wg.Wait()

	// 格式化灾备集群用户名和密码信息
	args := strings.Split(targetUserInfo, ":")

	wg.Add(1)
	go func() {
		str := fmt.Sprintf("export LD_LIBRARY_PATH=/data/app/dbscale/libs && /data/app/dbscale/dbscale_binlog_tool "+
			"-u%s -p'%s' -h127.0.0.1 -P%s "+
			"--remote-user=%s --remote-password='%s' --remote-host=127.0.0.1 --remote-port=%s "+
			"--gtid-set=\"%s\" "+
			"-v  "+
			"--end-position=%s --end-file=/data/mysqldata/%s/dbdata/%s", args[0], args[1], primaryPort, args[0], args[1], primaryPort, resSet, strconv.Itoa(*masterStatus.Position), result, masterStatus.File)
		res, _ := primaryClient.Run(str)
		if res == "" {
			log.Println("闪回程序dbscale_binlog_tool出错")
			os.Exit(-1)
		}

		strCmd := fmt.Sprintf("/data/app/mysql-8.0.26/bin/mysql -u%s -p'%s' -h127.0.0.1 -P%s -e \"stop slave;reset master;reset slave;set global gtid_purged='%s';\"", args[0], args[1], primaryPort, resSet)
		res, _ = primaryClient.Run(strCmd)
		log.Println(res)
		wg.Done()
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		strCmd := fmt.Sprintf("/data/app/mysql-8.0.26/bin/mysql -u%s -p'%s' -h127.0.0.1 -P%s -e \"stop slave;reset master;reset slave;set global gtid_purged='%s';start slave;\"", args[0], args[1], secondaryPort, resSet)
		res, _ := secondaryClient.Run(strCmd)
		log.Println(res)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		strCmd := fmt.Sprintf("/data/app/mysql-8.0.26/bin/mysql -u%s -p'%s' -h127.0.0.1 -P%s -e \"stop slave;reset master;reset slave;set global gtid_purged='%s';start slave;\"", args[0], args[1], joinerPort, resSet)
		res, _ := joinerClient.Run(strCmd)
		log.Println(res)
		wg.Done()
	}()
	wg.Wait()

	// 2.6 重新构建主集群和备集群的复制关系
	socket := strings.Split(targetSocket, ":")
	fields := strings.Split(targetUserInfo, ":")
	AddBackupCluster(sourceUserInfo, socket[0], socket[1], fields[0], fields[1])

	//wg.Add(1)
	//go func() {
	//	strCmd := fmt.Sprintf("/data/app/mysql-8.0.26/bin/mysql -u%s -p'%s' -h127.0.0.1 -P%s -e \"start slave;\"", args[0], args[1], primaryPort)
	//	res, _ := primaryClient.Run(strCmd)
	//	log.Println(res)
	//	wg.Done()
	//}()
	//wg.Wait()
	StartSlave()
}
