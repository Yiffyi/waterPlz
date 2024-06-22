package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/yiffyi/waterplz/upstream"
)

type doParams struct {
	Username string
	Password string
	SN       string
}

func doHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "I can't do that.", http.StatusMethodNotAllowed)
		return
	}

	p := &doParams{}
	err := json.NewDecoder(r.Body).Decode(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s, err := upstream.CreateSession(p.Username, p.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if s.CreateOrder(p.SN) == nil {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "热水开启咯！")
		return
	} else {
		// s.CloseOrder("C47F0ED55446")
		s.CloseOrder(p.SN)
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "热水关闭咯！")
		return
	}
}

func main() {
	http.HandleFunc("/do", doHandler)

	http.ListenAndServe(":8080", nil)
}
