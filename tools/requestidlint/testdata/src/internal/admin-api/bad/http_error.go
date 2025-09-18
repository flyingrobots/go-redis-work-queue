package bad

import "net/http"

func HandlerHTTPError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "oops", http.StatusInternalServerError) // want "use writeError helper to ensure X-Request-ID header is set instead of http.Error"
}
