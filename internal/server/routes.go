package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *Server) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", s.CreateUser)
	mux.HandleFunc("GET /users", s.GetAllUsers)
	mux.HandleFunc("GET /users/{id}", s.GetUserById)
	mux.HandleFunc("PUT /users/{id}", s.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", s.DeleteUser)

	return mux
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	user := &User{}
	if err := dec.Decode(&user); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// catch syntax error in JSON and send error message with location
		// of the problem to make it easier for client to fix
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// in some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			http.Error(w, msg, http.StatusBadRequest)

		// catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our User struct
		// send error message with field name and position of error
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// catch the error caused by extra unexpected fields in the request body
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// if body is empty
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// catch the error caused by the request body being too large
		case errors.Is(err, &http.MaxBytesError{}):
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// otherwise default to logging the error and sending a 500 Internal Server Error response
		default:
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// if the body has only one JSON io.EOF error will be returned
	// so if we get anything else, we know that there is additional data in the body
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	stmt := "INSERT INTO users (name, email) VALUES ($1, $2)"
	_, err := s.db.Exec(stmt, user.Name, user.Email)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	stmt := "SELECT id, name, email, created_at, updated_at FROM users"

	rows, err := s.db.Query(stmt)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		user := &User{}
		if err = rows.Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetUserById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	stmt := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"

	row := s.db.QueryRow(stmt, id)

	user := &User{}

	// the number of arguments must be exactly the same as the number of
	// columns returned by your statement
	if err = row.Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			msg := "User not found"
			http.Error(w, msg, http.StatusNotFound)
			return
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	user := &User{}
	if err := dec.Decode(&user); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// catch syntax error in JSON and send error message with location
		// of the problem to make it easier for client to fix
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// in some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			http.Error(w, msg, http.StatusBadRequest)

		// catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our User struct
		// send error message with field name and position of error
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// catch the error caused by extra unexpected fields in the request body
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// if body is empty
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// catch the error caused by the request body being too large
		case errors.Is(err, &http.MaxBytesError{}):
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// otherwise default to logging the error and sending a 500 Internal Server Error response
		default:
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// if the body has only one JSON io.EOF error will be returned
	// so if we get anything else, we know that there is additional data in the body
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	stmt := "UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3"
	_, err = s.db.Exec(stmt, &user.Name, &user.Email, id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	stmt := "DELETE FROM users WHERE id = $1"

	result, err := s.db.Exec(stmt, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		msg := "User not found"
		http.Error(w, msg, http.StatusNotFound)
		return
	}
}
