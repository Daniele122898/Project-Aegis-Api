package main

import (
	"github.com/gin-gonic/gin"
)


func main(){

	//start DB
	openConnection()
	//close connection when this function quits
	defer closeConnection()

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()


	admin := router.Group("/api/admin/blacklist")

	admin.Use(AdminAuth())
	{
		guilds := admin.Group("/guild")
		{
			//POST
			guilds.POST("/updatesec/:id", UpdateGuildSecurity)
			guilds.POST("/report/:id", PostGuildReportAdmin)
			//GET
			guilds.GET("/infolong/:id", GetGuildInfoLongAdmin)
			guilds.GET("/list", GetGuildListAdmin)
		}

		users := admin.Group("/user")
		{
			//GET
			users.GET("/requestToken/:id", HandleGenerateAdminNewToken)
			users.GET("/exists/:id", HandleDoesUserExistAdmin)
			users.GET("/getToken/:id", HandleGetTokenAdmin)
			//POST
			users.POST("/createUser/:id", HandleGenerateNewUserAdmin)
			users.POST("/syncProfile/:id", HandleSyncProfileAdmin)
		}
	}

	authorized := router.Group("/api/blacklist")

	//use our Bearer Token Authentication
	authorized.Use(Auth())
	{
		guilds := authorized.Group("/guild")
		{
			//GET
			guilds.GET("/info/:id", GetGuildInfo)
			guilds.GET("/infolong/:id", GetGuildInfoLong)
			//POST
			guilds.POST("/report/:id", PostGuildReport)
		}

		users := authorized.Group("/user")
		{
			//POST
			users.POST("/requestToken/:id", HandleGenerateNewToken)
		}
	}

	//start router
	router.Run(":8200")
}

