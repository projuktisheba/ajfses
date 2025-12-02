package middlewares

import (
	"log"
	"net/http"
	"os"
)

// ========================= LOGGER ==============================
func Logger(next http.Handler) http.Handler {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		infoLog.Println("Received request:", r.Method, r.URL.Path, "from", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
