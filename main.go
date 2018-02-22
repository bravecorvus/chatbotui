package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

var (
	redisAddress = flag.String("redis-address", ":6379", "Address to the Redis server")
)

func Pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir + "/"
}

type QueryAnswer struct {
	Query  string `json:"query"`
	Answer string `json:"answer"`
}

type QueryAnswerList struct {
	QueryAnswers []QueryAnswer `json:"list"`
}

//

func getQueries(c redis.Conn) *QueryAnswerList {
	list := &QueryAnswerList{}
	ret, _ := redis.Strings(c.Do("HKEYS", "responses"))
	for _, elem := range ret {
		fmt.Println(elem)
		answer, _ := redis.String(c.Do("HGET", "responses", elem))
		list.QueryAnswers = append(list.QueryAnswers, QueryAnswer{Query: elem, Answer: answer})
	}
	return list

}

func addNew(c redis.Conn, key, value string) {
	_, _ = redis.String(c.Do("HSET", "responses", key, value))
}

func remove(c redis.Conn, key string) {
	_, _ = redis.String(c.Do("HDEL", "responses", key))
}

func main() {

	redisconn, err := redis.Dial("tcp", ":6379")
	defer redisconn.Close()

	ret, _ := redisconn.Do("SELECT", "1")
	fmt.Printf("%s\n", ret)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, Pwd()+"public/index.html")
	}).Methods("GET")

	r.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(getQueries(redisconn))
	}).Methods("GET")

	// 3 slugs guaranteed to be add new element structure
	r.HandleFunc("/{slug1}/{slug2}/{slug3}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := slugToString(vars["slug2"])
		value := slugToString(vars["slug3"])
		addNew(redisconn, key, value)
	}).Methods("POST")

	// 2 slugs guaranteed to be delete element structure
	r.HandleFunc("/{slug1}/{slug2}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := slugToString(vars["slug2"])
		remove(redisconn, key)
	}).Methods("POST")

	err = http.ListenAndServe(":1005", r)
	// err = http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}

func slugToString(arg string) string {
	semifixed := strings.Replace(arg, "_", " ", -1)
	return strings.Replace(semifixed, "QUESTIONMARK", "?", -1)
}
