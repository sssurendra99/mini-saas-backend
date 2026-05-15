package main

import (
    "log"
    "net/http"
)

func main(){

    

    mux := http.NewServeMux()

    server := &http.Server{
        Addr: ":9000",
        Handler: mux,
    }
    
    log.Println("Server running on http://localhost:9000")

    if err := server.ListenAndServe() ; err != nil {
        log.Fatal(err)
    }
}
