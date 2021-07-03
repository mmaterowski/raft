package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-redis/redis"
)

var redisClient redis.Client

func ConnectToRedis(address string) bool {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	log.Print("Trying to connect")
	pong, err := client.Ping().Result()
	redisClient = *client
	return pong != "" && err == nil
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteToFile(id string, filename string, payload string) {
	data := []byte(payload)
	strings.Title(id)
	path := "persistence/" + strings.Title(id) + "/" + filename
	err := ioutil.WriteFile(path, data, 0644)
	check(err)
}
