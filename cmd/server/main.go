package main

import (
	"encoding/json"
	"github.com/romang000/java-semantic-web/internal/service"
	"log"
	"net/http"
	
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	
	// POST /analyze
	r.Post("/analyze", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, r.ContentLength)
		_, err := r.Body.Read(data)
		if err != nil && err.Error() != "EOF" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		errs, ast, err := service.CheckJavaSemantic(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		if len(errs) > 0 {
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"errors":  errs,
			})
			return
		}
		
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"ast":     ast,
		})
	})
	
	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", r)
}
