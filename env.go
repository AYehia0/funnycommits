// get the environment variables using viper

package main

import (
	"log"

	"github.com/spf13/viper"
)

// use viper to get the environment variables
func getEnv(key string) string {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}
	return viper.GetString(key)
}
