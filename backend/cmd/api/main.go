package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/simplix/api/internal/config"
	"github.com/simplix/api/internal/db"
	"github.com/simplix/api/internal/handlers"
	mw "github.com/simplix/api/internal/middleware"
	"github.com/simplix/api/internal/repository"
	"github.com/simplix/api/internal/service"
)

func main() {
	cfg := config.Load()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var logger zerolog.Logger
	if cfg.Env == "development" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}
	log.Logger = logger

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	pool, dbErr := db.Connect(ctx, cfg.DatabaseURL)
	cancel()
	if dbErr != nil {
		log.Fatal().Err(dbErr).Msg("db connect failed")
	}
	defer pool.Close()
	log.Info().Msg("database connected")

	// SSE broker
	sseBroker := service.NewSSEBroker()

	// Repositories
	userRepo     := repository.NewUserRepo(pool)
	contactRepo  := repository.NewContactRepo(pool)
	convRepo     := repository.NewConversationRepo(pool)
	msgRepo      := repository.NewMessageRepo(pool)
	labelRepo    := repository.NewLabelRepo(pool)
	webhookRepo  := repository.NewWebhookRepo(pool)
	settingsRepo := repository.NewSettingsRepo(pool)
	reportsRepo  := repository.NewReportsRepo(pool)
	noteRepo     := repository.NewNoteRepo(pool)
	attrRepo     := repository.NewCustomAttributeRepo(pool)
	inboxRepo    := repository.NewInboxRepo(pool)
	companyRepo  := repository.NewCompanyRepo(pool)

	// Services
	whatsappSvc := service.NewWhatsAppService()
	quepasaSvc  := service.NewQuePasaService()

	// Handlers
	authH        := handlers.NewAuthHandler(userRepo, cfg.JWTSecret)
	contactH     := handlers.NewContactHandler(contactRepo)
	convH        := handlers.NewConversationHandler(convRepo, msgRepo, inboxRepo, contactRepo, whatsappSvc, quepasaSvc)
	labelH       := handlers.NewLabelHandler(labelRepo)
	reportH      := handlers.NewReportHandler(reportsRepo)
	webhookH     := handlers.NewWebhookHandler(webhookRepo)
	userH        := handlers.NewUserHandler(userRepo)
	settingsH    := handlers.NewSettingsHandler(settingsRepo)
	noteH        := handlers.NewNoteHandler(noteRepo, sseBroker)
	attrH        := handlers.NewCustomAttributeHandler(attrRepo)
	sseH         := handlers.NewSSEHandler(sseBroker)
	inboxH       := handlers.NewInboxHandler(inboxRepo, whatsappSvc, quepasaSvc, cfg.PublicURL)
	companyH     := handlers.NewCompanyHandler(companyRepo)
	waWebhookH   := handlers.NewWhatsAppWebhookHandler(inboxRepo, contactRepo, convRepo, msgRepo, sseBroker, whatsappSvc)
	qpWebhookH   := handlers.NewQuePasaWebhookHandler(inboxRepo, contactRepo, convRepo, msgRepo, sseBroker)

	// Router
	r := chi.NewRouter()
	r.Use(hlog.NewHandler(log.Logger))
	r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("request")
	}))
	r.Use(hlog.RequestIDHandler("req_id", "X-Request-Id"))
	r.Use(hlog.RemoteAddrHandler("ip"))
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.JSON(w, 200, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/sign_in", authH.SignIn)
		r.Post("/auth/sign_up", authH.SignUp)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(cfg.JWTSecret))

			r.Delete("/auth/sign_out", authH.SignOut)
			r.Get("/auth/me", authH.Me)
			r.Patch("/auth/profile", authH.UpdateProfile)

			// Contacts
			r.Get("/contacts", contactH.List)
			r.Post("/contacts", contactH.Create)
			r.Get("/contacts/{id}", contactH.Get)
			r.Patch("/contacts/{id}", contactH.Update)
			r.Delete("/contacts/{id}", contactH.Delete)
			r.Get("/contacts/{id}/notes", noteH.List)
			r.Post("/contacts/{id}/notes", noteH.Create)
			r.Delete("/contacts/{id}/notes/{nid}", noteH.Delete)
			r.Get("/contacts/{id}/labels", contactH.GetLabels)
			r.Post("/contacts/{id}/labels", contactH.SetLabels)
			r.Get("/contacts/{id}/conversations", func(w http.ResponseWriter, rq *http.Request) {
				q := rq.URL.Query()
				q.Set("contact_id", chi.URLParam(rq, "id"))
				rq.URL.RawQuery = q.Encode()
				convH.List(w, rq)
			})

			// Conversations & Messages
			r.Get("/conversations", convH.List)
			r.Post("/conversations", convH.Create)
			r.Get("/conversations/{id}", convH.Get)
			r.Patch("/conversations/{id}", convH.Update)
			r.Get("/conversations/{id}/messages", convH.ListMessages)
			r.Post("/conversations/{id}/messages", convH.CreateMessage)
			r.Delete("/conversations/{id}/messages/{mid}", convH.DeleteMessage)
			r.Post("/conversations/{id}/assignments", convH.Assign)
			r.Get("/conversations/{id}/labels", convH.GetLabels)
			r.Post("/conversations/{id}/labels", convH.SetLabels)

			// Companies
			r.Get("/companies", companyH.List)
			r.Post("/companies", companyH.Create)
			r.Get("/companies/{id}", companyH.Get)
			r.Patch("/companies/{id}", companyH.Update)
			r.Delete("/companies/{id}", companyH.Delete)
			r.Get("/companies/{id}/contacts", companyH.ListContacts)

			// Labels
			r.Get("/labels", labelH.List)
			r.Post("/labels", labelH.Create)
			r.Patch("/labels/{id}", labelH.Update)
			r.Delete("/labels/{id}", labelH.Delete)

			// Reports
			r.Get("/reports/overview", reportH.Overview)
			r.Get("/reports/contacts", reportH.Contacts)
			r.Get("/reports/conversations", reportH.Conversations)
			r.Get("/reports/agents", reportH.Agents)
			r.Get("/reports/timeseries", reportH.TimeSeries)

			// Webhooks
			r.Get("/webhooks", webhookH.List)
			r.Post("/webhooks", webhookH.Create)
			r.Patch("/webhooks/{id}", webhookH.Update)
			r.Delete("/webhooks/{id}", webhookH.Delete)

			// Users (admin only)
			r.Group(func(r chi.Router) {
				r.Use(mw.RequireAdmin)
				r.Get("/users", userH.List)
				r.Post("/users", userH.Create)
				r.Patch("/users/{id}", userH.Update)
			})

			// Settings
			r.Get("/settings", settingsH.Get)
			r.Patch("/settings", settingsH.Update)

			// Custom Attributes
			r.Get("/custom-attributes", attrH.List)
			r.Post("/custom-attributes", attrH.Create)
			r.Patch("/custom-attributes/{id}", attrH.Update)
			r.Delete("/custom-attributes/{id}", attrH.Delete)

			// Inboxes
			r.Get("/inboxes", inboxH.List)
			r.Post("/inboxes", inboxH.Create)
			r.Get("/inboxes/{id}", inboxH.Get)
			r.Patch("/inboxes/{id}", inboxH.Update)
			r.Delete("/inboxes/{id}", inboxH.Delete)
			r.Get("/inboxes/{id}/templates", inboxH.ListTemplates)
			r.Post("/inboxes/{id}/templates/sync", inboxH.SyncTemplates)

			// SSE stream
			r.Get("/events", sseH.Stream)
		})
	})

	// WhatsApp webhook — public (no auth)
	r.Get("/webhook/whatsapp/{inbox_id}", waWebhookH.Verify)
	r.Post("/webhook/whatsapp/{inbox_id}", waWebhookH.Process)

	// QuePasa webhook — public (no auth)
	r.Post("/webhook/quepasa/{inbox_id}", qpWebhookH.Process)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info().Str("addr", addr).Str("env", cfg.Env).Msg("server listening")
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}
