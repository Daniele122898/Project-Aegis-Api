package utils

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"log"
	"strings"
	"github.com/pborman/uuid"
	"encoding/base64"
	"time"
)

const DISCORD_EPOCH int64 = 1420070400000

func GetToken(c *gin.Context) string{
	//Get Authorizationheader!
	auth := c.GetHeader("Authorization")
	if len(auth) <1 || auth == ""{
		return ""
	}
	//Check if bearer
	if !strings.HasPrefix(auth, "Bearer"){
		return ""
	}
	//split in 0 Bearer, 1 Token
	au := strings.Split(auth, " ")
	//if token is empty
	if len(au) <2 || len(au[1]) == 0{
		return ""
	}
	//return Token
	return au[1]
}

func GetGuildIdString(id string)(int64, bool){
	if id == ""{
		return 0, false
	} else{
		intId, err :=strconv.ParseInt(id, 10, 64)
		if err!= nil{
			log.Println("Failed parse of GuildId, ", err)
			return 0, false
		} else{
			return intId, true
		}
	}
}

func GetGuildId(params *gin.Params) (int64, bool){
	id := params.ByName("id")
	return GetGuildIdString(id)
}

func GenerateNewToken() string{
	id := uuid.New()
	str := base64.StdEncoding.EncodeToString([]byte(id))
	return str
}

func GetTimeFromSnowflake(id string) (time.Time, error){
	iid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return time.Now(), err
	}

	return time.Unix(((iid>>22)+DISCORD_EPOCH)/1000, 0).UTC(), nil
}


func GetTimeFromSnowflakeInt(id int64) (time.Time, error){
	return time.Unix(((id>>22)+DISCORD_EPOCH)/1000, 0).UTC(), nil
}

func ValidIdInt(id int64) bool{
	t, err := GetTimeFromSnowflakeInt(id)
	if err != nil{
		return false
	}
	if t.After(time.Now()){
		return false
	}
	if t.Unix() < DISCORD_EPOCH/1000{
		return false
	}
	return true
}

func ValidId(id string) bool{
	t, err := GetTimeFromSnowflake(id)
	if err != nil{
		return false
	}
	if t.After(time.Now()){
		return false
	}
	if t.Unix() < DISCORD_EPOCH/1000{
		return false
	}
	return true
}
