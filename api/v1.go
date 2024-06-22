package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yiffyi/waterplz/upstream"
)

type v1DoParams struct {
	Username string
	Password string
	SN       string
	Language string
}

func CreateV1Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/do",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Sorry, only POST is allowed", http.StatusMethodNotAllowed)
				return
			}

			p := &v1DoParams{}
			err := json.NewDecoder(r.Body).Decode(p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if strings.ToLower(p.Language) != "english" {
				http.Error(w, "Sorry, only English is supported", http.StatusBadRequest)
				return
			}

			t1 := time.Now()
			s, err := upstream.CreateSession(p.Username, p.Password)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			t2 := time.Now()
			if s.CreateOrder(p.SN) == nil {
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "Hot water is coming!\n")
				io.WriteString(w,
					fmt.Sprintf("Login spent %.2f seconds\nCreate order spent %.2f seconds\n", t2.Sub(t1).Seconds(), time.Since(t2).Seconds()))
			} else {
				// s.CloseOrder("C47F0ED55446")
				t3 := time.Now()
				s.CloseOrder(p.SN)
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "Valve closed.\n")
				io.WriteString(w,
					fmt.Sprintf("Login spent %.2f seconds\nCreate order spent %.2f seconds\nClose order spent %.2f seconds", t2.Sub(t1).Seconds(), t3.Sub(t2).Seconds(), time.Since(t3).Seconds()))
			}
		})
	return mux
}
