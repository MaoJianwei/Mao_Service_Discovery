package MaoDatabase

import (
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	"MaoServerDiscovery/util"
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"time"
)

const (
	MODULE_NAME = "MYSQL-DB-SYNC-incubator"
	MYSQL_DB_TABLE_CREATE_SQL =
		"create table if not exists MaoServiceDiscovery (" +
			"Service_IP nvarchar(1024)," +
			"Report_IP nvarchar(1024)," +
			"Alive BOOLEAN," +
			"Detect_Count BIGINT," +
			"Report_Count BIGINT," +
			"Last_Seen nvarchar(1024)," +
			"Rtt_Duration nvarchar(1024)," +
			"RttOutbound_or_Remote_Timestamp nvarchar(1024)," +
			"Aux_Data nvarchar(1024)," +
			"primary key (Service_IP)" +
		");"
	MYSQL_DB_DATA_INSERT_SQL =
		"insert into MaoServiceDiscovery (" +
			"Service_IP, Report_IP, Alive," +
			"Detect_Count, Report_Count," +
			"Last_Seen, Rtt_Duration, RttOutbound_or_Remote_Timestamp," +
			"Aux_Data" +
		") values (?, ?, ?, ?, ?, ?, ?, ?, ?);"
	MYSQL_DB_TABLE_CLEAR_SQL = "delete from MaoServiceDiscovery"

	URL_MYSQL_HOMEPAGE = "/configMysql"
	URL_MYSQL_CONFIG   = "/addMysqlInfo"
	URL_MYSQL_SHOW   = "/getMysqlInfo"

	MYSQL_INFO_CONFIG_PATH = "/mysql"
)

type MysqlDataPublisher struct {

	username string
	password string
	ipDomainName string
	port uint16
	databaseName string

	// username:password@tcp(ipDomainName:port)/databaseName
	dataSourceName string


	dbConn *sql.DB
}



func (m *MysqlDataPublisher) initDatabaseTable() bool {

	dbTx, err := m.dbConn.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to create a transaction, %s", err.Error())
		return false
	}

	result, err := dbTx.ExecContext(context.Background(), MYSQL_DB_TABLE_CREATE_SQL)
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to create the table: %s, %s", "MaoServiceDiscovery", err.Error())
		dbTx.Rollback()
		return false
	}

	lastInsertId, err := result.LastInsertId()
	rowsAffected, err := result.RowsAffected()
	util.MaoLogM(util.INFO, MODULE_NAME, "Create the table OK, %d, %d rows", lastInsertId, rowsAffected)

	err = dbTx.Commit()
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to commit the transaction, %s", err.Error())
		dbTx.Rollback()
		return false
	}

	return true
}

func (m *MysqlDataPublisher) databaseInsertServices(dbTx *sql.Tx) error {
	_, err := dbTx.ExecContext(context.Background(), MYSQL_DB_TABLE_CLEAR_SQL)
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to clear the table: %s, %s", "MaoServiceDiscovery", err.Error())
		return err
	}

	grpcKaModule := MaoCommon.ServiceRegistryGetGrpcKaModule()
	if grpcKaModule == nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get GrpcKaModule")
	} else {
		serviceInfos := grpcKaModule.GetServiceInfo()
		for _, s := range serviceInfos {
			_, err := dbTx.ExecContext(context.Background(), MYSQL_DB_DATA_INSERT_SQL,
				s.Hostname, strings.Join(s.Ips, "\n"), s.Alive, 0, s.ReportTimes, s.LocalLastSeen, fmt.Sprintf("%.3fms", float64(s.RttDuration.Nanoseconds())/1000000), s.ServerDateTime, s.OtherData)
			if err != nil {
				return err
			}
		}
	}

	icmpKaModule := MaoCommon.ServiceRegistryGetIcmpKaModule()
	if icmpKaModule == nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get IcmpKaModule")
	} else {
		serviceInfos := icmpKaModule.GetServices()
		for _, s := range serviceInfos {
			_, err := dbTx.ExecContext(context.Background(), MYSQL_DB_DATA_INSERT_SQL,
				s.Address, "/", s.Alive, s.DetectCount, s.ReportCount, s.LastSeen, fmt.Sprintf("%.3fms", float64(s.RttDuration.Nanoseconds())/1000000), s.RttOutboundTimestamp, "/")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MysqlDataPublisher) databaseSyncLoop() {
	for {
		time.Sleep(1 * time.Second)
		if m.dbConn == nil {
			continue
		}

		dbTx, err := m.dbConn.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			util.MaoLogM(util.WARN, MODULE_NAME, "Fail to create a transaction, %s", err.Error())
			continue
		}

		err = m.databaseInsertServices(dbTx)
		if err != nil {
			util.MaoLogM(util.WARN, MODULE_NAME, "Fail to insert data to the table: %s, %s", "MaoServiceDiscovery", err.Error())
			dbTx.Rollback()
			continue
		}

		err = dbTx.Commit()
		if err != nil {
			util.MaoLogM(util.WARN, MODULE_NAME, "Fail to commit the transaction, %s", err.Error())
			dbTx.Rollback()
			continue
		}
	}
}

