package main

import (
	"github.com/gin-gonic/gin"
	"strings"
	"net/http"
	"time"
	"log"
	"github.com/Daniele122898/GuildBlackList/src/utils"
)

/*t := time.Now()

// Set example variable
c.Set("example", "12345")

// before request

c.Next()

// after request
latency := time.Since(t)
log.Print(latency)

// access the status we are sending
status := c.Writer.Status()
log.Println(status)*/

func Auth() gin.HandlerFunc{
	return func(c *gin.Context) {
		//Get Authorizationheader!
		auth := c.GetHeader("Authorization")
		if len(auth) <1 || auth == ""{
			//abort with 401 - Unauthorized!
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}
		//Check if bearer
		if !strings.HasPrefix(auth, "Bearer"){
			//abort with 401 - Unauthorized!
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}
		//split in 0 Bearer, 1 Token
		au := strings.Split(auth, " ")
		//if token is empty
		if len(au) <2 || len(au[1]) == 0{
			//abort with 401 - Unauthorized!
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}

		ok, _ := CheckAuthentication(au[1])
		if ok == false{
			//abort with 401 - Unauthorized!
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}
		log.Println("["+time.Now().Format("2006-01-02 15:04:05")+"] - Request Authenticated by: ", au[1])
		//Check ratelimit
		if utils.CheckIfRatelimited(au[1]){
			c.JSON(http.StatusTooManyRequests, gin.H{"error":"You are being Ratelimited."})
			log.Println("["+time.Now().Format("2006-01-02 15:04:05")+"] - User Ratelimited: ", au[1])
			c.Abort()
			return
		}
		if utils.InvokeRatelimit(au[1]){
			c.JSON(http.StatusTooManyRequests, gin.H{"error":"You are being Ratelimited"})
			log.Println("["+time.Now().Format("2006-01-02 15:04:05")+"] - User Ratelimited: ", au[1])
			c.Abort()
			return
		}
		c.Next()
	}
}

func AdminAuth()gin.HandlerFunc{
	return func(c *gin.Context) {
		//Get Authorizationheader!
		auth := c.GetHeader("Authorization")
		if len(auth) <1 || auth == ""{
			//abort with 401 - Unauthorized!
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}
		//Check if bearer
		if !strings.HasPrefix(auth, "Bearer"){
			//abort with 401 - Unauthorized!
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}
		//split in 0 Bearer, 1 Token
		au := strings.Split(auth, " ")
		//if token is empty
		if len(au) <2 || len(au[1]) == 0{
			//abort with 401 - Unauthorized!
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}
		ok, _ := CheckAdmin(au[1])
		if ok == false{
			//abort with 401 - Unauthorized!
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}
		log.Println("ADMIN Request Authenticated by: ", au[1])
		c.Next()
	}
}

