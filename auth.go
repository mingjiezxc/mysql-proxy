package main

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/go-mysql-org/go-mysql/server"
)

var mysqlProxyAuth *server.InMemoryProvider
var localAccount = make(map[string]string)

func AuthInit() {
	mysqlProxyAuth = server.NewInMemoryProvider()

	var account []CoreAccount
	db.Find(&account)
	for _, a := range account {
		if a.ProxyAuth != "" {
			localAccount[a.Username] = a.ProxyAuth
			mysqlProxyAuth.AddUser(a.Username, Decrypt(a.ProxyAuth))
			// log.Printf("auth add user %s, password %s ", a.Username, Decrypt(a.ProxyAuth))
		}
	}

}

func MysqlProxyAuthCronJob() {

	var account []CoreAccount
	db.Find(&account)

	for _, a := range account {
		if a.ProxyAuth != "" {
			if localAccount[a.Username] == a.ProxyAuth {
				continue
			}
			mysqlProxyAuth.AddUser(a.Username, Decrypt(a.ProxyAuth))
			// log.Printf("auth add user %s, password %s ", a.Username, Decrypt(a.ProxyAuth))
		}
	}
}

func MysqlProxyAuthUpdate(username string) {
	var account CoreAccount

	db.Where(CoreDataSource{
		Username: username, // 0写 1读 2读写
	}).Find(&account)

	mysqlProxyAuth.AddUser(account.Username, Decrypt(account.ProxyAuth))

}

func PermissionsCheck(user string, dbSource string) bool {
	var userP CoreGrained
	db.Where(CoreGrained{Username: user}).First(&userP)
	var p []CoreRoleGroup
	var tmpGroup []string
	tmpGroup = append(tmpGroup, userP.Group...)

	db.Where("name IN (?)", tmpGroup).Find(&p)

	for _, v := range p {
		for _, q := range v.Permissions.QuerySource {
			if q == dbSource {
				return true
			}
		}
	}

	return false

}

type CoreGrained struct {
	ID       uint    `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Username string  `gorm:"type:varchar(50);not null;index:user_idx" json:"username"`
	Group    strArry `gorm:"type:json" json:"group"`
}

type CoreRoleGroup struct {
	ID          uint        `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Name        string      `gorm:"type:varchar(50);not null" json:"name"`
	Permissions Permissions `gorm:"type:json" json:"permissions"`
}

type strArry []string

func (p strArry) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *strArry) Scan(data interface{}) error {
	return json.Unmarshal(data.([]byte), &p)
}

type Permissions struct {
	Auditor     []string `json:"auditor"`
	DdlSource   []string `json:"ddl_source"`
	DmlSource   []string `json:"dml_source"`
	QuerySource []string `json:"query_source"`
}

func (p Permissions) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Permissions) Scan(data interface{}) error {
	return json.Unmarshal(data.([]byte), &p)
}
