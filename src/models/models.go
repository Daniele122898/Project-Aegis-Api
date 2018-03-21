package models

type User struct {
	Id string `json:"id"`
	Token string `json:"token"`
	AvatarUrl string `json:"avatarUrl"`
	Created int `json:"created"`
	Username string `json:"username"`
	Discrim string `json:"discrim"`
}

type Guild struct{
	Id string `json:"id"`
	SecurityLevel int `json:"securityLevel"`
	ReportCount int `json:"reportCount"`
	Checked bool `json:"checked"`
}

type GuildLong struct {
	Id string `json:"id"`
	SecurityLevel int `json:"securityLevel"`
	ReportCount int `json:"reportCount"`
	Checked bool `json:"checked"`
	Reports []Report `json:"reports"`
}

type Report struct{
	Id int `json:"id"`
	GuildId string `json:"guildId"`
	UserId string `json:"userId"`
	Text string `json:"text"`
	Date int `json:"date"`
	Closed bool`json:"closed"`
}

type PostUpdateGuildSec struct{
	SecurityLevel int `json:"securityLevel"`
}

type PostReport struct{
	Reason string `json:"reason"`
}

type PostReportAdmin struct{
	Reason string `json:"reason"`
	UserId string `json:"userId"`
}

type GenerateTokenPost struct {
	Token string `json:"token"`
}

type NewTokenResponse struct {
	OldToken string `json:"oldToken"`
	NewToken string `json:"newToken"`
}

type GuildWeb struct{
	Id string `json:"id"`
	SecurityLevel int `json:"securityLevel"`
	ReportCount int `json:"reportCount"`
	Checked bool `json:"checked"`
	Reports []ReportWeb `json:"reports"`
}

type ReportWeb struct{
	Text string `json:"text"`
	Date int `json:"date"`
	Closed bool`json:"closed"`
	User GenUserDataPost `json:"user"`
}

type GenUserDataPost struct{
	Avatar string `json:"avatar"`
	Username string `json:"username"`
	Discrim string `json:"discrim"`
}

type LoginUserData struct{
	Username string `json:"username"`
	Discriminator string `json:"discriminator"`
	MfaEnabled bool `json:"mfa_enabled"`
	Id string `json:"id"`
	Avatar string `json:"avatar"`
}