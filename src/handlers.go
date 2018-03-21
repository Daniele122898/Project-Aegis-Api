package main

import (
	"github.com/gin-gonic/gin"
	"github.com/Daniele122898/Project-Aegis-Api/src/utils"
	"encoding/json"
	"log"
	"github.com/Daniele122898/Project-Aegis-Api/src/models"
	"net/http"
	"time"
	"strconv"
	"bytes"
	"io"
)

//------------POST-------------

func HandleSyncProfileAdmin(c *gin.Context){
	userId, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		log.Println("invalid userId")
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid User ID!"})
		return
	}

	defer c.Request.Body.Close()

	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, c.Request.Body)

	if err != nil{
		log.Println("Failed to copy to buffer, ",err)
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid JSON: "+err.Error()})
		return
	}
	var gotUser models.GenUserDataPost
	err = json.Unmarshal(buf.Bytes(), &gotUser)

	if err !=nil{
		log.Println("failed to unmarshall, ",err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wrong JSON Format: "+ err.Error()})
		return
	}

	//update
	err = UpdateUserInfo(userId, gotUser.Avatar, gotUser.Username, gotUser.Discrim)
	if err != nil{
		log.Println("failed to update userinfo, ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wrong JSON Format: "+ err.Error()})
		return
	}
	//success
	c.JSON(http.StatusOK, gin.H{"status":"Successfully updated userinfo"})
}

func HandleGenerateNewUserAdmin(c *gin.Context){
	userId, ok := utils.GetGuildId(&c.Params)
	if !ok {
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid User ID!"})
		return
	}

	defer c.Request.Body.Close()

	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, c.Request.Body)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid JSON: "+err.Error()})
		return
	}
	var user models.GenUserDataPost
	err = json.Unmarshal(buf.Bytes(), &user)
	//bind POST to JSON
	if err !=nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wrong JSON Format: "+ err.Error()})
		return
	}
	token := utils.GenerateNewToken()
	time := time.Now().Unix()
	err = CreateUser(userId, time, user.Avatar, token, user.Username, user.Discrim)
	if err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong when creating the user! He might already exist!"})
		return
	}
	//everything went well
	c.JSON(http.StatusOK, gin.H{"status":"User Created"})
}

func HandleGenerateNewToken(c *gin.Context){
	userId, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid User ID!"})
		return
	}

	defer c.Request.Body.Close()

	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, c.Request.Body)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid JSON: "+err.Error()})
		return
	}
	var genToken models.GenerateTokenPost
	err = json.Unmarshal(buf.Bytes(), &genToken)
	//bind POST to JSON
	if err !=nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wrong JSON Format: "+ err.Error()})
		return
	}
	//Check if token is valid
	user, err := GetUserByToken(genToken.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Token and UserId don't match!"})
		return
	}
	foundUserId, _ := utils.GetGuildIdString(user.Id)
	if foundUserId != userId {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Token and UserId don't match!"})
		return
	}
	//Token is valid so generate a new one
	newToken := utils.GenerateNewToken()

	//update token in DB
	err = UpdateUserToken(userId, newToken)
	if err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to generate new Token"})
		return
	}

	response := models.NewTokenResponse{OldToken:genToken.Token, NewToken:newToken}
	//Respond
	json.NewEncoder(c.Writer).Encode(&response)
}

func PostGuildReportAdmin(c *gin.Context){
	guildId, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID!"})
		return
	}
	//close the body at the end
	defer c.Request.Body.Close()

	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, c.Request.Body)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid JSON: "+err.Error()})
		return
	}
	var report models.PostReportAdmin
	err = json.Unmarshal(buf.Bytes(), &report)
	//bind POST to JSON
	if err !=nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wrong JSON Format: "+ err.Error()})
		return
	}
	//Do shit with available json
	//Get user that tried to post
	user, err := GetUserString(report.UserId)
	if err != nil{
		log.Println("Error in PostGuildReportAdmin, ",err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal Server Error"})
		return
	}
	//build actual report
	//Get time in unix
	date := time.Now().Unix()

	//get user INt64 id
	userIntId, err :=strconv.ParseInt(user.Id, 10, 64)
	if err!= nil {
		log.Println("Failed parse of UserId, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status":"Internal Server Error"})
		return
	}

	err = PostReport(userIntId, guildId, date, report.Reason, false)
	if err!= nil{
		log.Println("FAILED TO POST REPORT, ",err)
		c.JSON(http.StatusInternalServerError, gin.H{"status":"Internal Server Error"})
		return
	}
	//everything went well
	c.JSON(http.StatusOK, gin.H{"status":"Report Added or Updated"})
}

func PostGuildReport(c *gin.Context){
	guildId, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID!"})
		return
	}

	//validate guildId
	ok = utils.ValidIdInt(guildId)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID!"})
		return
	}

	//close the body at the end
	defer c.Request.Body.Close()

	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, c.Request.Body)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid JSON: "+err.Error()})
		return
	}
	var report models.PostReport
	err = json.Unmarshal(buf.Bytes(), &report)
	//bind POST to JSON
	if err !=nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wrong JSON Format: "+ err.Error()})
		return
	}
	//Do shit with available json
	//Get user that tried to post
	token := utils.GetToken(c)
	//get user
	user, err := GetUserByToken(token)
	if err != nil{
		log.Println("Error in PostGuildReport, ",err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal Server Error"})
		return
	}
	if report.Reason == ""{
		//Missing crucial information
		c.JSON(http.StatusBadRequest, gin.H{"error":"Bad Request: Provide GuildId and Reasoning"})
		return
	}
	//build actual report
	//Get time in unix
	date := time.Now().Unix()

	//get user INt64 id
	userIntId, err :=strconv.ParseInt(user.Id, 10, 64)
	if err!= nil {
		log.Println("Failed parse of UserId, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal Server Error"})
		return
	}

	err = PostReport(userIntId, guildId, date, report.Reason, false)
	if err!= nil{
		log.Println("FAILED TO POST REPORT, ",err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal Server Error"})
		return
	}
	//everything went well
	c.JSON(http.StatusOK, gin.H{"status":"Report Added or Updated"})
}

