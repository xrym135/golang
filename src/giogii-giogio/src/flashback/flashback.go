package flashback

import (
	"fmt"
	"giogii/src/entity"
	"giogii/src/mapper"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var secondaryClient Client
var primaryClient Client
var joinerClient Client
var MasterSqlMapper mapper.SqlScaleOperator
var SlaveSqlMapper mapper.SqlScaleOperator
var SlaveStatus entity.SlaveStatus
var ServerName string
var Host string
var Port string
var MasterHost string
var MasterPort string

func initSshConnection(primary string, secondary string, joiner string, sshUser string, sshPass string) {
	primaryClient = Client{
		Username: sshUser,
		Password: sshPass,
		Socket:   fmt.Sprintf("%s:22", primary),
	}
	secondaryClient = Client{
		Username: sshUser,
		Password: sshPass,
		Socket:   fmt.Sprintf("%s:22", secondary),
	}
	joinerClient = Client{
		Username: sshUser,
		Password: sshPass,
		Socket:   fmt.Sprintf("%s:22", joiner),
	}
}

func InitMasterConnection(sourceUserInfo string, sourceSocket string) {
	s := mapper.InitSourceConn(sourceUserInfo, sourceSocket, "information_schema")
	MasterSqlMapper = &s
}

func InitTmpConnection(sourceUserInfo string, sourceSocket string) (s mapper.SqlStruct) {
	s = mapper.InitSourceConn(sourceUserInfo, sourceSocket, "information_schema")
	return
}

func InitSlaveConnection(targetUserInfo string, targetSocket string) {
	s := mapper.InitSourceConn(targetUserInfo, targetSocket, "information_schema")
	SlaveSqlMapper = &s
}

func GetSshIp() (p string, s string, j string) {
	strSql := fmt.Sprint("dbscale show dataservers")
	m := SlaveSqlMapper.DoQueryParseToDataServers(strSql)
	if len(m) > 0 {
		for i := 0; i < len(m); i++ {
			switch ms := m[i].MasterOnlineStatus.String; ms {
			case "Master_Online":
				p = m[i].Host.String
				MasterHost = m[i].Host.String
				MasterPort = m[i].Port.String
			default:
				if s == "" {
					s = m[i].Host.String
					ServerName = m[i].Servername.String
					Host = m[i].Host.String
					Port = m[i].Port.String
				} else {
					j = m[i].Host.String
				}
			}
		}
	}
	return p, s, j
}

func RemoveSlaveCluster() {
	var strSql string
	strSql = fmt.Sprint("dbscale dynamic remove datasource slave_dbscale_source")
	MasterSqlMapper.DoQueryWithoutRes(strSql)
	strSql = fmt.Sprint("dbscale dynamic remove dataserver slave_dbscale_server")
	MasterSqlMapper.DoQueryWithoutRes(strSql)
}

func AddBackupCluster(sourceUserInfo string, host string, port string, user string, password string) {
	var strSql string
	var id string
	strSql = fmt.Sprintf("dbscale request cluster info")
	info := MasterSqlMapper.DoQueryParseToClusterInfo(strSql)
	if len(info) > 0 {
		for i := 0; i < len(info); i++ {
			if info[i].MasterDbscale == "master" {
				tmpConnection := InitTmpConnection(sourceUserInfo, info[i].Host)
				strSql = fmt.Sprintf("dbscale request next group id")
				id = tmpConnection.DoQueryParseSingleValue(strSql)
				tmpConnection.DoClose()
			}
		}
	}

	strSql = fmt.Sprintf("dbscale dynamic ADD DATASERVER server_name=slave_dbscale_server,server_host=\"%s\",server_port=%s,server_user=\"%s\",server_password=\"%s\",dbscale_server", host, port, user, password)
	MasterSqlMapper.DoQueryWithoutRes(strSql)

	strSql = fmt.Sprintf("dbscale dynamic add server datasource slave_dbscale_source slave_dbscale_server-1-1000-400-800 group_id = %s", id)
	MasterSqlMapper.DoQueryWithoutRes(strSql)

	strSql = fmt.Sprintf("dbscale dynamic add slave slave_dbscale_source to normal_0")
	MasterSqlMapper.DoQueryWithoutRes(strSql)
}

func StartSlave() {
	var strSql string
	strSql = fmt.Sprint("dbscale set global 'enable-slave-dbscale-server'=1")
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
	strSql = fmt.Sprint("dbscale set global 'slave-dbscale-mode'=1")
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
	strSql = fmt.Sprint("start slave")
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
}

func AddData() {
	var strSql string
	strSql = fmt.Sprintf("create database a")
	MasterSqlMapper.DoQueryWithoutRes(strSql)
	strSql = fmt.Sprintf("drop database a")
	MasterSqlMapper.DoQueryWithoutRes(strSql)
}

func CloseReplication() {
	var strSql string
	strSql = fmt.Sprint("stop slave")
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
	strSql = fmt.Sprint("dbscale set global 'slave-dbscale-mode'=0")
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
	strSql = fmt.Sprint("dbscale set global 'enable-slave-dbscale-server'=0")
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
}

func GetSlaveGTIDSet() {
	var strSql string
	strSql = fmt.Sprint("show slave status")
	SlaveStatus = SlaveSqlMapper.DoQueryParseSlave(strSql)
}

func DisableDataServer() {
	var strSql string
	strSql = fmt.Sprintf("dbscale disable dataserver %s", ServerName)
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
	log.Println("剔除孤岛节点: ", ServerName)
}
func EnableDataServer() {
	var strSql string
	strSql = fmt.Sprintf("dbscale enable dataserver %s", ServerName)
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
}

func CloseReadOnly() {
	var strSql string
	strSql = fmt.Sprint("dbscale set global \"enable-read-only\" = 0")
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
}

func EnableReadOnly() {
	var strSql string
	strSql = fmt.Sprint("dbscale set global \"enable-read-only\" = 1")
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
}

func ForceOnline() {
	var strSql string
	strSql = fmt.Sprintf("DBSCALE FLASHBACK DATASERVER %s FORCE ONLINE", ServerName)
	SlaveSqlMapper.DoQueryWithoutRes(strSql)
}

var (
	wg sync.WaitGroup
)

func DoStartFlashback(targetUserInfo string, targetSocket string, sshUser string, sshPass string) {
	defer func() {
		SlaveSqlMapper.DoClose()
		MasterSqlMapper.DoClose()
	}()
	/**
	主集群移除灾备集群、灾备集群端口主集群，可以幂等操作
	*/
	log.Println("准备主集群移除灾备集群")
	RemoveSlaveCluster()
	log.Println("主集群移除灾备集群完成")
	log.Println("准备备集群关闭复制功能")
	CloseReplication()
	log.Println("备集群关闭复制功能完成")

	/**
	获取灾备集群的GTID，确保灾备集群的数据全部回放完成
	*/
	for {
		GetSlaveGTIDSet()
		if SlaveStatus.SecondsBehindMaster.Int64 == 0 {
			log.Println(fmt.Sprintf("记录gtid [ %s ]", SlaveStatus.ExecutedGtidSet))
			break
		}
		time.Sleep(3 * time.Second)
	}
	/**
	移除节点
	*/
	p, s, j := GetSshIp()
	log.Println("准备备集群剔除孤岛节点")
	DisableDataServer()
	log.Println("备集群剔除孤岛节点完成")

	log.Println("准备备集群关闭只读功能")
	CloseReadOnly()
	log.Println("备集群关闭只读功能完成")
	log.Println("********************************************************************************************")
	log.Println("*********************************备集群可以进行业务写入操作*********************************")
	log.Println("********************************************************************************************")

	initSshConnection(s, j, p, sshUser, sshPass)
	primary, _ := primaryClient.Connect()
	secondary, _ := secondaryClient.Connect()
	joiner, _ := joinerClient.Connect()
	defer func() {
		primary.client.Close()
		secondary.client.Close()
		joiner.client.Close()
	}()

	/**
	灾备集群孤岛节点安装clone插件，clone user 授权
	*/
	var scriptPath = getCurrentAbPath()
	wg.Add(1)
	go func(client *ssh.Client) {
		log.Println(fmt.Sprintf("准备孤岛节点:%s上传clone脚本", s))
		primaryClient.UploadFile(scriptPath+"/installClonePlugin.sh", "/home/mysql/installClonePlugin.sh", client)
		result, _ := primaryClient.Run("chmod 755 *")
		log.Println(result)
		log.Println("孤岛节点上传clone脚本完成")
		wg.Done()
	}(primaryClient.client)

	wg.Add(1)
	go func(client *ssh.Client) {
		log.Println(fmt.Sprintf("准备在%s节点上传initInstance/clone/check脚本", j))
		secondaryClient.UploadFile(scriptPath+"/initInstance.sh", "/home/mysql/initInstance.sh", client)
		secondaryClient.UploadFile(scriptPath+"/clone.sh", "/home/mysql/clone.sh", client)
		secondaryClient.UploadFile(scriptPath+"/check.sh", "/home/mysql/check.sh", client)
		result, _ := secondaryClient.Run("chmod 755 *")
		log.Println(result)
		log.Println(fmt.Sprintf("%s节点上传initInstance/clone/check脚本完成", j))
		wg.Done()
	}(secondaryClient.client)

	wg.Add(1)
	go func(client *ssh.Client) {
		log.Println(fmt.Sprintf("准备在%s节点上传initInstance/clone/check脚本", p))
		joinerClient.UploadFile(scriptPath+"/initInstance.sh", "/home/mysql/initInstance.sh", client)
		joinerClient.UploadFile(scriptPath+"/clone.sh", "/home/mysql/clone.sh", client)
		joinerClient.UploadFile(scriptPath+"/check.sh", "/home/mysql/check.sh", client)
		result, _ := joinerClient.Run("chmod 755 *")
		log.Println(result)
		log.Println(fmt.Sprintf("%s节点上传initInstance/clone/check脚本完成", p))
		wg.Done()
	}(joinerClient.client)
	wg.Wait()

	/**
	执行插件安装
	*/
	wg.Add(1)
	go func(client *ssh.Client) {
		log.Println("准备孤岛节点安装clone插件")
		result, _ := primaryClient.Run("bash /home/mysql/installClonePlugin.sh")
		log.Println(result)
		log.Println("孤岛节点安装clone插件完成")
		wg.Done()
	}(primaryClient.client)
	wg.Wait()

	/**
	执行clone动作
	*/
	wg.Add(1)
	go func(client *ssh.Client) {
		log.Println("准备初始化第一个clone实例")
		secondaryClient.Run("bash /home/mysql/initInstance.sh")
		log.Println("初始化第一个clone实例完成")

		time.Sleep(5 * time.Second)

		log.Println("准备执行第一个clone命令")
		fields := strings.Split(targetUserInfo, ":")
		scriptStr := fmt.Sprintf("bash /home/mysql/clone.sh %s %s %s %s", fields[0], fields[1], Host, Port)
		result, _ := secondaryClient.Run(scriptStr)
		log.Println(result)
		log.Println("执行第一个clone命令完成")
		wg.Done()
	}(secondaryClient.client)
	wg.Wait()

	wg.Add(1)
	go func(client *ssh.Client) {

		log.Println("准备初始化第二个clone实例")
		joinerClient.Run("bash /home/mysql/initInstance.sh")
		log.Println("初始化第二个clone实例完成")

		time.Sleep(5 * time.Second)

		log.Println("准备执行第二个clone命令")
		fields := strings.Split(targetUserInfo, ":")
		scriptStr := fmt.Sprintf("bash /home/mysql/clone.sh %s %s %s %s", fields[0], fields[1], Host, Port)
		result, _ := joinerClient.Run(scriptStr)
		log.Println(result)
		log.Println("执行第二个clone命令完成")
		wg.Done()
	}(joinerClient.client)
	wg.Wait()

}

func DoStopFlashback(sourceUserInfo string, targetUserInfo string, targetSocket string, sshUser string, sshPass string) {

	p, s, j := GetSshIp()

	initSshConnection(s, j, p, sshUser, sshPass)
	primary, _ := primaryClient.Connect()
	secondary, _ := secondaryClient.Connect()
	joiner, _ := joinerClient.Connect()
	defer func() {
		primary.client.Close()
		secondary.client.Close()
		joiner.client.Close()
	}()

	log.Println("准备备集群打开只读功能")
	EnableReadOnly()
	log.Println("备集群关闭只读功能完成")

	wg.Add(1)
	go func(client *ssh.Client) {
		log.Println("准备还原第一个clone实例")
		fields := strings.Split(targetUserInfo, ":")
		scriptStr := fmt.Sprintf("bash /home/mysql/check.sh %s %s", fields[0], fields[1])
		result, _ := secondaryClient.Run(scriptStr)
		log.Println(result)
		log.Println("还原第一个clone实例完成")
		wg.Done()
	}(secondaryClient.client)

	wg.Add(1)
	go func(client *ssh.Client) {
		log.Println("准备还原第二个clone实例")
		fields := strings.Split(targetUserInfo, ":")
		scriptStr := fmt.Sprintf("bash /home/mysql/check.sh %s %s", fields[0], fields[1])
		result, _ := joinerClient.Run(scriptStr)
		log.Println(result)
		log.Println("还原第二个clone实例完成")
		wg.Done()
	}(joinerClient.client)
	wg.Wait()

	wg.Add(1)
	go func() {

		EnableReadOnly()
		time.Sleep(20 * time.Second)
		EnableReadOnly()

		time.Sleep(20 * time.Second)
		EnableReadOnly()

		time.Sleep(3 * time.Second)
		EnableDataServer()

		time.Sleep(3 * time.Second)
		ForceOnline()

		time.Sleep(3 * time.Second)
		EnableReadOnly()

		log.Println("准备修复flashback")
		socket := strings.Split(targetSocket, ":")
		fields := strings.Split(targetUserInfo, ":")
		scriptStr := fmt.Sprintf("/data/app/mysql-8.0.26/bin/mysql -u%s -p%s -h%s -P%s -e \"stop slave;reset slave all;\"", fields[0], fields[1], MasterHost, MasterPort)
		result, _ := primaryClient.Run(scriptStr)
		log.Println(result)
		log.Println("修复flashback完成")

		AddData()
		time.Sleep(5 * time.Second)
		AddBackupCluster(sourceUserInfo, socket[0], socket[1], fields[0], fields[1])
		time.Sleep(2 * time.Second)
		StartSlave()
		wg.Done()
	}()
	wg.Wait()

	defer func() {
		SlaveSqlMapper.DoClose()
		MasterSqlMapper.DoClose()
	}()

}

func getCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}
