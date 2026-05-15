package response

import (
    "net/http"
    "encoding/json"
)

func writeJSON(w http.ResponseWriter, status int, data any){
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    json.NewEncoder(w).Encode(data)
}
