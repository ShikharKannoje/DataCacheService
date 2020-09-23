package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

const SIZE = 100

const (
	hostname     = "localhost"
	hostport     = 5432
	username     = "postgres"
	password     = "root"
	databasename = "employee"
)

var conStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", hostname, hostport, username, password, databasename)

type Node struct {
	employeeid   string
	employeename string
	Left         *Node
	Right        *Node
}

// double linked list
type Queue struct {
	Head   *Node
	Tail   *Node
	Length int
}

// maps string to node in Queue
type Hash map[string]*Node

type Cache struct {
	Queue Queue
	Hash  Hash
	sync.Mutex
}

func NewCache() Cache {
	return Cache{Queue: NewQueue(), Hash: Hash{}}
}

func NewQueue() Queue {
	head := &Node{}
	tail := &Node{}
	head.Right = tail
	tail.Left = head

	return Queue{Head: head, Tail: tail}
}

func (c *Cache) IfHit(str string) (n Node, b bool) {
	node := &Node{}
	fmt.Println(c.Hash)
	if val, ok := c.Hash[str]; ok {
		fmt.Println("Cache HIT", val.employeename, val.employeeid, ok)
		node = c.Remove(val)
		c.Add(node)
		c.Hash[str] = node
		return *node, true
	}
	return *node, false
}

func (c *Cache) Remove(n *Node) *Node {
	fmt.Printf("remove: %s %s\n", n.employeeid, n.employeename)
	left := n.Left
	right := n.Right
	left.Right = right
	right.Left = left
	c.Queue.Length -= 1

	delete(c.Hash, n.employeeid)
	return n
}

func (c *Cache) Update(n *Node) {
	log.Println("Updating ", n.employeeid, n.employeename)
	c.Hash[n.employeeid] = n

}

func (c *Cache) Add(n *Node) {
	fmt.Printf("add: %s %s\n", n.employeeid, n.employeename)
	tmp := c.Queue.Head.Right
	c.Queue.Head.Right = n
	n.Left = c.Queue.Head
	n.Right = tmp
	tmp.Left = n

	c.Queue.Length++
	if c.Queue.Length > SIZE {
		c.Remove(c.Queue.Tail.Left)
	}

}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
func startupServer() {
	r := mux.NewRouter()
	log.Println("The cache is up")
	r.HandleFunc("/", home)
	r.HandleFunc("/employeeDetails/{id}", employeeDetails).Methods("GET")
	r.HandleFunc("/alterEmployee", alterEmployee).Methods("PUT")
	// http.HandleFunc("/getSortedByWBS", getSortedWBS)
	// http.HandleFunc("/getSortedByDatesAndWBS", getSortedByDatesAndWBS)

	log.Fatal(http.ListenAndServe("localhost:8000", r))

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
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		log.Println(err)
		WriteJSONResponse(w, 500, "database connection failed")
		return
	}
	log.Println("Database connected")

	statement := `UPDATE employee SET employeename = $1 WHERE employeeid = $2;`

	_, err = db.Exec(statement, emp.Employeename, emp.Employeeid)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Employee Name is  : ", emp.Employeename, emp.Employeeid)
	node := &Node{employeeid: emp.Employeeid, employeename: emp.Employeename}
	cash.Update(node)
	WriteJSONResponse(w, 200, emp.Employeename)
	return

}

func home(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "The cache is running")
}

var cash = NewCache()

func employeeDetails(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]
	cash.Lock()
	defer cash.Unlock()
	node, b := cash.IfHit(id)
	fmt.Println(node.employeename, b)
	if b == false {
		db, err := sql.Open("postgres", conStr)
		if err != nil {
			log.Println(err)
			WriteJSONResponse(w, 500, "database connection failed")
			return
		}
		log.Println("Database connected")

		statement := `Select employeeName from employee where employeeid = $1`
		name := ""
		err = db.QueryRow(statement, id).Scan(&name)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Employee Name is  : ", name)
		node := &Node{employeeid: id, employeename: name}
		cash.Add(node)
		cash.Hash[id] = node
		WriteJSONResponse(w, 200, name)
		return
	}
	WriteJSONResponse(w, 200, node.employeename)

}

func main() {

	startupServer()
}
