package bad

import "net/http"

func HandlerWriteHeader(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot) // want "use writeError helper to ensure X-Request-ID header is set instead of calling WriteHeader directly"
}
