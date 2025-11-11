package main

import (
	"encoding/json"
	"github.com/romang000/java-semantic-web/internal/service"
	"io"
	"log"
	"net/http"
	
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	
	r.Post("/analyze", func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		
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
