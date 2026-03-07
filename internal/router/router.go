package router

import (
	"github.com/gin-gonic/gin"
	"github.com/vtmtea/fenfenqing-api/internal/handler"
)

// SetupRouter 设置路由
func SetupRouter(
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
		// 房间相关
		rooms := api.Group("/rooms")
		{
			rooms.GET("", roomHandler.GetRoomList)          // 获取房间列表
			rooms.POST("", roomHandler.CreateRoom)          // 创建房间
			rooms.GET("/:id", roomHandler.GetRoomByID)      // 根据 ID 获取房间
			rooms.GET("/roomId/:roomID", roomHandler.GetRoomByRoomID) // 根据房间号获取房间
			rooms.DELETE("/:id", roomHandler.DeleteRoom)    // 删除房间
		}

		// 成员相关
		rooms.GET("/:roomID/members", memberHandler.GetMemberList) // 获取成员列表
		rooms.POST("/:roomID/members", memberHandler.AddMember)    // 添加成员
		rooms.DELETE("/:roomID/members/:memberID", memberHandler.DeleteMember) // 删除成员

		// 分数相关
		rooms.GET("/:roomID/scores", scoreHandler.GetScoreList) // 获取分数记录
		rooms.POST("/:roomID/scores", scoreHandler.AddScore)    // 添加分数记录
		rooms.DELETE("/:roomID/scores/:scoreID", scoreHandler.DeleteScore) // 删除分数记录
	}

	// 404 处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "route not found"})
	})

	return r
}
