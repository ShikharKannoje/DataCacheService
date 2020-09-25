package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var cachehost = `http://` + os.Getenv("CACHE_DOMAIN") + "/" // localhost:8000
var apiGateserve = os.Getenv("API_GATE_SERVER")             //localhost:8080

// WriteJSONResponse represents a utility function which writes status code and JSON to response
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func startupServer() {
	r := mux.NewRouter()
	log.Println("The APIGatway is up on ", apiGateserve)
	r.HandleFunc("/", home)
	r.HandleFunc("/getEmployees/{id}", getEmployees).Methods("GET")
	r.HandleFunc("/alterEmployee", alterEmployee).Methods("PUT")
	r.HandleFunc("/relodeDataFromDB", reloadDataFromDB).Methods("GET")
	r.HandleFunc("/pageData/{from}/{to}", pageData).Methods("GET")

	log.Fatal(http.ListenAndServe(apiGateserve, r))
}

func home(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Hello, The server is running")
}

func pageData(w http.ResponseWriter, r *http.Request) {
	log.Println("ApiGatway api reached")
	vars := mux.Vars(r)
	to := vars["to"]
	from := vars["from"]
	TO, err := strconv.Atoi(to)
	if err != nil {
		WriteJSONResponse(w, 403, "Invalid Request")
		return
	}
	FROM, err := strconv.Atoi(from)
	if err != nil {
		WriteJSONResponse(w, 403, "Invalid Request")
		return
	}

	if FROM > TO {
		WriteJSONResponse(w, 403, "Invalid Request")
		return
	}
	check, err := http.Get(cachehost + "pageData/" + from + "/" + to)
	if err != nil {
		log.Println(err)
	}
	defer check.Body.Close()

	checkbody, err := ioutil.ReadAll(check.Body)
	if err != nil {
		log.Println(err)
	}

	log.Println(string(checkbody))
	if checkbody == nil {
		WriteJSONResponse(w, 404, "Not Found in DB")
		return
	}
	WriteJSONResponse(w, 200, strings.Trim(string(checkbody), "\n"))
	log.Println("Transaction over")
	return
}

func reloadDataFromDB(w http.ResponseWriter, r *http.Request) {
	log.Println("ApiGatway api reached")
	check, err := http.Get(cachehost + "RelodeDataFromDB")
	if err != nil {
		log.Println(err)
	}
	defer check.Body.Close()

	checkbody, err := ioutil.ReadAll(check.Body)
	if err != nil {
		log.Println(err)
	}

	log.Println(string(checkbody))
	if checkbody == nil {
		WriteJSONResponse(w, 404, "Not Found in DB")
		return
	}
	WriteJSONResponse(w, 200, strings.Trim(string(checkbody), "\n"))
	log.Println("Transaction over")
	return
}

type employeeStruct struct {
	Employeeid   string `json:"employeeid"`
	Employeename string `json:"employeename"`
}

func alterEmployee(w http.ResponseWriter, r *http.Request) {

	var emp employeeStruct
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		log.Println(err)
		WriteJSONResponse(w, 404, "Invalid Request")
		return
	}
	json, err := json.Marshal(emp)
	if err != nil {
		panic(err)
	}
	// initialize http client
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, cachehost+"alterEmployee", bytes.NewBuffer(json))
	if err != nil {
		log.Println(err)
		log.Println(err)
		WriteJSONResponse(w, 404, "Invalid Request")
		return
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	log.Println(req)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	checkbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	log.Println(string(checkbody))
	if checkbody == nil {
		WriteJSONResponse(w, 404, "Not Found in DB")
		return
	}
}

func getEmployees(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]
	_, err := strconv.Atoi(id)
	if err != nil {
		WriteJSONResponse(w, 403, "Invalid Request")
		return
	}
	check, err := http.Get(cachehost + "employeeDetails/" + id)
	if err != nil {
		log.Println(err)
	}
	defer check.Body.Close()

	checkbody, err := ioutil.ReadAll(check.Body)
	if err != nil {
		log.Println(err)
	}

	log.Println(checkbody)
	if checkbody == nil {
		WriteJSONResponse(w, 404, "Not Found in DB")
		return
	}
	WriteJSONResponse(w, 200, strings.Trim(string(checkbody), "\n"))
	log.Println("Transaction over")
	return
}

func main() {

	startupServer()

}
