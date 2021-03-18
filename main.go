package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type userHandlers struct {
	sync.Mutex
	store map[string]User
}

func (u *userHandlers) upsert(w http.ResponseWriter, r *http.Request) {
	u.Lock()
	defer u.Unlock()

	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	params := mux.Vars(r)
	id := params["id"]
	u.store[id] = user

	w.WriteHeader(http.StatusOK)
}

func (u *userHandlers) get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	if v, ok := u.store[id]; ok {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		w.Header().Add("content-type", "application/json")
		w.Write(jsonBytes)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (u *userHandlers) list(w http.ResponseWriter, r *http.Request) {
	u.Lock()
	defer u.Unlock()

	var users []User
	for _, user := range u.store {
		users = append(users, user)
	}

	jsonBytes, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.Write(jsonBytes)
}

func (u *userHandlers) delete(w http.ResponseWriter, r *http.Request) {
	u.Lock()
	defer u.Unlock()

	params := mux.Vars(r)
	id := params["id"]
	delete(u.store, id)
	w.WriteHeader(http.StatusOK)
}

func main() {
	handler := userHandlers{store: map[string]User{
		"1": {
			Name: "Yuri",
			Age:  32,
		},
	}}
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", handler.upsert).Methods(http.MethodPost)
	r.HandleFunc("/users/{id}", handler.get).Methods(http.MethodGet)
	r.HandleFunc("/users", handler.list).Methods(http.MethodGet)
	r.HandleFunc("/users/{id}", handler.delete).Methods(http.MethodDelete)
	log.Fatal(http.ListenAndServe(":8000", r))
}
