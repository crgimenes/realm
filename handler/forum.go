package handler

import "net/http"

func Forum(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}
