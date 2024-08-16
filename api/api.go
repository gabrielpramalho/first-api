package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
)

type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Biography string `json:"biography"`
}

type Application struct {
	Data map[uuid.UUID]User
}

func NewHandler(db Application) http.Handler {

	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Post("/api/users", handleInsert(db))
	r.Delete("/api/user/{id}", handleDelete(db))
	r.Put("/api/user/{id}", handleUpdate(db))
	r.Get("/api/users", handleFindAll(db))
	r.Get("/api/user/{id}", handleFindById(db))

	return r
}

func handleInsert(db Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body User
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			sendJSON(w, Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
			return
		}

		newID := uuid.New()

		db.Data[newID] = body

		sendJSON(w, Response{Data: newID}, http.StatusCreated)
	}
}
func handleDelete(db Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		i := chi.URLParam(r, "id")

		id, err := uuid.Parse(i)

		if err != nil {
			sendJSON(w, Response{Error: "invalid Param"}, http.StatusUnprocessableEntity)
			return
		}

		_, ok := db.Data[id]

		if !ok {
			sendJSON(w, Response{Data: "user not found"}, http.StatusNotFound)
			return
		}

		delete(db.Data, id)
		sendJSON(w, Response{}, http.StatusNoContent)
	}
}
func handleUpdate(db Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body User
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			sendJSON(w, Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
			return
		}

		i := chi.URLParam(r, "id")

		id, err := uuid.Parse(i)

		if err != nil {
			sendJSON(w, Response{Error: "invalid Param"}, http.StatusUnprocessableEntity)
			return
		}

		_, ok := db.Data[id]

		if !ok {
			sendJSON(w, Response{Data: "user not found"}, http.StatusNotFound)
			return
		}

		db.Data[id] = User{
			FirstName: body.FirstName,
			LastName: body.LastName,
			Biography: body.Biography,
		}

		sendJSON(w, Response{}, http.StatusNoContent)
	}
}

type Items struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Biography string `json:"biography"`
}

func handleFindAll(db Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var allData []Items
		for id, user := range db.Data {
			allData = append(allData, Items{
				ID:        id.String(),
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Biography: user.Biography,
			})
		}

		sendJSON(w, Response{Data: allData}, http.StatusCreated)
	}
}

func handleFindById(db Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		i := chi.URLParam(r, "id")

		id, err := uuid.Parse(i)

		if err != nil {
			sendJSON(w, Response{Error: "invalid Param"}, http.StatusUnprocessableEntity)
			return
		}

		user, ok := db.Data[id]

		if !ok {
			sendJSON(w, Response{Data: "user not found"}, http.StatusNotFound)
			return
		}

		sendJSON(w, Response{Data: user}, http.StatusOK)
	}
}

func sendJSON(w http.ResponseWriter, resp Response, status int) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal json data", "error", err)
		sendJSON(
			w,
			Response{Error: "something went wrong"},
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(status)

	if _, err := w.Write(data); err != nil {
		slog.Error("failed to write response to client", "error", err)
		return
	}
}

type Response struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}
