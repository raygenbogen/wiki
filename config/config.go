package config
import (
    "encoding/json"
    "os"
    "fmt"
 )


type Config struct{
	Domain string
}
func ReadConfig(section string)(string){
	file, err := os.Open("./config/wiki.cfg")
	defer file.Close()
	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}
	return config.Domain
}
