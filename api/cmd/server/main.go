package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sidryenireddy/prism/api/internal/database"
	"github.com/sidryenireddy/prism/api/internal/engine"
	"github.com/sidryenireddy/prism/api/internal/handlers"
	"github.com/sidryenireddy/prism/api/internal/ontology"
)

func main() {
	migrate := flag.Bool("migrate", false, "run database migrations and exit")
	flag.Parse()

	ctx := context.Background()

	pool, err := database.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if *migrate {
		fmt.Println("Migrations completed successfully")
		return
	}

	ontologyClient := ontology.NewClient()
	eng := engine.New(ontologyClient)

	analysisHandler := handlers.NewAnalysisHandler(pool)
	cardHandler := handlers.NewCardHandler(pool, eng)
	dashboardHandler := handlers.NewDashboardHandler(pool)
	aiHandler := handlers.NewAIHandler(pool)
	mockDataHandler := handlers.NewMockDataHandler()

	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Analyses
		r.Get("/analyses", analysisHandler.List)
		r.Post("/analyses", analysisHandler.Create)
		r.Get("/analyses/{id}", analysisHandler.Get)
		r.Patch("/analyses/{id}", analysisHandler.Update)
		r.Delete("/analyses/{id}", analysisHandler.Delete)

		// Cards
		r.Get("/analyses/{analysisId}/cards", cardHandler.List)
		r.Post("/analyses/{analysisId}/cards", cardHandler.Create)
		r.Patch("/analyses/{analysisId}/cards/{cardId}", cardHandler.Update)
		r.Delete("/analyses/{analysisId}/cards/{cardId}", cardHandler.Delete)

		// Execute
		r.Post("/analyses/{analysisId}/execute", cardHandler.Execute)

		// Dashboards
		r.Get("/dashboards", dashboardHandler.List)
		r.Post("/dashboards", dashboardHandler.Create)
		r.Get("/dashboards/{id}", dashboardHandler.Get)
		r.Patch("/dashboards/{id}", dashboardHandler.Update)
		r.Delete("/dashboards/{id}", dashboardHandler.Delete)

		// AI
		r.Post("/ai/generate", aiHandler.Generate)
		r.Post("/ai/configure/{cardId}", aiHandler.Configure)

		// Mock data
		r.Get("/mock/object-types", mockDataHandler.ObjectTypes)
		r.Get("/mock/objects", mockDataHandler.Objects)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Prism API starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
