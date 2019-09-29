//接口限流和统计服务
//用于获取接口访问令牌以及统计各个接口调用的每日次数
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

var ch = make(chan string, 100000)

func main() {
	if err := Init(); err != nil {
		fmt.Println(err)
		return
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(Token())
	//http://localhost:6123/ratelimit/qianyaozu
	router.GET("/ratelimit/:name", ratelimit)
	router.GET("/ratelimit/getch", getch)
	router.GET("/ratelimit/getconfig", getconfig)
	router.GET("/ratelimit/getinfomation", getinfomation)
	router.POST("/ratelimit/editconfig", editconfig)
	router.Run(":6123")
}

//token认证
func Token() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("token") != "ratelimittoken:qqqqqqqqq" {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}

//获取接口每秒次数
func ratelimit(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.String(http.StatusOK, "-1")
		return
	}
	if TryTake(name) {
		c.String(http.StatusOK, "1")
		ch <- name
	} else {
		c.String(http.StatusOK, "0")
	}
}

//获取所有接口配置
func getconfig(c *gin.Context) {
	//todo  用于前台展示配置信息
}

//修改接口配置
func editconfig(c *gin.Context) {
	//todo  用于前台修改配置信息   从accessmap或者配置文件中获取相关信息
}

//获取所有接口的当日统计信息(或者指定日期)
func getinfomation(c *gin.Context) {
	//todo  用于前台获取当日所有接口的调用统计，以便前台生成统计报表，以及总调用次数 从redis中读取相关数据
}

//获取队列长度
func getch(c *gin.Context) {
	c.String(http.StatusOK, fmt.Sprint(len(ch)))
}