func (m *MysqlDataPublisher) configRestControlInterface() {
	restfulServer := MaoCommon.ServiceRegistryGetRestfulServerModule()
	if restfulServer == nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get RestfulServerModule, unable to register restful apis.")
		return
	}

	restfulServer.RegisterUiPage(URL_MYSQL_HOMEPAGE, m.showMysqlPage)
	restfulServer.RegisterGetApi(URL_MYSQL_SHOW, m.showMysqlInfo)
	restfulServer.RegisterPostApi(URL_MYSQL_CONFIG, m.processMysqlInfo)
}

func (m *MysqlDataPublisher) showMysqlPage(c *gin.Context) {
	c.HTML(200, "index-mysql.html", nil)
}

func (m *MysqlDataPublisher) showMysqlInfo(c *gin.Context) {
	data := make(map[string]interface{})
	data["username"] = m.username
	data["mysqlServerAddr"] = m.ipDomainName
	data["mysqlServerPort"] = m.port
	data["databaseName"] = m.databaseName

	// Attention: password can't be outputted !!!
	c.JSON(200, data)
}

func (m *MysqlDataPublisher) processMysqlInfo(c *gin.Context) {

	username, ok := c.GetPostForm("username")
	if !ok {
		c.String(200, "Fail to parse username.")
		return
	}

	password, ok := c.GetPostForm("password")
	if !ok {
		c.String(200, "Fail to parse password.")
		return
	}

	mysqlServerAddr, ok := c.GetPostForm("mysqlServerAddr")
	if !ok {
		c.String(200, "Fail to parse mysqlServerAddr.")
		return
	}

	mysqlServerPort, ok := c.GetPostForm("mysqlServerPort")
	var port64 uint64
	var err error
	if ok {
		port64, err = strconv.ParseUint(mysqlServerPort, 10, 16)
		if err != nil {
			util.MaoLogM(util.WARN, MODULE_NAME, "Fail to update mysql config, port number is error, %s", err.Error())
			c.String(200, "Fail to update mysql config, port number is error, %s", err.Error())
			return
		}
	}

	databaseName, ok := c.GetPostForm("databaseName")
	if !ok {
		c.String(200, "Fail to parse databaseName.")
		return
	}

	m.username = username
	m.password = password
	m.ipDomainName = mysqlServerAddr
	m.port = uint16(port64)
	m.databaseName = databaseName

	configModule := MaoCommon.ServiceRegistryGetConfigModule()
	if configModule == nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get config module instance, can't save mysql info")
	} else {
		data := make(map[string]interface{})
		data["username"] = m.username
		data["ipDomainName"] = m.ipDomainName
		data["port"] = m.port
		data["databaseName"] = m.databaseName

		// Attention: password can't be outputted !!!
		configModule.PutConfig(MYSQL_INFO_CONFIG_PATH, data)
	}

	if !m.reConstructMysqlConnection() {
		c.String(200, "Fail to re-construct MYSQL connection.")
	} else {
		m.showMysqlPage(c)
	}
}

func (m * MysqlDataPublisher) reConstructMysqlConnection() bool {
	m.dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", m.username, m.password, m.ipDomainName, m.port, m.databaseName)

	db, err := sql.Open("mysql", m.dataSourceName)
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to open database connection, %s", err.Error())
		return false
	}

	if m.dbConn != nil {
		err = m.dbConn.Close()
		if err != nil {
			util.MaoLogM(util.WARN, MODULE_NAME, "Fail to close previous database connection, %s", err.Error())
		}
	}
	m.dbConn = db

	m.dbConn.SetConnMaxLifetime(0)
	m.dbConn.SetMaxOpenConns(60)
	m.dbConn.SetMaxIdleConns(60)

	return m.initDatabaseTable()
}

func (m *MysqlDataPublisher) InitMysqlDataPublisher() bool {

	//todo: read MYSQL config from config file.
	//m.username = username
	//m.password = password
	//m.ipDomainName = ipDomainName
	//m.port = port
	//m.databaseName = databaseName

	m.reConstructMysqlConnection()

	go m.databaseSyncLoop()

	m.configRestControlInterface()

	return true
}
