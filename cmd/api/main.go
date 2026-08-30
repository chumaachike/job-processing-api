package main

import (
	"log"
	"net/http"
	"encoding/json"
)

func main(){
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("content-type", "application/json")

		json.NewEncoder(w).Encode(map[string]string{
			"status" : "ok",
		})
	})


	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("API is listening on :8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}