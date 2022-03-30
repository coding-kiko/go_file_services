package server

import (
	// std lib
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	// Internal
	"github.com/coding-kiko/file_storage_testing/pkg/auth"
	"github.com/coding-kiko/file_storage_testing/pkg/service"
)

func GetFileHandler(sv service.ImageService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// check if filename is not passed correctly through query params
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": 400, "message": "Bad Request"}`))
			return
		}
		// check if another method other than GET has reached the endpoint
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"status": 405, "message": "Method not allowed"}`))
			return
		}
		err := sv.GetFile(w, r, filename)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf(`{"status": 404, "message": "%s"}`, err.Error())))
			return
		}
		w.Write([]byte(`{"status": 200, "message": "file retrieved successfully"}`))
		w.Write([]byte("\n"))
	})
}

func CreateFileHandler(sv service.ImageService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// check if another method other than POST has reached the endpoint
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"status": 405, "message": "Method not allowed"}`))
			return
		}
		err := sv.CreateFile(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"status": 400, "message": "%s"}`, err.Error())))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": 200, "message": "file stored successfully"}`))
		w.Write([]byte("\n"))
	})
}

func AuthenticateHandler(rd auth.RedisRespository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var creds auth.Credentials
		w.Header().Set("Content-Type", "application/json")
		// check if another method other than POST has reached the endpoint
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"status": 405, "message": "Method not allowed"}`))
			return
		}
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": 400, "message": "Bad request"}`))
			return
		}
		token, err := rd.Authenticate(creds)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(fmt.Sprintf(`{"status": 422, "message": "%s"}`, err.Error())))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"token": "%s"}`, token)))
		w.Write([]byte("\n"))
	})
}

func RegisterHandler(rd auth.RedisRespository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var creds auth.Credentials
		w.Header().Set("Content-Type", "application/json")
		// check if another method other than POST has reached the endpoint
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"status": 405, "message": "Method not allowed"}`))
			return
		}
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": 400, "message": "Bad request"}`))
			return
		}
		err = rd.Register(creds)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"status": 422, "message": "invalid credentials"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": 200, "message": "user registered"}`))
		w.Write([]byte("\n"))
	})
}

func JwtMiddleware(next http.Handler, rd auth.RedisRespository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Authorization"] == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": 400, "message": "Missing header: 'Authorization'"}`))
			w.Write([]byte("\n"))
			return
		}
		authorization := strings.Split(r.Header["Authorization"][0], " ")
		if authorization[0] != "Bearer" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": 400, "message": "Malformed header: wrong authentication scheme"}`))
			w.Write([]byte("\n"))
			return
		}
		err := rd.ValidateJwt(authorization[1])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf(`{"status": 401, "message": "%s"}`, err.Error())))
			w.Write([]byte("\n"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
