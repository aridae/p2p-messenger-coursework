package appconfig

import (
	"os"
	"os/user"
)

type ClientOptions struct {
	Port     int
	Host     string
	Username string
}

const (
	PORT = 35035
)

func GetClientOptions() *ClientOptions {
	hostName, _ := os.Hostname()
	currentUser, _ := user.Current()
	return &ClientOptions{
		Host:     hostName,
		Port:     PORT,
		Username: currentUser.Name + "@" + hostName,
	}
}
