package main

import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type Config struct {
	port string
	token string
}

var userCache = make(map[string]string)

func fetchUser(id string, w http.ResponseWriter) {

	cachedUser, cached := userCache[id]
	if cached {
		fmt.Fprintf(w, cachedUser)
	} else {
		fmt.Printf("Fetching user: %s\n", id)
		client := &http.Client{}
		req, getErr := http.NewRequest("GET", "https://discord.com/api/v8/users/" + id, nil)
		if getErr != nil {
			log.Fatal(getErr)
		}
		req.Header.Add("Authorization", "Bot " + viper.GetString("token"))
		res, getErr := client.Do(req)
		fmt.Printf("HTTP: %s\n", res.Status)
		if getErr != nil {
			log.Fatal(getErr)
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}
		userCache[id] = string(body)
		fmt.Fprintf(w, string(body))
	}
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fetchUser(id, w)
}

func main() {

	viper.SetConfigFile("config.yml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	router := mux.NewRouter().StrictSlash(true)
	router.Path("/user/{id:[0-9]{16,32}}").HandlerFunc(homeLink)
	log.Fatal(http.ListenAndServe(":" + viper.GetString("port"), router))
}
