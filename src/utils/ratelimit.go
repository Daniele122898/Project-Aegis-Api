package utils

import "time"

type UserRate struct {
	start int64
	counter uint8 //0-255
	limited bool
	limitedUntil int64
}

const(
	limitTime int64 = 20 		//ratelimit user for this amount
	deprecateTime int64 = 10 	//after this amount reset the ratelimit
	maxCounter uint8 = 5 		//max amount in time until ratelimit kicks in.
)

var(
	ratelimits = make(map[string]*UserRate)
)

func ResetUser(user *UserRate, token string) *UserRate {
	user.start = 0
	user.counter = 0
	user.limited = false
	user.limitedUntil = 0
	ratelimits[token] = user
	return user
}

//Fast check if user is ratelimited. If found and YES check if still valid
func CheckIfRatelimited(token string) bool {
	user, ok := ratelimits[token]
	if !ok {
		return false
	}
	if user.limited{
		//user is still limited
		if user.limitedUntil > time.Now().Unix(){
			return true
		}
		//reset user
		ResetUser(user, token)
		return false
	}
	//arbitrary return so go stfu <3
	return false
}

func InvokeRatelimit(token string) bool{
	user, ok := ratelimits[token]
	if !ok{
		//create user
		u := UserRate{start:time.Now().Unix(), counter: 1, limited: false, limitedUntil:0}
		ratelimits[token] = &u
		return false
	}
	if user.limited {
		// this should theoretically never happen but just to be save.
		return true
	}
	//check if deprecated
	if user.start < time.Now().Unix() - deprecateTime {
		//start anew
		user.start = time.Now().Unix()
		user.counter = 1
		ratelimits[token] = user
		return false
	}
	//its still in ratelimit time frame so check if i need to limit
	if user.counter <maxCounter{
		user.counter++
		ratelimits[token] = user
		return false
	}
	//user is in time and counter is 5 or greater. ratelimit
	user.counter++
	user.limited = true
	user.limitedUntil = time.Now().Unix() + limitTime
	ratelimits[token] = user
	return true
}
