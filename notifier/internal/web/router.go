package web

import (
	"delayednotifier/internal/service"
	"delayednotifier/internal/web/handlers"

	"github.com/gin-gonic/gin"
	f "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/wb-go/wbf/ginext"
)

func SetRoutes(router *ginext.Engine, s *service.Service) {
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Static("/static", "./templates/static")

	router.GET(
		"/swagger/*any", ginSwagger.WrapHandler(f.Handler),
	)

	router.GET("/", handlers.Main(s))

	router.POST("/notify", handlers.CreateNotify(s))
	router.GET("/notify/:id", handlers.GetNotify(s))
	router.PATCH("/notify/:id", handlers.UpdateNotify(s))
	router.DELETE("/notify/:id", handlers.DeleteNotify(s))
}
