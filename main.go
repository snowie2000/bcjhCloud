// bcjhCloud project main.go
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("bcjh 1.0 started")
	mux := http.NewServeMux()
	st := &Store{}
	st.Init("store.db")
	mux.HandleFunc("/put", st.Put)
	mux.HandleFunc("/get", st.Get)

	addr := flag.String("addr", "0.0.0.0:80", "Bind address")
	flag.Parse()
	http.ListenAndServe(*addr, mux)
}

type Store struct {
	db *sql.DB
}

func (s *Store) Init(filename string) {
	var err error
	s.db, err = sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalln(err)
	}
}

type GenericReq struct {
	User  string
	Key   string
	Value string
}

func (s *Store) Put(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	de := json.NewDecoder(r.Body)
	var req GenericReq
	if err := de.Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Key == "" || req.User == "" {
		http.Error(w, "Empty key or user", http.StatusBadRequest)
		return
	}
	var id int32 = -1
	row := s.db.QueryRow("select id from [store] where name=? and key=?", req.User, req.Key)
	if row.Scan(&id) != nil || id < 0 {
		s.db.Exec("insert into [store] (name, key, value) values(?,?,?)", req.User, req.Key, req.Value)
	} else {
		s.db.Exec("update [store] set value=? where name=? and key=?", req.Value, req.User, req.Key)
	}
	w.WriteHeader(200)
}

func (s *Store) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	de := json.NewDecoder(r.Body)
	var req GenericReq
	if err := de.Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Key == "" || req.User == "" {
		http.Error(w, "Empty key or user", http.StatusBadRequest)
		return
	}
	var (
		id    int32 = -1
		value string
	)
	row := s.db.QueryRow("select id, value from [store] where name=? and key=?", req.User, req.Key)
	if row.Scan(&id, &value) == nil && id >= 0 {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(value))
	}
	w.WriteHeader(200)
}
