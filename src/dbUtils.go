package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"github.com/Daniele122898/Project-Aegis-Api/src/models"
	"strconv"
	"github.com/Daniele122898/Project-Aegis-Api/src/utils"
	"github.com/Daniele122898/Project-Aegis-Api/src/config"
)

var (
	db *sql.DB
)

func closeConnection() {
	if db != nil{
		db.Close()
	}
}

func UpdateUserInfo(userId int64, avatarUrl, username, discrim string) error{
	//first check if its already up to date so we dont have to do any writing.
	user ,err:= GetUser(userId)
	if err!= nil{
		log.Println("Failed to get user in UpdateUserInfo, ", err)
		return err
	}
	//if identical no need to do any updating.
	if user.Username == username && user.Discrim == discrim && user.AvatarUrl == avatarUrl {
		return nil
	}
	//otherwise update the entry
	//stmt, err := db.Prepare("UPDATE Reports SET `Text`=?, `Date`=?, `Closed`=? WHERE UserId = ? AND GuildId =?")
	stmt, err := db.Prepare("UPDATE Users SET AvatarUrl=?, Username=?, Discrim=? WHERE Id=?")
	if err != nil {
		log.Println("Failed to prepare db statement for UpdateUserInfo, ", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(avatarUrl, username, discrim, userId)
	if err != nil{
		log.Println("Failed to execute db statement for UpdateUserInfo, ", err)
		return err
	}
	//everything went well so return nil
	return nil
}

func PostReport(userId, guildId, date int64, text string, closed bool) error{
	//check if entry already exists
	stmt, err := db.Prepare("SELECT COUNT(*) FROM Reports WHERE GuildId = ? AND UserId = ?")
	if err != nil {
		log.Println("Failed to prepare db statement for GetUserByToken, ", err)
		return err
	}
	defer stmt.Close()
	r := stmt.QueryRow(guildId, userId)
	var count int
	err = r.Scan(&count)
	if err != nil{
		log.Println("Failed to scan for count in PostReport")
		return err
	}
	if count ==0{
		//DOESNT EXIST YET
		return CreateReport(userId, guildId, date, text, closed)
	} else if count == 1 {
		//Exists once
		return UpdateReport(userId, guildId, date, text, closed)
	} else {
		//exists multiple times... we fucked up lol
		return &utils.ReportTextCharLimitError{Msg:"TEMPORARY ERROR!"}
	}
}

func UpdateUserToken(userId int64, newToken string) error{
	stmt, err := db.Prepare("UPDATE Users SET `Token`=? WHERE Id=?")
	if err != nil{
		log.Println("Failed to prepare statement for token Update, ",err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(newToken, userId)
	if err != nil{
		log.Println("Failed to update Token, ",err)
		return  err
	}
	//token was updated
	return nil
}

func UpdateReport(userId, guildId, date int64, text string, closed bool) error{
	//first lets get the TEXT of the report ticket that we already got to add the EDIT
	stm, err := db.Prepare("SELECT Text FROM Reports WHERE UserId = ? AND GuildId =?")
	if err != nil{
		log.Println("Failed to Get Previous Text, ",err)
		return err
	}
	defer stm.Close()
	r:= stm.QueryRow(userId, guildId)
	var Text string
	err = r.Scan(&Text)
	if err != nil{
		log.Println("Failed to scan for Text in Update Report, ",err)
		return err
	}
	//now that we have the old text combine it with the new Text
	Text = Text + " -- EDIT -- "+text
	//check if Text would now cross boundary of max LENGTH!
	if len(Text) > utils.MAX_REPORT_TEXT_LENGTH {
		log.Println("EXCEEDING CHAR LIMIT FOR TEXT!")
		//custom error
		return &utils.ReportTextCharLimitError{Msg:"Exceeded Char Limit for Report Text of 1000! Update declined"}
	}
	stmt, err := db.Prepare("UPDATE Reports SET `Text`=?, `Date`=?, `Closed`=? WHERE UserId=? AND GuildId=?")
	if err != nil{
		log.Println("Failed to create Report, ",err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(Text, date, closed, userId, guildId)
	if err != nil{
		log.Println("Failed to execute creation Report, ",err)
		return err
	}
	//row was updated so return nil
	//update checked in guild table
	st, err := db.Prepare("UPDATE Guilds SET Checked=0 WHERE Id = ?")
	if err != nil {
		log.Println("Failed to prepare db statement for updating guilds in Creating Report, ", err)
		return err
	}
	defer st.Close()
	_, err = st.Exec(guildId)
	if err != nil{
		log.Println("Failed to update guild in report, ",err)
		return err
	}
	return nil
}

func CreateUser(id, created int64, avatarUrl, token, username, discrim string) error{
	stmt, err := db.Prepare("INSERT INTO Users(Id, AvatarUrl, Created, Token, Username, Discrim) VALUES(?,?,?,?,?,?)")
	if err != nil{
		log.Println("Failed to prepare statement to add user, ", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, avatarUrl, created, token, username, discrim)
	if err != nil{
		log.Println("Failed to execute creation User, ",err)
		return err
	}
	//User created
	return nil
}

func CreateGuild(id int64, securityLevel, reportcount int, checked bool)error{
	stmt, err := db.Prepare("INSERT INTO Guilds(Id, SecurityLevel, Reports, Checked) VALUES(?,?,?,?)")
	if err != nil{
		log.Println("Failed to prepare statement to add guild, ", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, securityLevel, reportcount, checked)
	if err != nil{
		log.Println("Failed to execute creation Guild, ",err)
		return err
	}
	//guild created
	return nil
}

func CreateReport(userId, guildId, date int64, text string, closed bool) error{
	//IMPORTANT HERE is to UPDATE report count on the guild as well as Create a guild if it doesnt exists yet!
	stmt, err := db.Prepare("INSERT INTO Reports(`GuildId`, `UserId`, `Text`, `Date`, `Closed`) VALUES(?,?,?,?,?)")
	if err != nil{
		log.Println("Failed to create Report, ",err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(guildId, userId, text, date, closed)
	if err != nil{
		log.Println("Failed to execute creation Report, ",err)
		return err
	}
	//row was inserted and created at this point so return nil
	//since report was filed lets now check the associated guild
	stm, err := db.Prepare("SELECT COUNT(*) FROM Guilds WHERE Id = ?")
	if err != nil {
		log.Println("Failed to prepare db statement for getting guilds in Creating Report, ", err)
		return err
	}
	defer stm.Close()
	r := stm.QueryRow(guildId)
	var count int
	err = r.Scan(&count)
	if err != nil{
		log.Println("Failed to scan for count in create report for guilds")
		return err
	}
	if count == 0{
		//Guild doesnt exists yet. CREATE IT
		err = CreateGuild(guildId, utils.SecurityLevel_Good, 1, false)
	} else {
		//guilds exists UPDATE IT
		//add 1 to report count and checked to 0
		st, err := db.Prepare("UPDATE Guilds SET Reports = Reports+1, Checked=0 WHERE Id=?")
		if err != nil {
			log.Println("Failed to prepare db statement for updating guilds in Creating Report, ", err)
			return err
		}
		defer st.Close()
		_, err = st.Exec(guildId)
		if err != nil{
			log.Println("Failed to update guild in report, ",err)
			return err
		}
	}
	return nil
}

func GetUserByToken(token string) (*models.User, error){
	stmt, err := db.Prepare("SELECT Id,AvatarUrl, Created, Username, Discrim FROM Users WHERE Token = ?")
	if err != nil {
		log.Println("Failed to prepare db statement for GetUserByToken, ", err)
		return nil, err
	}
	defer stmt.Close()
	r := stmt.QueryRow(token)
	var Id string
	var AvatarUrl string
	var Created int
	var Username string
	var Discrim string
	err = r.Scan(&Id, &AvatarUrl, &Created, &Username, &Discrim)
	if r == nil || err != nil || Id == ""{
		return nil, err
	} else{
		user := models.User{Id: Id,AvatarUrl: AvatarUrl, Created: Created, Token: token, Username: Username, Discrim:Discrim}
		return &user , nil
	}
}

func CheckAuthentication(token string) (bool, error){
	stmt, err := db.Prepare("SELECT Id FROM Users WHERE Token = ?")
	if err != nil {
		log.Println("Failed to prepare db statement for CheckAuth, ", err)
		return false, err
	}
	defer stmt.Close()
	r := stmt.QueryRow(token)
	var Id string
	err = r.Scan(&Id)
	if r == nil || err != nil || Id == ""{
		return false, err
	} else{
		return true, nil
	}
}

func CheckAdmin(token string)(bool, error){
	//check if admin
	stmt, err := db.Prepare("SELECT COUNT(*) FROM Admins WHERE Token=?")
	if err!= nil{
		log.Println("Failed to prepare db statement for CheckAdmin, ",err)
		return false, err
	}
	defer stmt.Close()
	r := stmt.QueryRow(token)
	var count int
	err = r.Scan(&count)
	if err!= nil{
		log.Println("Failed to query count on CheckAdmin, ",err)
		return false, err
	} else if count == 0{
		return false, nil
	} else if count== 1{
		return true, nil
	} else{
		return false, nil
	}
}

func GetGuildLongString(guildId string) (*models.GuildLong, error){
	id, err :=strconv.ParseInt(guildId, 10, 64)
	if err!= nil{
		log.Println("Failed parse of GuildId, ", err)
		return nil, err
	}
	return GetGuildLong(id)
}

func GetGuildWeb(guildId int64)(*models.GuildWeb, error){
	guild, err := GetGuildLong(guildId)
	if err!=nil{
		log.Println("failed to get guildLong, ",err)
		return nil, err
	}

	var resp models.GuildWeb

	resp.Id = guild.Id
	resp.ReportCount = guild.ReportCount
	resp.SecurityLevel = guild.SecurityLevel
	resp.Checked = guild.Checked

	//REFACTOR REPORTS
	users := make(map[string]*models.GenUserDataPost)
	for _, el:= range guild.Reports {
		if rep, ok := users[el.UserId]; ok {
			resp.Reports = append(resp.Reports, models.ReportWeb{Text:el.Text, Date:el.Date, Closed:el.Closed, User:models.GenUserDataPost{Username:rep.Username, Discrim:rep.Discrim, Avatar:rep.Avatar}})
			continue
		} else {
			//Get user
			user ,err := GetUserString(el.UserId)
			if err!= nil{
				resp.Reports = append(resp.Reports, models.ReportWeb{Text:el.Text, Date:el.Date, Closed:el.Closed})
				continue
			}
			userData := models.GenUserDataPost{Username:user.Username, Avatar:user.AvatarUrl, Discrim:user.Discrim}
			//add to map
			users[user.Id] = &userData
			resp.Reports = append(resp.Reports, models.ReportWeb{Text:el.Text, Date:el.Date, Closed:el.Closed, User:userData})

		}
	}
	return &resp, nil
}

func GetGuildLong(guildId int64) (*models.GuildLong, error){
	//Get guild info
	stmt, err := db.Prepare("SELECT SecurityLevel, Reports, Checked FROM Guilds WHERE Id = ?")
	if err != nil {
		log.Println("Failed to prepare db statement for GetGuildLong, ", err)
		return nil, err
	}
	defer stmt.Close()
	var SecurityLevel int
	var Reports int
	var CheckedInt uint8
	err = stmt.QueryRow(guildId).Scan(&SecurityLevel, &Reports, &CheckedInt)
	if err != nil {
		log.Println("Failed to query row in GetGuildLong",err)
		return nil, err
	}
	id := strconv.FormatInt(guildId, 10)
	guild := models.GuildLong{Id:id, SecurityLevel:SecurityLevel, ReportCount: Reports, Checked:convertToBool(CheckedInt)}
	//Get Reports
	stmt, err = db.Prepare("SELECT `Id`, `UserId`, `Text`, `Date`, `Closed` FROM `Reports` WHERE `GuildId` = ?")
	if err != nil {
		log.Println("Failed to prepare db statement for GetGuildLong, ", err)
		return nil, err
	}
	rows, err := stmt.Query(guildId)
	defer rows.Close()
	if err != nil {
		log.Println("Failed to prepare db statement for GetGuildLong, ", err)
		return &guild, err
	}
	for rows.Next(){
		var Id int
		var UserId string
		var Text string
		var Date int
		var ClosedInt uint8
		err = rows.Scan(&Id, &UserId, &Text, &Date, &ClosedInt)
		if err != nil{
			log.Println("Failed to get reports, ", err)
			return &guild, err
		}
		guild.Reports = append(guild.Reports, models.Report{Id:Id, UserId:UserId, Text:Text, Date: Date, Closed:convertToBool(ClosedInt)})
	}

	return &guild, nil
}


func GetGuildList() ([]models.Guild, error){
	rows, err := db.Query("SELECT * FROM Guilds ORDER BY Checked ASC, Reports DESC")
	if err != nil {
		log.Println("Failed to query guild list, ", err)
		return nil, err
	}
	guilds := make([]models.Guild, 0)
	for rows.Next()  {
		var Id string
		var SecurityLevel int
		var Reports int
		var CheckedInt uint8
		err = rows.Scan(&Id, &SecurityLevel, &Reports, &CheckedInt)
		if err!= nil{
			log.Println("Failed to scan guild, ", err)
			continue
		}
		guilds = append(guilds, models.Guild{Id:Id,SecurityLevel:SecurityLevel, ReportCount:Reports, Checked:convertToBool(CheckedInt)})
	}
	return guilds, nil
}

func GetGuildString (guildId string) (*models.Guild, error){
	id, err :=strconv.ParseInt(guildId, 10, 64)
	if err!= nil{
		log.Println("Failed parse of GuildId, ", err)
		return nil, err
	}
	return GetGuild(id)
}

func GetGuild (guildId int64) (*models.Guild, error){
	stmt, err := db.Prepare("SELECT SecurityLevel, Reports FROM Guilds WHERE Id = ?")
	if err != nil {
		log.Println("Failed to prepare db statement for GetGuild, ", err)
		return nil, err
	}
	defer stmt.Close()
	var SecurityLevel int
	var Reports int
	err = stmt.QueryRow(guildId).Scan(&SecurityLevel, &Reports)
	if err != nil {
		log.Println("Failed to query row in GetGuild",err)
		return nil, err
	}
	id := strconv.FormatInt(guildId, 10)
	guild := models.Guild{Id:id, SecurityLevel:SecurityLevel, ReportCount: Reports}
	return &guild, nil
}

func GetUserString(userId string) (*models.User, error){
	id, err :=strconv.ParseInt(userId, 10, 64)
	if err!= nil{
		log.Println("Failed parse of UserID, ", err)
		return nil, err
	}
	return GetUser(id)
}

func DoesUserExist(userId int64) bool {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM Users WHERE Id=?")
	if err!= nil{
		log.Println("Failed to prepare db statement for DoesUserExist, ",err)
		return false
	}
	defer stmt.Close()
	r := stmt.QueryRow(userId)
	var count int
	err = r.Scan(&count)
	if err!= nil{
		log.Println("Failed to query count on DoesUserExist, ",err)
		return false
	} else if count == 0{
		return false
	} else{
		//if there is 1 or more (idk how but whatver) return true
		return true
	}
}

func GetUser(userID int64) (*models.User, error){
	//rows, err:= db.Query("SELECT `AvatarUrl`, `Created`, `Token` FROM `Users` WHERE `Id` =")
	stmt, err := db.Prepare("SELECT AvatarUrl, Created, Token, Username, Discrim FROM Users WHERE Id = ?")
	if err != nil {
		log.Println("Failed to prepare db statement for GetUser, ", err)
		return nil, err
	}
	defer stmt.Close()
	var AvatarUrl string
	var Created int
	var Token string
	var Username string
	var Discrim string
	err = stmt.QueryRow(userID).Scan(&AvatarUrl, &Created, &Token, &Username, &Discrim)
	if err != nil {
		log.Println("Failed to query row in GetUser",err)
		return nil, err
	}
	id := strconv.FormatInt(userID, 10)
	user := models.User{Id: id, AvatarUrl: AvatarUrl, Created: Created, Token: Token, Username:Username, Discrim:Discrim}
	return &user, nil
}

func getDB() *sql.DB{
	return db
}

//Admin shit
func UpdateGuildSecurityDb(guildId int64, SecurityLevel int) error{
	//First check if guild Exists,
	stmt, err := db.Prepare("SELECT COUNT(*) FROM Guilds WHERE Id=?")
	if err != nil{
		log.Println("Failed to prepare statement for guild seurity level update, ",err)
		return err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRow(guildId).Scan(&count)
	if err != nil{
		log.Println("Failed to get count of guild, ",err)
		return err
	}
	//Check count
	if count == 0{
		//create guild
		//"INSERT INTO Reports(`GuildId`, `UserId`, `Text`, `Date`, `Closed`) VALUES(?,?,?,?,?)")
		s1, err := db.Prepare("INSERT INTO Guilds(`Id`, `SecurityLevel`, `Reports`, `Checked`) VALUES(?,?,?,?)")
		if err != nil{
			log.Println("Failed to create stmt to create guild in security update, ",err)
			return err
		}
		_, err = s1.Exec(guildId, SecurityLevel, 0, 1)
		if err !=nil{
			log.Println("Failed to insert new guild in security update, ",err)
			return err
		}
		//inserted guild, no reports so no need to close any. we can return
		return nil
	} else {
		//update guild
		s2, err := db.Prepare("UPDATE Guilds SET Checked=1, SecurityLevel=? WHERE Id=?")
		if err != nil{
			log.Println("Failed to create stmt to update guild, ",err)
			return err
		}
		_, err = s2.Exec(SecurityLevel, guildId)
		if err != nil{
			log.Println("Failed to update guild security level, ",err)
			return err
		}
		//successfulyl updated guild Security level
		//Close all reports
		s3, err := db.Prepare("UPDATE Reports SET Closed=1 WHERE GuildId=?")
		if err != nil{
			log.Println("Failed to prepare stmt to close all reorts in security update, ",err)
			return err
		}
		_, err = s3.Exec(guildId)
		//whatever happens here. If we fail we still updated the guild so fuck it return nil
		if err!=nil {
			//log it tho. So i can investiage in case of failiures
			log.Println("Failed to close reports, ",err)
		}
		return nil
	}
}

func convertToBool(i uint8) bool {
	if i == 1 {
		return true
	} else {
		return false
	}
}

func openConnection(){
	var err error
	//open connection
	db, err = sql.Open("mysql", config.Get().DbConnection)
	//quit if DB is not found!
	if err!=nil{
		log.Fatal(err)
	}
}
