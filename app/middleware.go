package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/user"
	"github.com/andrewbenington/queue-share-api/util"
)

type middleware func(next http.Handler) http.Handler

var allMiddleware []middleware = []middleware{
	authMW,
	contentMW,
	timeoutMW,
	logMW,
	corsMW,
}

func withMiddleware(handler http.Handler) http.Handler {
	for _, mw := range allMiddleware {
		handler = mw(handler)
	}

	return handler
}

func authMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) < 2 {
			next.ServeHTTP(w, r)
			return
		}
		reqToken = splitToken[1]

		id, err := user.GetTokenID(reqToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(requests.ErrorResponse{Error: err.Error()})
			return
		}

		friendID := r.URL.Query().Get("friend_id")
		if r.Method != "GET" || !(strings.HasPrefix(r.RequestURI, "/user") || strings.HasPrefix(r.RequestURI, "/admin")) {
			util.EnqueueRequestLog(util.LogEntry{
				Timestamp: time.Now(),
				UserID:    id,
				Endpoint:  r.RequestURI,
				FriendID:  friendID,
			})
		}
		ctx := context.WithValue(r.Context(), auth.UserContextKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func contentMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		}
	})
}

func corsMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, *")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			log.Printf("%s - %s (%s)", r.Method, r.URL.Path, r.RemoteAddr)
		}

		next.ServeHTTP(w, r)
	})
}

func timeoutMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*300)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
