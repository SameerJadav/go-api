package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/SameerJadav/go-api/internal/logger"
	"github.com/SameerJadav/go-api/internal/middleware"
	"github.com/justinas/alice"
)

type user struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", s.CreateUser)
	mux.HandleFunc("GET /users", s.GetAllUsers)
	mux.HandleFunc("GET /users/{id}", s.GetUserById)
	mux.HandleFunc("PUT /users/{id}", s.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", s.DeleteUser)

	return alice.New(middleware.RecoverPanic, middleware.LogRequest, middleware.SecureHeaders).Then(mux)
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !validateContentType(w, r) {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	user := &user{}
	if err := dec.Decode(&user); err != nil {
		handleJSONDecodeErrors(w, err)
		return
	}

	if !validateSignleJSONObject(w, r) {
		return
	}

	stmt := "INSERT INTO users (name, email) VALUES ($1, $2)"

	_, err := s.db.Exec(stmt, user.Name, user.Email)
	if err != nil {
		logger.Error.Fatalln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	stmt := "SELECT id, name, email, created_at, updated_at FROM users"

	rows, err := s.db.Query(stmt)
	if err != nil {
		logger.Error.Fatalln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []*user{}

	for rows.Next() {
		user := &user{}
		if err = rows.Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			logger.Error.Fatalln(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(users); err != nil {
		logger.Error.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetUserById(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(w, r)
	if err != nil {
		return
	}

	stmt := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"

	row := s.db.QueryRow(stmt, id)

	user := &user{}

	// the number of arguments must be exactly the same as the number of
	// columns returned by your statement
	if err = row.Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			msg := "User not found"
			http.Error(w, msg, http.StatusNotFound)
			return
		} else {
			logger.Error.Fatalln(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(user); err != nil {
		logger.Error.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if !validateContentType(w, r) {
		return
	}

	id, err := parseIDFromPath(w, r)
	if err != nil {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	user := &user{}
	if err := dec.Decode(&user); err != nil {
		handleJSONDecodeErrors(w, err)
		return
	}

	if !validateSignleJSONObject(w, r) {
		return
	}

	stmt := "UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3"

	_, err = s.db.Exec(stmt, &user.Name, &user.Email, id)
	if err != nil {
		logger.Error.Fatalln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(w, r)
	if err != nil {
		return
	}

	stmt := "DELETE FROM users WHERE id = $1"

	result, err := s.db.Exec(stmt, id)
	if err != nil {
		logger.Error.Fatalln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error.Fatalln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		msg := "User not found"
		http.Error(w, msg, http.StatusNotFound)
		return
	}
}
