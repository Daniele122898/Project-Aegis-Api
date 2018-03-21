package config

import (
	"sync"
	"log"
	"io/ioutil"
	"encoding/json"
)

type configFile struct {
	DbConnection string `json:dbConnection`
}

const (
	CONFIG_FILE = "config.json"
)

var (
	config *configFile
	mutex = &sync.Mutex{}
)

func getConfig() configFile {
	mutex.Lock()
	defer mutex.Unlock()

	if config == nil { //check if config is "nil"
		//load config
		loadConfig()
	}

	return *config
}

func loadConfig(){
	log.Println("Loading Config...")
	raw, err := ioutil.ReadFile(CONFIG_FILE)
	if err != nil {
		log.Println("Couldn't find Config File!", err)
		panic(err)
	}
	json.Unmarshal(raw, &config)
}

func Get() configFile{
	return getConfig()
}
