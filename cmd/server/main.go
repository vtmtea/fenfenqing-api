package main

import (
	"log"

	"github.com/vtmtea/fenfenqing-api/internal/config"
	"github.com/vtmtea/fenfenqing-api/internal/handler"
	"github.com/vtmtea/fenfenqing-api/internal/model"
	"github.com/vtmtea/fenfenqing-api/internal/router"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := model.InitDatabase(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 自动迁移数据库表
	if err := model.InitDB(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化处理器
	roomHandler := handler.NewRoomHandler(db)
	memberHandler := handler.NewMemberHandler(db)
	scoreHandler := handler.NewScoreHandler(db)

	// 设置路由
	r := router.SetupRouter(roomHandler, memberHandler, scoreHandler)

	// 启动服务器
	addr := ":" + cfg.Server.Port
	log.Printf("Starting server on %s (mode: %s)", addr, cfg.Server.Mode)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
