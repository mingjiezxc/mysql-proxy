package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

type CoreAccount struct {
	ID         uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Username   string `gorm:"type:varchar(50);not null;index:user_idx" json:"username"`
	Password   string `gorm:"type:varchar(150);not null" json:"password"`
	Rule       string `gorm:"type:varchar(10);not null" json:"rule"`
	Department string `gorm:"type:varchar(50);" json:"department"`
	RealName   string `gorm:"type:varchar(50);" json:"real_name"`
	Email      string `gorm:"type:varchar(50);" json:"email"`
	ProxyAuth  string `gorm:"type:varchar(150);not null" json:"proxy_auth"`
}

type CoreDataSource struct {
	ID        uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	IDC       string `gorm:"type:varchar(50);not null" json:"idc"`
	Source    string `gorm:"type:varchar(50);not null" json:"source"`
	IP        string `gorm:"type:varchar(200);not null" json:"ip"`
	Port      int    `gorm:"type:int(10);not null" json:"port"`
	Username  string `gorm:"type:varchar(50);not null" json:"username"`
	Password  string `gorm:"type:varchar(150);not null" json:"password"`
	IsQuery   int    `gorm:"type:tinyint(2);not null" json:"is_query"` // 0写 1读 2读写
	ProxyIP   string `gorm:"type:varchar(200);not null" json:"proxy_ip"`
	ProxyPort int    `gorm:"type:int(10);not null" json:"proxy_port"`
}

type OpsMysqlProxyLog struct {
	ID         uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	RemoteAddr string `gorm:"type:varchar(50);not null" json:"remote_addr"`
	UserName   string `gorm:"type:varchar(50);index:idx_username;not null" json:"username"`
	Sql        string `gorm:"type:varchar(5000);not null" json:"sql"`
	Type       byte   `gorm:"type:varchar(50);not null" json:"type"`
	Source     string `gorm:"type:varchar(50);not null" json:"source"`
	Database   string `gorm:"type:varchar(100);not null" json:"database"`
	Data       string `gorm:"type:longtext;not null" json:"data"`
	CreatedAt  time.Time
}

func DBinit() (err error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Mysql.User,
		config.Mysql.Password,
		config.Mysql.Host,
		config.Mysql.Port,
		config.Mysql.Db,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	db.AutoMigrate(
		&CoreAccount{},
		&CoreDataSource{},
		&OpsMysqlProxyLog{},
	)

	return
}

func LogSave(source string, remoteAddr string, user string, database string, sql string, sqltype byte, data interface{}) {
	var result string

	switch v := data.(type) {
	case error:
		result = v.Error()
	case [][]interface{}:
		tmpByte, _ := json.Marshal(v)
		result = string(tmpByte)
	default:
		result = fmt.Sprintf("%#v", v)
	}

	l := OpsMysqlProxyLog{
		UserName:   user,
		RemoteAddr: remoteAddr,
		Sql:        sql,
		Type:       sqltype,
		Data:       result,
		Source:     source,
		Database:   database,
	}

	res := db.Create(&l)
	if res.Error != nil {
		log.Println("save log err: ", res.Error.Error())
	}

}
