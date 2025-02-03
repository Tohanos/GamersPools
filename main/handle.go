package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"game.com/pool/gamer"
	"game.com/pool/service"
	"github.com/gorilla/mux"
)

var (
	gpservice service.GPService
	storeInDB bool
)

func main() {

	var err error

	gpservice = service.NewGPService()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	router := mux.NewRouter()
	router.HandleFunc("/gamer", handlePostGamerRequest).Methods("POST")
	router.HandleFunc("/groups", handleGetGroupsRequest).Methods("GET")
	router.HandleFunc("/groups/reset", handleGetResetGroupsRequest).Methods("GET")
	router.HandleFunc("/groups/{number}", handleGetStatistics).Methods("GET")
	router.HandleFunc("/gamer/{name}", handleDeleteGamerRequest).Methods("DELETE")

	log.Printf("Запуск сервера на %s\n", addr)
	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatal(err)
	}
}

func handlePostGamerRequest(w http.ResponseWriter, r *http.Request) {
	data := make([]byte, r.ContentLength)
	r.Body.Read(data)
	var g gamer.Gamer
	log.Print("Входящий POST запрос на создание игрока: " + string(data))
	err := json.Unmarshal(data, &g)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		log.Print(err)
		return
	}

	fmt.Fprintf(w, "Полученные данные: %s\n", string(data))
	g.ConTime = time.Now()
	gpservice.AddGamer(g)

}

func handleGetGroupsRequest(w http.ResponseWriter, r *http.Request) {
	log.Print("Входящий запрос GET groups")
	groups := gpservice.GetGroups()
	v, err := json.Marshal(groups.Groups)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		log.Print(err)
		return
	}

	fmt.Fprintf(w, string(v))
}

func handleGetResetGroupsRequest(w http.ResponseWriter, r *http.Request) {
	log.Print("Входящий запрос GET reset groups")
	groups := gpservice.ResetGroups()
	v, err := json.Marshal(groups.Groups)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		log.Print(err)
		return
	}

	fmt.Fprintf(w, string(v))
}

func handleGetStatistics(w http.ResponseWriter, r *http.Request) {
	log.Print("Получен GET запрос для чтения статистики группы")
	vars := mux.Vars(r)
	pathVal := vars["number"]
	number, err := strconv.Atoi(pathVal)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		log.Print(err)
		return
	}
	statistics := gpservice.Gg.CalculateGroupStats(number)
	v, err := json.Marshal(statistics)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		log.Print(err)
		return
	}

	fmt.Fprintf(w, string(v))
}

func handleDeleteGamerRequest(w http.ResponseWriter, r *http.Request) {
	log.Print("Входящий запрос DELETE на удаление игрока")

	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("Не задано имя игрока")
		fmt.Fprint(w, "Не задано имя игрока")
		return
	}

	gamer, err := gpservice.Gp.Get(name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		fmt.Fprint(w, err)
		return
	}
	gpservice.DeleteGamer(*gamer)
}
