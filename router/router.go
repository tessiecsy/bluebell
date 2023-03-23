package router

import (
	"bluebell/controller"
	"bluebell/logger"
	"bluebell/middleware"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"

	"github.com/gin-gonic/gin"
)

func SetUpRouter(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) //gin设置成发布模式
	}

	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true), middleware.RateLimitMiddleware(2*time.Second, 1))

	//r.Use(logger.GinLogger(), logger.GinRecovery(true))

	/*	r.LoadHTMLFiles("./templates/index.html")
		r.Static("/static", "./static")
		r.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		}) */

	v1 := r.Group("/api/v1")

	// 注册
	v1.POST("/signup", controller.SignUpHandler)
	// 登录
	v1.POST("/login", controller.LoginHandler)

	//v1.GET("/posts2", controller.GetPostListHandler2)

	v1.Use(middleware.JWTAuthMiddleware()) //应用JWT认证中间件
	{
		v1.GET("/community", controller.CommunityHandler)
		v1.GET("/community/:id", controller.CommunityDetailHandler)

		v1.POST("/post", controller.CreatePostHandler)
		v1.GET("/post/:id", controller.GetPostDetailHandler)
		v1.GET("/posts", controller.GetPostListHandler)
		// 根据时间或分数获取帖子列表
		v1.GET("/posts2", controller.GetPostListHandler2)

		// 投票
		v1.POST("/vote", controller.PostVoteHandler)
	}

	pprof.Register(r) // 注册pprof相关路由
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
