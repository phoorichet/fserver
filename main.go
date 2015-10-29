package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

var (
	keyCountPckRecv = "countPckRecv"
	keyCountPckSent = "countPckSent"
	keyCountPckErr  = "countPckErr"

	userPrefix = "/users/"
)

type User struct {
	Dn  string
	Age int
}

type Users struct {
	users map[string]*User
}

func NewUsers(from, to int) *Users {
	users := make(map[string]*User)
	var dn string
	var age int
	for i := from; i < to; i++ {
		dn = fmt.Sprintf("uid=%015d", i)
		age = rand.Int()
		users[dn] = &User{Dn: dn, Age: age % 70}
	}
	return &Users{users: users}
}

func (u *Users) Get(dn string) *User {
	return u.users[dn]
}

type Response struct {
	ReturnCode int
	Data       *User
}

func main() {
	countPckRecv := expvar.NewInt(keyCountPckRecv)
	countPckRecv.Set(0)

	countPckSent := expvar.NewInt(keyCountPckSent)
	countPckSent.Set(0)

	countPckErr := expvar.NewInt(keyCountPckErr)
	countPckErr.Set(0)

	UserMap := NewUsers(0, 100)

	http.HandleFunc(userPrefix, func(w http.ResponseWriter, r *http.Request) {
		countPckRecv.Add(1)

		uid := strings.TrimPrefix(r.URL.Path, userPrefix)

		w.Header().Set("Content-Type", "application/json")

		u := UserMap.Get(uid)
		ret := 0
		switch r.Method {
		case "GET":
			if u == nil {
				ret = 2
			}
			res := Response{ReturnCode: ret, Data: u}

			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
				return
			}
			// 		case "PUT":
			// 			if u == nil {
			// 				http.Error(w, fmt.Sprintf("%s not found", uid), http.StatusInternalServerError)
			// 				return
			// 			}

		case "POST":
		case "DELETE":
		default:
			http.Error(w, "Method not allow", http.StatusMethodNotAllowed)
			return
		}

		countPckSent.Add(1)
		return
	})

	log.Printf("Running server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
