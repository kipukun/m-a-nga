package main

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/kipukun/m-a-nga/db"
	"golang.org/x/crypto/bcrypt"
)

type state struct {
	tmpl *template.Template
	srv  *http.Server
	qs   *db.Queries
}

func (s *state) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := s.tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}
}

func (s *state) handleLogin(w http.ResponseWriter, r *http.Request) {
	err := s.tmpl.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}
}

func (s *state) handleAuthLogin(w http.ResponseWriter, r *http.Request) {

}

func (s *state) handleAuthSignup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	user := r.Form.Get("user")
	pass := r.Form.Get("pass")

	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	p := db.CreateUserParams{
		User: user,
		Pass: string(hashed),
	}
	err = s.qs.CreateUser(r.Context(), p)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
}

//go:embed static/*
var content embed.FS

func main() {
	s := new(state)
	t, err := template.ParseFS(content, "static/*.html")
	if err != nil {
		log.Fatalln(err)
	}
	s.tmpl = t

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conn, err := sqlx.ConnectContext(ctx, "postgres", "user=foo dbname=bar sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	s.qs = db.New(conn)

	r := mux.NewRouter()
	r.HandleFunc("/", s.handleIndex).Methods("GET")
	r.HandleFunc("/login", s.handleLogin).Methods("GET")
	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", s.handleAuthLogin).Methods("POST")
	auth.HandleFunc("/signup", s.handleAuthSignup).Methods("POST")

	s.srv = &http.Server{
		Addr:         "127.0.0.1:5000",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		Handler:      r,
	}

	go func() {
		<-ctx.Done()
		s.srv.Close()
	}()
	s.srv.ListenAndServe()
}
