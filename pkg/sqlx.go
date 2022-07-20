package pkg

import (
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

//SetUpDatabase 数据库
func (b *BaseComponent) SetUpDatabase() {
	var (
		err  error
		conn *sqlx.DB
	)

	b.Cfg.DBLink = ProdLoadConfig(Mysql{User: b.Cfg.DB.Nodes[0].UserName, Pwd: b.Cfg.DB.Nodes[0].PassWord, Ip: b.Cfg.DB.Nodes[0].Server, Port: b.Cfg.DB.Nodes[0].Port, Schema: b.Cfg.DB.Nodes[0].SID}, b.Cfg.DBLinkCfg)

	dns := strings.Join([]string{b.Cfg.DBLink.MysqlWrite.User, ":", b.Cfg.DBLink.MysqlWrite.Pwd, "@tcp(", b.Cfg.DBLink.MysqlWrite.Ip + ":" + b.Cfg.DBLink.MysqlWrite.Port, ")/", b.Cfg.DBLink.MysqlWrite.Schema, "?charset=utf8mb4&parseTime=true&loc=Local"}, "")
	log.Printf("DB Instance initialize action dns:%s\n", dns)
	conn, err = sqlx.Open("mysql", dns)
	conn.SetConnMaxLifetime(time.Second * 10)
	conn.SetMaxOpenConns(b.Cfg.DB.MaxOpenConns)
	conn.SetMaxIdleConns(b.Cfg.DB.MaxIdleConns) // Ares note: default 2 -> now change to 10
	if err != nil {
		log.Panicf("Common Instance initialize error,errMsg='%s'", err.Error())

	}
	err = conn.Ping()
	if err != nil {
		log.Panicf("Common Instance initialize success,host_address=%s", b.Cfg.DB.Nodes[0].Server)
	}
	b.DB = conn
	// b.DB.Log = b.Log
}
