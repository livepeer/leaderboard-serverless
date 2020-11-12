package middleware

import "net/http"

func HandlePreflightRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodOptions {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Write(nil)
}
