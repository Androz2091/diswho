package main

import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"

	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type Config struct {
	port string
	token string
}

var userCache = make(map[string]string)

func userRoute(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

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
		if res.Status != "200 OK" {
			http.Error(w, "Cannot fetch the Discord API", http.StatusInternalServerError)
		} else {
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
}

func inviteRoute(w http.ResponseWriter, r *http.Request) {

	code := mux.Vars(r)["code"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := &http.Client{}
	req, getErr := http.NewRequest("GET", "https://discord.com/api/v8/invites/" + code, nil)
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
	fmt.Fprintf(w, string(body))
	
}

func main() {

	viper.SetConfigFile("config.yml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	router := mux.NewRouter().StrictSlash(true)

	lmt := tollbooth.NewLimiter(11, nil)
	lmt.SetMessage("You have reached maximum request limit.")
	lmt.SetMessageContentType("text/plain; charset=utf-8")
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("A request was rejected")
	})
	lmt.SetIPLookups([]string{"X-Forwarded-For", "X-Real-IP"})

	router.Handle("/user/{id:[0-9]{16,32}}", tollbooth.LimitFuncHandler(lmt, userRoute)).Methods("GET")
	router.Handle("/invite/{code:[a-zA-Z0-9]+}", tollbooth.LimitFuncHandler(lmt, inviteRoute)).Methods("GET")
	log.Fatal(http.ListenAndServe(":" + viper.GetString("port"), router))
}
