package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/server"
	"github.com/siddontang/go/hack"
)

type ProxyServer struct {
	Listener net.Listener
	Data     CoreDataSource
}

func (s *ProxyServer) MysqlProxyServerCreate() (err error) {

	// 监听端口
	s.Listener, err = net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(s.Data.ProxyPort))
	if err != nil {
		log.Println(err)
		os.Exit(99)
		return
	}
	log.Printf("add db proxy %s %s , connect address %s, port %d", s.Data.Source, s.Data.IP, s.Data.ProxyIP, s.Data.ProxyPort)
	defer s.Listener.Close()

	// 获取到 TCP 请求，创建与客户端对接连接
	for {
		c, err := s.Listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.RunServer(c)
	}
}

func (s *ProxyServer) MysqlClientCreate() (mysqlClient *client.Conn, err error) {
	for {
		mysqlClient, err = client.Connect(s.Data.IP+":"+strconv.Itoa(s.Data.Port), s.Data.Username, Decrypt(s.Data.Password), "")
		if err != nil {
			log.Printf("连接至代理服务器错误: %s", err.Error())
			time.Sleep(time.Duration(10) * time.Second)
			continue
		}
		err = mysqlClient.Ping()
		if err != nil {
			log.Printf("Ping 服务器错误: %s", err.Error())
			continue
		}
		return
	}

}

func (s *ProxyServer) RunServer(c net.Conn) {

	// 异常恢复：旧连接重发包，造成内存错误
	defer func() {
		if e := recover(); e != nil {
			log.Printf("Panicing %s\r\n", e)
		}
	}()

	// 与客户端对接，分配 连接编号 & 认证
	var h server.Handler
	conn, err := server.NewCustomizedConn(c, server.NewDefaultServer(), mysqlProxyAuth, h)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("login auth access user: ", conn.GetUser(), " ip: ", c.RemoteAddr())

	// 认证通过后进行权限检查
	permissionsStatus := PermissionsCheck(conn.GetUser(), s.Data.Source)
	if permissionsStatus {
		log.Printf("login Permissions Check access user: %s, db: %s", conn.GetUser(), s.Data.Source)
	} else {
		log.Printf("login Permissions Check Fail user: %s, db: %s", conn.GetUser(), s.Data.Source)
		conn.WriteValue(fmt.Errorf("login Permissions Check Fail user: %s, db: %s", conn.GetUser(), s.Data.Source))
		conn.Close()
		return
	}

	// 对应每个 TCP 连接创建 mysql 代理连接
	mysqlClient, err := s.MysqlClientCreate()
	if err != nil {
		log.Println("创建 mysql 失败： ", err)
		return
	}

	// 处理 TCP 连接请求的 SQL 命令
	for {
		if err := s.GetCommand(conn, mysqlClient); err != nil {
			log.Println(err.Error())
			return
		}

	}
}

func (s *ProxyServer) GetCommand(c *server.Conn, mysqlClient *client.Conn) (err error) {
	if c.Conn == nil {
		return fmt.Errorf("connection closed")
	}

	data, err := c.ReadPacket()
	if err != nil {

		c.Close()
		c.Conn = nil
		return err
	}

	cmdtype := data[0]
	cmddata := hack.String(data[1:])

	switch cmdtype {
	// use database
	case 2:
		err = mysqlClient.UseDB(cmddata)
		if err != nil {
			err = c.WriteValue(err)
		} else {
			var v *mysql.Result
			err = c.WriteValue(v)
		}
	// 未知回调信息
	case 4:
		var v *mysql.Result
		err = c.WriteValue(v)

	// 执行SQL
	default:
		v, err := mysqlClient.Execute(cmddata)

		if err != nil {

			LogSave(s.Data.Source, c.RemoteAddr().String(), c.GetUser(), mysqlClient.GetDB(), cmddata, cmdtype, err)

			// 执行SQL 错误返回错误
			err = c.WriteValue(err)

		} else {

			// 执行SQL 返回数据
			err = c.WriteValue(v)

			// 转换执行SQL 返回数据为JOSN
			var data [][]interface{}
			if v != nil && v.Resultset != nil {
				v.GetInt(0, 0)

				for _, row := range v.Values {
					var tmpRow []interface{}
					for _, val := range row {

						switch val.Type {
						case mysql.FieldValueTypeString:
							tmpRow = append(tmpRow, string(val.AsString()))
						default:
							tmpRow = append(tmpRow, val.Value())
						}
					}
					data = append(data, tmpRow)
				}
			}

			LogSave(s.Data.Source, c.RemoteAddr().String(), c.GetUser(), mysqlClient.GetDB(), cmddata, cmdtype, data)

		}
	}

	if c.Conn != nil {
		c.ResetSequence()
	}

	if err != nil {
		c.Close()
		c.Conn = nil
	}
	return err
}
