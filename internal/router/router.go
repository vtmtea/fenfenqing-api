package router

import (
	"github.com/gin-gonic/gin"
	"github.com/vtmtea/fenfenqing-api/internal/handler"
	"github.com/vtmtea/fenfenqing-api/internal/middleware"
)

// SetupRouter 设置路由
func SetupRouter(
	authHandler *handler.AuthHandler,
	roomHandler *handler.RoomHandler,
	memberHandler *handler.MemberHandler,
	scoreHandler *handler.ScoreHandler,
) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API 路由组
	api := r.Group("/api")
	{
		// 认证相关（无需登录）
		api.POST("/auth/login", authHandler.Login)           // 微信登录
		api.GET("/auth/wechat", authHandler.Login)           // 兼容小程序码扫码
		api.POST("/auth/qrcode", authHandler.GenerateQRCode) // 生成小程序码

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(middleware.JWTAuth())
		{
			// 用户相关
			protected.GET("/user/info", authHandler.GetUserInfo)     // 获取用户信息
			protected.PUT("/user", authHandler.UpdateUserInfo)       // 更新用户信息

			// 房间相关
			rooms := protected.Group("/rooms")
			{
				rooms.GET("", roomHandler.GetRoomList)                    // 获取房间列表
				rooms.POST("", roomHandler.CreateRoom)                    // 创建房间
				rooms.GET("/:id", roomHandler.GetRoomByID)                // 根据 ID 获取房间
				rooms.DELETE("/:id", roomHandler.DeleteRoom)              // 删除房间
				rooms.POST("/:id/close", roomHandler.CloseRoom)           // 关闭房间
				rooms.POST("/:id/reopen", roomHandler.ReopenRoom)         // 重新打开房间

				// 成员相关
				rooms.GET("/:id/members", memberHandler.GetMemberList)    // 获取成员列表
				rooms.POST("/:id/members", memberHandler.AddMember)       // 添加成员
				rooms.DELETE("/:id/members/:memberID", memberHandler.DeleteMember) // 删除成员

				// 分数相关
				rooms.GET("/:id/scores", scoreHandler.GetScoreList)       // 获取分数记录
				rooms.POST("/:id/scores", scoreHandler.AddScore)          // 添加分数记录
				rooms.DELETE("/:id/scores/:scoreID", scoreHandler.DeleteScore) // 删除分数记录
			}

			// 根据房间号获取房间（无需认证，用于加入房间）
			api.GET("/rooms/roomId/:roomID", roomHandler.GetRoomByRoomID)
		}
	}

	// 404 处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "route not found"})
	})

	return r
}
