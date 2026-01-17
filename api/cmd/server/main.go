package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/forgo/saga/api/internal/config"
	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/handler"
	"github.com/forgo/saga/api/internal/jobs"
	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/service"
	"github.com/forgo/saga/api/pkg/jwt"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize database connection
	db := database.NewSurrealDB(database.Config{
		Host:      cfg.Database.Host,
		Port:      cfg.Database.Port,
		User:      cfg.Database.User,
		Password:  cfg.Database.Password,
		Namespace: cfg.Database.Namespace,
		Database:  cfg.Database.Database,
	})

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("connected to database",
		slog.String("host", cfg.Database.Host),
		slog.String("database", cfg.Database.Database),
	)

	// Initialize JWT service
	jwtService, err := jwt.NewService(jwt.Config{
		PrivateKeyPath: cfg.JWT.PrivateKeyPath,
		PublicKeyPath:  cfg.JWT.PublicKeyPath,
		Issuer:         cfg.JWT.Issuer,
		ExpirationMins: cfg.JWT.ExpirationMins,
	})
	if err != nil {
		slog.Error("failed to initialize JWT service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	identityRepo := repository.NewIdentityRepository(db)
	passkeyRepo := repository.NewPasskeyRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	// TODO: Implement Guild repository (renamed from Circle)
	// circleRepo := repository.NewCircleRepository(db)
	// memberRepo := repository.NewMemberRepository(db) // TODO: Wire up when guild service is implemented
	personRepo := repository.NewPersonRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	timerRepo := repository.NewTimerRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	interestRepo := repository.NewInterestRepository(db)
	questionnaireRepo := repository.NewQuestionnaireRepository(db)
	availabilityRepo := repository.NewAvailabilityRepository(db)
	resonanceRepo := repository.NewResonanceRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	eventRepo := repository.NewEventRepository(db)
	eventRoleRepo := repository.NewEventRoleRepository(db)
	trustRepo := repository.NewTrustRepository(db)
	trustRatingRepo := repository.NewTrustRatingRepository(db)
	roleCatalogRepo := repository.NewRoleCatalogRepository(db)
	rideshareRoleRepo := repository.NewRideshareRoleRepository(db)
	voteRepo := repository.NewVoteRepository(db)
	adventureRepo := repository.NewAdventureRepository(db)
	adventureAdmissionRepo := repository.NewAdventureAdmissionRepository(db)
	// TODO: Implement Rideshare repository (renamed from Commute)
	// commuteRepo := repository.NewCommuteRepository(db)
	poolRepo := repository.NewPoolRepository(db)
	// TODO: Implement Adventure repository (renamed from Trip)
	// tripRepo := repository.NewTripRepository(db)
	moderationRepo := repository.NewModerationRepository(db)

	// Initialize services
	tokenService := service.NewTokenService(service.TokenServiceConfig{
		JWTService: jwtService,
		TokenRepo:  tokenRepo,
	})

	authService := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:     userRepo,
		IdentityRepo: identityRepo,
		PasskeyRepo:  passkeyRepo,
		TokenService: tokenService,
	})

	oauthService := service.NewOAuthService(service.OAuthServiceConfig{
		Config: service.OAuthConfig{
			Google: service.GoogleOAuthConfig{
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				RedirectURI:  os.Getenv("GOOGLE_REDIRECT_URI"),
			},
			Apple: service.AppleOAuthConfig{
				ClientID:    os.Getenv("APPLE_CLIENT_ID"),
				TeamID:      os.Getenv("APPLE_TEAM_ID"),
				KeyID:       os.Getenv("APPLE_KEY_ID"),
				PrivateKey:  os.Getenv("APPLE_PRIVATE_KEY"),
				RedirectURI: os.Getenv("APPLE_REDIRECT_URI"),
			},
		},
		AuthService:  authService,
		IdentityRepo: identityRepo,
		UserRepo:     userRepo,
		TokenService: tokenService,
	})

	passkeyService := service.NewPasskeyService(service.PasskeyServiceConfig{
		Config: service.PasskeyConfig{
			RPID:            os.Getenv("PASSKEY_RP_ID"),
			RPName:          "Saga",
			RPOrigins:       cfg.Server.AllowedOrigins,
			Timeout:         60 * time.Second,
			RequireUV:       false,
			AttestationType: "none",
		},
		PasskeyRepo:  passkeyRepo,
		UserRepo:     userRepo,
		TokenService: tokenService,
	})

	// TODO: Implement Guild service (renamed from Circle)
	// circleService := service.NewCircleService(service.CircleServiceConfig{
	// 	CircleRepo:   circleRepo,
	// 	MemberRepo:   memberRepo,
	// 	PersonRepo:   personRepo,
	// 	ActivityRepo: activityRepo,
	// 	TimerRepo:    timerRepo,
	// })
	_ = personRepo
	_ = activityRepo
	_ = timerRepo

	profileService := service.NewProfileService(service.ProfileServiceConfig{
		ProfileRepo: profileRepo,
		UserRepo:    userRepo,
	})

	interestService := service.NewInterestService(service.InterestServiceConfig{
		InterestRepo: interestRepo,
	})

	questionnaireService := service.NewQuestionnaireService(service.QuestionnaireServiceConfig{
		Repo: questionnaireRepo,
	})

	compatibilityService := service.NewCompatibilityService(service.CompatibilityServiceConfig{
		QuestionnaireRepo: questionnaireRepo,
	})

	availabilityService := service.NewAvailabilityService(service.AvailabilityServiceConfig{
		Repo: availabilityRepo,
	})

	resonanceService := service.NewResonanceService(service.ResonanceServiceConfig{
		Repo: resonanceRepo,
	})

	reviewService := service.NewReviewService(service.ReviewServiceConfig{
		Repo: reviewRepo,
	})

	trustService := service.NewTrustService(trustRepo)

	trustRatingService := service.NewTrustRatingService(service.TrustRatingServiceConfig{
		Repo: trustRatingRepo,
	})

	roleCatalogService := service.NewRoleCatalogService(service.RoleCatalogServiceConfig{
		CatalogRepo:   roleCatalogRepo,
		RideshareRepo: rideshareRoleRepo,
		GuildRepo:     nil, // TODO: Add when guild repository is implemented
	})

	voteService := service.NewVoteService(service.VoteServiceConfig{
		VoteRepo:  voteRepo,
		GuildRepo: nil, // TODO: Add when guild repository is implemented
	})

	adventureService := service.NewAdventureService(service.AdventureServiceConfig{
		AdventureRepo: adventureRepo,
		AdmissionRepo: adventureAdmissionRepo,
		GuildRepo:     nil, // TODO: Add when guild repository is implemented
	})

	eventRoleService := service.NewEventRoleService(eventRoleRepo, interestService)

	// TODO: Fix EventRepository interface - missing GetByCircle method
	// eventService := service.NewEventService(eventRepo, compatibilityService, questionnaireService, eventRoleService)
	_ = eventRepo
	_ = compatibilityService
	_ = eventRoleService

	// TODO: Implement Rideshare service (renamed from Commute)
	// commuteService := service.NewCommuteService(commuteRepo, trustService)

	// TODO: Fix PoolService - references CircleRepo which no longer exists
	// poolService := service.NewPoolService(service.PoolServiceConfig{
	// 	PoolRepo:      poolRepo,
	// 	CircleRepo:    circleRepo,
	// 	MemberRepo:    memberRepo,
	// 	Compatibility: compatibilityService,
	// })
	_ = poolRepo

	discoveryService := service.NewDiscoveryService(service.DiscoveryServiceConfig{
		AvailabilityRepo:  availabilityRepo,
		CompatibilityRepo: questionnaireRepo,
		InterestRepo:      interestRepo,
		ProfileRepo:       profileRepo,
	})

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
		Rate:   100, // 100 requests per minute
		Window: time.Minute,
		Burst:  20, // Allow bursts up to 20
	})
	defer rateLimiter.Stop()

	// Initialize idempotency store
	idempotencyStore := middleware.NewIdempotencyStore(middleware.IdempotencyConfig{
		TTL:     24 * time.Hour,
		Cleanup: time.Hour,
	})
	defer idempotencyStore.Stop()

	// Initialize event hub for real-time updates
	eventHub := service.NewEventHub()
	defer eventHub.Close()

	// TODO: Implement Adventure service (renamed from Trip)
	// tripService := service.NewTripService(service.TripServiceConfig{
	// 	Repo:     tripRepo,
	// 	EventHub: eventHub,
	// })

	// Initialize moderation service
	moderationService := service.NewModerationService(moderationRepo, eventHub)

	// Initialize threshold monitor for timer alerts
	thresholdMonitor := service.NewThresholdMonitor(service.ThresholdMonitorConfig{
		Checker:  timerRepo,
		EventHub: eventHub,
		Interval: 30 * time.Second,
		Cooldown: 5 * time.Minute,
	})
	thresholdMonitor.Start()
	defer thresholdMonitor.Stop()

	// TODO: Re-enable pool matcher when pool service is fixed
	// poolMatcher := jobs.NewPoolMatcher(poolService, 1*time.Hour)
	// poolMatcher.Start()
	// defer poolMatcher.Stop()

	// Initialize nudge service and processor
	nudgeService := service.NewNudgeService(service.NudgeServiceConfig{
		AvailabilityRepo: availabilityRepo,
		PoolRepo:         poolRepo,
		EventHub:         eventHub,
	})
	nudgeProcessor := jobs.NewNudgeProcessor(nudgeService, 15*time.Minute)
	nudgeProcessor.Start()
	defer nudgeProcessor.Stop()

	// Initialize Nexus monthly job (calculates on 1st of each month)
	nexusMonthlyJob := jobs.NewNexusMonthlyJob(resonanceService, resonanceService)
	nexusMonthlyJob.Start()
	defer nexusMonthlyJob.Stop()

	// Initialize Vote status processor (checks every minute)
	voteStatusProcessor := jobs.NewVoteStatusProcessor(voteService, 1*time.Minute)
	voteStatusProcessor.Start()
	defer voteStatusProcessor.Stop()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	oauthHandler := handler.NewOAuthHandler(oauthService)
	passkeyHandler := handler.NewPasskeyHandler(passkeyService)
	// TODO: Implement Guild handler (renamed from Circle)
	// circleHandler := handler.NewCircleHandler(circleService)
	// personHandler := handler.NewPersonHandler(circleService, eventHub)
	// activityHandler := handler.NewActivityHandler(circleService, eventHub)
	// timerHandler := handler.NewTimerHandler(circleService, eventHub)
	eventsHandler := handler.NewEventsHandler(eventHub)
	profileHandler := handler.NewProfileHandler(profileService)
	interestHandler := handler.NewInterestHandler(interestService)
	// TODO: Fix questionnaire handler - needs compatibilityService
	// questionnaireHandler := handler.NewQuestionnaireHandler(questionnaireService, compatibilityService)
	_ = questionnaireService
	availabilityHandler := handler.NewAvailabilityHandler(availabilityService, profileService)
	resonanceHandler := handler.NewResonanceHandler(resonanceService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	// TODO: Fix event handler - needs eventService
	// eventHandler := handler.NewEventHandler(eventService)
	// eventRoleHandler := handler.NewEventRoleHandler(eventRoleService)
	trustHandler := handler.NewTrustHandler(trustService)
	trustRatingHandler := handler.NewTrustRatingHandler(trustRatingService)
	roleCatalogHandler := handler.NewRoleCatalogHandler(roleCatalogService)
	voteHandler := handler.NewVoteHandler(voteService)
	adventureHandler := handler.NewAdventureHandler(adventureService)
	// TODO: Implement Rideshare handler (renamed from Commute)
	// commuteHandler := handler.NewCommuteHandler(commuteService)
	// TODO: Fix pool handler - needs poolService and circleService
	// poolHandler := handler.NewPoolHandler(poolService, circleService)
	discoveryHandler := handler.NewDiscoveryHandler(discoveryService)
	// TODO: Implement Adventure handler (renamed from Trip)
	// tripHandler := handler.NewTripHandler(tripService, circleService)
	moderationHandler := handler.NewModerationHandler(moderationService, userRepo)

	// Create router and register routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", handler.Health)

	// Auth endpoints (public)
	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/refresh", authHandler.Refresh)

	// OAuth endpoints (public)
	mux.HandleFunc("POST /v1/auth/oauth/google", oauthHandler.Google)
	mux.HandleFunc("POST /v1/auth/oauth/apple", oauthHandler.Apple)

	// Passkey login endpoints (public)
	mux.HandleFunc("POST /v1/auth/passkey/login/start", passkeyHandler.LoginStart)
	mux.HandleFunc("POST /v1/auth/passkey/login/finish", passkeyHandler.LoginFinish)

	// Auth endpoints (protected)
	authMiddleware := middleware.Auth(tokenService)
	mux.Handle("POST /v1/auth/logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /v1/auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))

	// Passkey registration endpoints (protected - user must be logged in)
	mux.Handle("POST /v1/auth/passkey/register/start", authMiddleware(http.HandlerFunc(passkeyHandler.RegisterStart)))
	mux.Handle("POST /v1/auth/passkey/register/finish", authMiddleware(http.HandlerFunc(passkeyHandler.RegisterFinish)))
	mux.Handle("DELETE /v1/auth/passkey/", authMiddleware(http.HandlerFunc(passkeyHandler.Delete)))

	// TODO: Guild access middleware (checks membership) - needs Guild service
	// circleAccess := middleware.GuildAccess(circleService)
	//
	// // Wrap handler with auth + guild access + path params extraction
	// withGuild := func(h http.HandlerFunc) http.Handler {
	// 	return authMiddleware(
	// 		circleAccess(
	// 			middleware.ExtractPathParams(
	// 				http.HandlerFunc(h),
	// 			),
	// 		),
	// 	)
	// }

	// TODO: Guilds endpoints (need Guild handler)
	// mux.Handle("GET /v1/guilds", authMiddleware(http.HandlerFunc(circleHandler.List)))
	// mux.Handle("POST /v1/guilds", authMiddleware(http.HandlerFunc(circleHandler.Create)))
	// mux.Handle("GET /v1/guilds/{guildId}", withGuild(circleHandler.Get))
	// ... etc

	// SSE events endpoint - simplified without guild access for now
	mux.Handle("GET /v1/events/stream", authMiddleware(http.HandlerFunc(eventsHandler.Stream)))
	_ = eventsHandler

	// Profile endpoints (auth required)
	mux.Handle("GET /v1/profile", authMiddleware(http.HandlerFunc(profileHandler.Get)))
	mux.Handle("PATCH /v1/profile", authMiddleware(http.HandlerFunc(profileHandler.Update)))
	mux.Handle("GET /v1/users/{userId}/profile", authMiddleware(http.HandlerFunc(profileHandler.GetUser)))
	mux.Handle("GET /v1/profiles/nearby", authMiddleware(http.HandlerFunc(profileHandler.GetNearby)))

	// Discovery endpoints (global people matching)
	mux.Handle("GET /v1/discover/people", authMiddleware(http.HandlerFunc(discoveryHandler.DiscoverPeople)))
	mux.Handle("GET /v1/discover/interest/{interestId}", authMiddleware(http.HandlerFunc(discoveryHandler.DiscoverByInterest)))
	mux.Handle("GET /v1/discover/teach-learn", authMiddleware(http.HandlerFunc(discoveryHandler.DiscoverTeachLearn)))
	mux.HandleFunc("GET /v1/discover/hangout-types", discoveryHandler.GetHangoutTypes)

	// Interest endpoints (public and auth)
	mux.HandleFunc("GET /v1/interests", interestHandler.ListInterests)
	mux.HandleFunc("GET /v1/interests/categories", interestHandler.GetCategories)
	mux.Handle("GET /v1/profile/interests", authMiddleware(http.HandlerFunc(interestHandler.GetUserInterests)))
	mux.Handle("POST /v1/profile/interests", authMiddleware(http.HandlerFunc(interestHandler.AddUserInterest)))
	mux.Handle("PATCH /v1/profile/interests/{interestId}", authMiddleware(http.HandlerFunc(interestHandler.UpdateUserInterest)))
	mux.Handle("DELETE /v1/profile/interests/{interestId}", authMiddleware(http.HandlerFunc(interestHandler.RemoveUserInterest)))
	mux.Handle("GET /v1/profile/interests/stats", authMiddleware(http.HandlerFunc(interestHandler.GetInterestStats)))
	mux.Handle("GET /v1/interests/matches/teaching", authMiddleware(http.HandlerFunc(interestHandler.FindTeachingMatches)))
	mux.Handle("GET /v1/interests/matches/learning", authMiddleware(http.HandlerFunc(interestHandler.FindLearningMatches)))
	mux.Handle("GET /v1/interests/shared", authMiddleware(http.HandlerFunc(interestHandler.FindSharedInterests)))

	// TODO: Questionnaire endpoints - needs questionnaireHandler
	// mux.HandleFunc("GET /v1/questions", questionnaireHandler.ListQuestions)
	// ... etc

	// Availability endpoints
	mux.HandleFunc("GET /v1/hangout-types", availabilityHandler.GetHangoutTypes)
	mux.Handle("POST /v1/availability", authMiddleware(http.HandlerFunc(availabilityHandler.CreateAvailability)))
	mux.Handle("GET /v1/availability/{availabilityId}", authMiddleware(http.HandlerFunc(availabilityHandler.GetAvailability)))
	mux.Handle("PATCH /v1/availability/{availabilityId}", authMiddleware(http.HandlerFunc(availabilityHandler.UpdateAvailability)))
	mux.Handle("DELETE /v1/availability/{availabilityId}", authMiddleware(http.HandlerFunc(availabilityHandler.DeleteAvailability)))
	mux.Handle("GET /v1/profile/availability", authMiddleware(http.HandlerFunc(availabilityHandler.GetMyAvailabilities)))
	mux.Handle("GET /v1/discover/availability", authMiddleware(http.HandlerFunc(availabilityHandler.FindNearby)))
	mux.Handle("GET /v1/discover/availability/type/{type}", authMiddleware(http.HandlerFunc(availabilityHandler.FindByType)))
	mux.Handle("POST /v1/availability/{availabilityId}/request", authMiddleware(http.HandlerFunc(availabilityHandler.RequestHangout)))
	mux.Handle("GET /v1/availability/{availabilityId}/requests", authMiddleware(http.HandlerFunc(availabilityHandler.GetPendingRequests)))
	mux.Handle("POST /v1/requests/{requestId}/respond", authMiddleware(http.HandlerFunc(availabilityHandler.RespondToRequest)))
	mux.Handle("GET /v1/profile/hangouts", authMiddleware(http.HandlerFunc(availabilityHandler.GetUserHangouts)))
	mux.Handle("PATCH /v1/hangouts/{hangoutId}/status", authMiddleware(http.HandlerFunc(availabilityHandler.UpdateHangoutStatus)))

	// Resonance endpoints
	mux.Handle("GET /v1/resonance", authMiddleware(http.HandlerFunc(resonanceHandler.GetMyResonance)))
	mux.Handle("GET /v1/resonance/ledger", authMiddleware(http.HandlerFunc(resonanceHandler.GetLedger)))
	mux.Handle("POST /v1/resonance/recalculate", authMiddleware(http.HandlerFunc(resonanceHandler.RecalculateScore)))
	mux.HandleFunc("GET /v1/resonance/explain", resonanceHandler.GetResonanceExplainer)
	mux.Handle("GET /v1/users/{userId}/resonance", authMiddleware(http.HandlerFunc(resonanceHandler.GetUserResonance)))

	// Review endpoints
	mux.Handle("POST /v1/reviews", authMiddleware(http.HandlerFunc(reviewHandler.CreateReview)))
	mux.Handle("GET /v1/reviews/{reviewId}", authMiddleware(http.HandlerFunc(reviewHandler.GetReview)))
	mux.Handle("GET /v1/profile/reviews/given", authMiddleware(http.HandlerFunc(reviewHandler.GetReviewsGiven)))
	mux.Handle("GET /v1/profile/reviews/received", authMiddleware(http.HandlerFunc(reviewHandler.GetReviewsReceived)))
	mux.Handle("GET /v1/profile/reputation", authMiddleware(http.HandlerFunc(reviewHandler.GetMyReputation)))
	mux.Handle("GET /v1/users/{userId}/reputation", authMiddleware(http.HandlerFunc(reviewHandler.GetUserReputation)))
	mux.HandleFunc("GET /v1/reviews/tags/positive", reviewHandler.GetPositiveTags)
	mux.HandleFunc("GET /v1/reviews/tags/improvement", reviewHandler.GetImprovementTags)

	// TODO: Event endpoints - needs eventHandler
	// mux.Handle("POST /v1/events", authMiddleware(http.HandlerFunc(eventHandler.CreateEvent)))
	// ... etc

	// TODO: Event role endpoints - needs eventRoleHandler
	// mux.Handle("POST /v1/events/{eventId}/roles", authMiddleware(http.HandlerFunc(eventRoleHandler.CreateRole)))
	// ... etc

	// Trust endpoints
	mux.Handle("GET /v1/trust", authMiddleware(http.HandlerFunc(trustHandler.GetTrustedUsers)))
	mux.Handle("GET /v1/trust/{userId}", authMiddleware(http.HandlerFunc(trustHandler.GetTrustSummary)))
	mux.Handle("POST /v1/trust/{userId}", authMiddleware(http.HandlerFunc(trustHandler.GrantTrust)))
	mux.Handle("DELETE /v1/trust/{userId}", authMiddleware(http.HandlerFunc(trustHandler.RevokeTrust)))
	mux.Handle("GET /v1/profile/trust", authMiddleware(http.HandlerFunc(trustHandler.GetTrustProfile)))
	mux.Handle("GET /v1/irl", authMiddleware(http.HandlerFunc(trustHandler.GetIRLConnections)))
	mux.Handle("POST /v1/irl/{userId}", authMiddleware(http.HandlerFunc(trustHandler.ConfirmIRL)))

	// TODO: Rideshare endpoints (renamed from Commute) - needs rideshareHandler
	// mux.Handle("GET /v1/rideshares", authMiddleware(http.HandlerFunc(rideshareHandler.GetUserRideshares)))
	// ... etc

	// TODO: Pool endpoints - needs poolHandler and guild service
	// mux.Handle("GET /v1/guilds/{guildId}/pools", withGuild(poolHandler.ListPools))
	// ... etc

	// Trust Rating endpoints
	mux.Handle("POST /v1/trust-ratings", authMiddleware(http.HandlerFunc(trustRatingHandler.Create)))
	mux.Handle("GET /v1/trust-ratings/{ratingId}", authMiddleware(http.HandlerFunc(trustRatingHandler.GetByID)))
	mux.Handle("PATCH /v1/trust-ratings/{ratingId}", authMiddleware(http.HandlerFunc(trustRatingHandler.Update)))
	mux.Handle("DELETE /v1/trust-ratings/{ratingId}", authMiddleware(http.HandlerFunc(trustRatingHandler.Delete)))
	mux.Handle("GET /v1/users/{userId}/trust-ratings/received", authMiddleware(http.HandlerFunc(trustRatingHandler.GetReceivedRatings)))
	mux.Handle("GET /v1/users/{userId}/trust-ratings/given", authMiddleware(http.HandlerFunc(trustRatingHandler.GetGivenRatings)))
	mux.Handle("GET /v1/users/{userId}/trust-aggregate", authMiddleware(http.HandlerFunc(trustRatingHandler.GetAggregate)))
	mux.Handle("POST /v1/trust-ratings/{ratingId}/endorsements", authMiddleware(http.HandlerFunc(trustRatingHandler.CreateEndorsement)))
	mux.Handle("GET /v1/trust-ratings/{ratingId}/endorsements", authMiddleware(http.HandlerFunc(trustRatingHandler.GetEndorsements)))
	mux.Handle("GET /v1/admin/distrust-signals", authMiddleware(http.HandlerFunc(trustRatingHandler.GetDistrustSignals)))

	// Role Catalog endpoints - Guild catalogs
	mux.Handle("GET /v1/guilds/{guildId}/role-catalogs", authMiddleware(http.HandlerFunc(roleCatalogHandler.GetGuildCatalogs)))
	mux.Handle("POST /v1/guilds/{guildId}/role-catalogs", authMiddleware(http.HandlerFunc(roleCatalogHandler.CreateGuildCatalog)))
	// Role Catalog endpoints - User catalogs
	mux.Handle("GET /v1/users/me/role-catalogs", authMiddleware(http.HandlerFunc(roleCatalogHandler.GetUserCatalogs)))
	mux.Handle("POST /v1/users/me/role-catalogs", authMiddleware(http.HandlerFunc(roleCatalogHandler.CreateUserCatalog)))
	// Role Catalog endpoints - Common
	mux.Handle("GET /v1/role-catalogs/{catalogId}", authMiddleware(http.HandlerFunc(roleCatalogHandler.GetCatalogByID)))
	mux.Handle("PATCH /v1/role-catalogs/{catalogId}", authMiddleware(http.HandlerFunc(roleCatalogHandler.UpdateCatalog)))
	mux.Handle("DELETE /v1/role-catalogs/{catalogId}", authMiddleware(http.HandlerFunc(roleCatalogHandler.DeleteCatalog)))
	// Rideshare role endpoints
	mux.Handle("GET /v1/rideshares/{rideshareId}/roles", authMiddleware(http.HandlerFunc(roleCatalogHandler.GetRideshareRoles)))
	mux.Handle("POST /v1/rideshares/{rideshareId}/roles", authMiddleware(http.HandlerFunc(roleCatalogHandler.CreateRideshareRole)))
	mux.Handle("GET /v1/rideshares/{rideshareId}/roles/detailed", authMiddleware(http.HandlerFunc(roleCatalogHandler.GetRideshareRolesWithAssignments)))
	mux.Handle("PATCH /v1/rideshares/{rideshareId}/roles/{roleId}", authMiddleware(http.HandlerFunc(roleCatalogHandler.UpdateRideshareRole)))
	mux.Handle("DELETE /v1/rideshares/{rideshareId}/roles/{roleId}", authMiddleware(http.HandlerFunc(roleCatalogHandler.DeleteRideshareRole)))
	mux.Handle("POST /v1/rideshares/{rideshareId}/roles/assign", authMiddleware(http.HandlerFunc(roleCatalogHandler.AssignRideshareRole)))
	mux.Handle("DELETE /v1/rideshares/{rideshareId}/roles/assignments/{assignmentId}", authMiddleware(http.HandlerFunc(roleCatalogHandler.UnassignRideshareRole)))
	mux.Handle("GET /v1/rideshares/{rideshareId}/my-roles", authMiddleware(http.HandlerFunc(roleCatalogHandler.GetUserRideshareRoles)))

	// Adventure endpoints
	mux.Handle("POST /v1/adventures", authMiddleware(http.HandlerFunc(adventureHandler.Create)))
	mux.Handle("GET /v1/adventures/{adventureId}", authMiddleware(http.HandlerFunc(adventureHandler.GetByID)))
	mux.Handle("POST /v1/guilds/{guildId}/adventures", authMiddleware(http.HandlerFunc(adventureHandler.CreateGuildAdventure)))
	mux.Handle("POST /v1/users/me/adventures", authMiddleware(http.HandlerFunc(adventureHandler.CreateUserAdventure)))
	// Adventure admission endpoints
	mux.Handle("POST /v1/adventures/{adventureId}/admission/request", authMiddleware(http.HandlerFunc(adventureHandler.RequestAdmission)))
	mux.Handle("GET /v1/adventures/{adventureId}/admission", authMiddleware(http.HandlerFunc(adventureHandler.GetAdmission)))
	mux.Handle("DELETE /v1/adventures/{adventureId}/admission", authMiddleware(http.HandlerFunc(adventureHandler.WithdrawAdmission)))
	mux.Handle("GET /v1/adventures/{adventureId}/admitted", authMiddleware(http.HandlerFunc(adventureHandler.CheckAdmission)))
	// Adventure admission management
	mux.Handle("GET /v1/adventures/{adventureId}/admissions", authMiddleware(http.HandlerFunc(adventureHandler.GetAdmissions)))
	mux.Handle("GET /v1/adventures/{adventureId}/admissions/pending", authMiddleware(http.HandlerFunc(adventureHandler.GetPendingAdmissions)))
	mux.Handle("POST /v1/adventures/{adventureId}/admissions/{userId}/respond", authMiddleware(http.HandlerFunc(adventureHandler.RespondToAdmission)))
	mux.Handle("POST /v1/adventures/{adventureId}/admissions/invite", authMiddleware(http.HandlerFunc(adventureHandler.InviteToAdventure)))
	// Adventure organizer management
	mux.Handle("POST /v1/adventures/{adventureId}/transfer", authMiddleware(http.HandlerFunc(adventureHandler.TransferAdventure)))
	mux.Handle("POST /v1/adventures/{adventureId}/unfreeze", authMiddleware(http.HandlerFunc(adventureHandler.UnfreezeAdventure)))

	// Vote endpoints
	mux.Handle("POST /v1/votes", authMiddleware(http.HandlerFunc(voteHandler.Create)))
	mux.Handle("GET /v1/votes/{voteId}", authMiddleware(http.HandlerFunc(voteHandler.GetByID)))
	mux.Handle("PATCH /v1/votes/{voteId}", authMiddleware(http.HandlerFunc(voteHandler.Update)))
	mux.Handle("DELETE /v1/votes/{voteId}", authMiddleware(http.HandlerFunc(voteHandler.Delete)))
	mux.Handle("POST /v1/votes/{voteId}/open", authMiddleware(http.HandlerFunc(voteHandler.Open)))
	mux.Handle("POST /v1/votes/{voteId}/close", authMiddleware(http.HandlerFunc(voteHandler.Close)))
	mux.Handle("POST /v1/votes/{voteId}/cancel", authMiddleware(http.HandlerFunc(voteHandler.Cancel)))
	// Vote option endpoints
	mux.Handle("GET /v1/votes/{voteId}/options", authMiddleware(http.HandlerFunc(voteHandler.GetOptions)))
	mux.Handle("POST /v1/votes/{voteId}/options", authMiddleware(http.HandlerFunc(voteHandler.CreateOption)))
	mux.Handle("POST /v1/votes/{voteId}/options/batch", authMiddleware(http.HandlerFunc(voteHandler.BatchCreateOptions)))
	mux.Handle("PATCH /v1/votes/{voteId}/options/{optionId}", authMiddleware(http.HandlerFunc(voteHandler.UpdateOption)))
	mux.Handle("DELETE /v1/votes/{voteId}/options/{optionId}", authMiddleware(http.HandlerFunc(voteHandler.DeleteOption)))
	// Vote ballot endpoints
	mux.Handle("POST /v1/votes/{voteId}/ballot", authMiddleware(http.HandlerFunc(voteHandler.CastBallot)))
	mux.Handle("GET /v1/votes/{voteId}/ballot", authMiddleware(http.HandlerFunc(voteHandler.GetMyBallot)))
	mux.Handle("GET /v1/votes/{voteId}/ballots", authMiddleware(http.HandlerFunc(voteHandler.GetBallots)))
	// Vote results endpoints
	mux.Handle("GET /v1/votes/{voteId}/results", authMiddleware(http.HandlerFunc(voteHandler.GetResults)))
	mux.Handle("GET /v1/votes/{voteId}/stats", authMiddleware(http.HandlerFunc(voteHandler.GetVoteStats)))
	// Vote scoped query endpoints
	mux.Handle("GET /v1/guilds/{guildId}/votes", authMiddleware(http.HandlerFunc(voteHandler.GetGuildVotes)))
	mux.Handle("GET /v1/votes/global", authMiddleware(http.HandlerFunc(voteHandler.GetGlobalVotes)))

	// Moderation endpoints
	moderationHandler.RegisterRoutes(mux)

	// Apply global middleware
	wrapped := middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.Logger,
		middleware.Recovery,
		middleware.CORS(cfg.Server.AllowedOrigins),
		middleware.RateLimit(rateLimiter),
		middleware.Idempotency(idempotencyStore),
		middleware.Compress,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      wrapped,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("starting server",
			slog.String("port", cfg.Server.Port),
			slog.String("env", cfg.Server.Env),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", slog.String("error", err.Error()))
	}

	slog.Info("server exited")
}
