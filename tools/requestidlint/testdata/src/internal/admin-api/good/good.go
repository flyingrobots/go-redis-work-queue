package good

import "net/http"

func writeError(w http.ResponseWriter, status int, code, message string) {}

func HandlerOK(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusBadRequest, "BAD", "bad")
}
