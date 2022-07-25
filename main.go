package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func main() {
	c := cron.New()

	c.AddFunc("@every 60s", MysqlProxyAuthCronJob)
	c.AddFunc("@every 600s", MysqlProxyDBs)
	c.Start()

	r := gin.Default()
	r.GET("/auth", func(c *gin.Context) {
		AuthInit()
		c.JSON(http.StatusOK, gin.H{
			"message": "auth flush done",
		})
	})

	r.GET("/servers", func(c *gin.Context) {
		MysqlProxyDBs()
		c.JSON(http.StatusOK, gin.H{
			"message": "servers flush doen",
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	if err := Configinit(); err != nil {
		os.Exit(1)
	}

	log.Println("init config done")
	if err := DBinit(); err != nil {
		os.Exit(2)
	}
	log.Println("init db done")

	AuthInit()
	log.Println("init auth done")

	MysqlProxyDBs()
	log.Println("init mysql proxy servers done")
}

var localRunProxyServers = make(map[uint]int)

func MysqlProxyDBs() {
	var dbs []CoreDataSource
	db.Where(CoreDataSource{
		IsQuery: 1, // 0写 1读 2读写
	}).Find(&dbs)

	for _, db := range dbs {

		//检查是否已启动服务
		if _, o := localRunProxyServers[db.ID]; o {
			continue
		}

		//检查需要监听的端口配置
		if db.ProxyPort == 0 {
			continue
		}

		//检查是否存在重复端口
		for k, v := range localRunProxyServers {
			if v == db.ProxyPort {
				log.Printf("端口重复请检查,数据库ID: %d, %d", k, db.ID)
				continue
			}
		}

		//标记数据库ID，服务已启动
		localRunProxyServers[db.ID] = db.ProxyPort

		// 初始化&启动
		s := ProxyServer{
			Data: db,
		}
		go s.MysqlProxyServerCreate()
	}

}