//------------------GET-----------------

func HandleDoesUserExistAdmin(c *gin.Context){
	userId, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid User ID!"})
		return
	}
	ok = DoesUserExist(userId)
	if ok {
		c.JSON(http.StatusOK, gin.H{"exists":true})
	} else {
		c.JSON(http.StatusOK, gin.H{"exists":false})
	}
}

func HandleGenerateAdminNewToken(c *gin.Context){
	userId, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid User ID!"})
		return
	}

	//Token is valid so generate a new one
	newToken := utils.GenerateNewToken()

	//update token in DB
	err := UpdateUserToken(userId, newToken)
	if err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to generate new Token"})
		return
	}

	log.Println("NEW TOKEN: ", newToken)

	response := models.NewTokenResponse{OldToken:"undefined", NewToken:newToken}
	log.Println(response)
	sendData, err:=json.Marshal(response)
	if err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Something went wrong"})
		return
	}
	log.Println(sendData, ", OR , ", string(sendData))
	//Respond
	//json.NewEncoder(c.Writer).Encode(&response)
	c.Writer.Write(sendData)

}

func GetGuildInfo (c *gin.Context){
	id, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID"})
		return
	}

	//validate guildId
	ok = utils.ValidIdInt(id)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID!"})
		return
	}

	guild, err:= GetGuild(id)
	if err != nil{
		//ID was not found or smth went wrong
		log.Println("Failed to get guild, ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error":"Guild wasn't found, "+err.Error()})
		return
	}
	//otherwise return json encoded guild Info
	jsn, err := json.Marshal(&guild)
	if err != nil{
		//marshall failed
		log.Println("Failed to marshall guild, ",err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal Server Error or Invalid JSON: "+err.Error()})
		return
	}
	//finally write json
	c.Writer.Write(jsn)
}

func HandleGetTokenAdmin(c *gin.Context){
	userId, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid User ID"})
		return
	}

	user, err := GetUser(userId)
	if err!=nil{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Couldn't find user!"})
		return
	}

	token := models.GenerateTokenPost{Token:user.Token}
	json.NewEncoder(c.Writer).Encode(token)

}

func GetGuildInfoLongAdmin(c *gin.Context){
	id, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID"})
		return
	}

	//validate guildId
	ok = utils.ValidIdInt(id)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID!"})
		return
	}

	guild, err := GetGuildWeb(id)
	if err != nil && guild ==nil{
		//ID was not found or smth went wrong
		log.Println("failed guild web completely, ",err)
		c.JSON(http.StatusBadRequest, gin.H{"error":"Guild wasn't found, "+err.Error()})
		return
	}
	//otherwise return json encoded guild Info
	jsn, err := json.Marshal(&guild)
	if err != nil{
		log.Println("failed marshal guild long, ",err)
		//marshall failed
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal Server Error or Invalid JSON: "+err.Error()})
		return
	}
	//finally write json
	c.Writer.Write(jsn)
}

func GetGuildInfoLong (c *gin.Context){
	id, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID"})
		return
	}

	//validate guildId
	ok = utils.ValidIdInt(id)
	if !ok{
		//user entered smth that is NOT an ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID!"})
		return
	}

	guild, err := GetGuildLong(id)
	if err != nil && guild ==nil{
		//ID was not found or smth went wrong
		log.Println("failed guild long completely, ",err)
		c.JSON(http.StatusBadRequest, gin.H{"error":"Guild wasn't found, "+err.Error()})
		return
	}
	//otherwise return json encoded guild Info
	jsn, err := json.Marshal(&guild)
	if err != nil{
		log.Println("failed marshal guild long, ",err)
		//marshall failed
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal Server Error or Invalid JSON: "+err.Error()})
		return
	}
	//finally write json
	c.Writer.Write(jsn)
}

//ADMIN STUFF

func GetGuildListAdmin(c *gin.Context){
	g, err := GetGuildList()
	if err != nil{
		log.Println("Failed to get Guild list, ",err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to create List"})
		return
	}
	json.NewEncoder(c.Writer).Encode(g)

}

func UpdateGuildSecurity(c *gin.Context){
	id, ok := utils.GetGuildId(&c.Params)
	if !ok{
		//incalid ID
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Guild ID"})
		return
	}
	//Get json
	//close the body at the end
	defer c.Request.Body.Close()

	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, c.Request.Body)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid JSON: "+err.Error()})
		return
	}
	var updateG models.PostUpdateGuildSec
	err = json.Unmarshal(buf.Bytes(), &updateG)
	//bind POST to JSON
	if err !=nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: "+ err.Error()})
		return
	}
	//Post new Security level
	err = UpdateGuildSecurityDb(id, updateG.SecurityLevel)
	if err!= nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Operation: "+ err.Error()})
		return
	}
	//everything went well
	c.JSON(http.StatusOK, gin.H{"status":"SecurityLevel Updated!"})
}

