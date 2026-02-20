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

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", slog.String("error", err.Error()))
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
	defer func() { _ = db.Close() }()

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
	guildRepo := repository.NewGuildRepository(db)
	memberRepo := repository.NewMemberRepository(db)
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
	moderationRepo := repository.NewModerationRepository(db)
	deviceTokenRepo := repository.NewDeviceTokenRepository(db)

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
				ClientID:     cfg.OAuth.Google.ClientID,
				ClientSecret: cfg.OAuth.Google.ClientSecret,
				RedirectURI:  cfg.OAuth.Google.RedirectURI,
			},
			Apple: service.AppleOAuthConfig{
				ClientID:    cfg.OAuth.Apple.ClientID,
				TeamID:      cfg.OAuth.Apple.TeamID,
				KeyID:       cfg.OAuth.Apple.KeyID,
				PrivateKey:  cfg.OAuth.Apple.PrivateKey,
				RedirectURI: cfg.OAuth.Apple.RedirectURI,
			},
		},
		AuthService:  authService,
		IdentityRepo: identityRepo,
		UserRepo:     userRepo,
		TokenService: tokenService,
	})

	passkeyService := service.NewPasskeyService(service.PasskeyServiceConfig{
		Config: service.PasskeyConfig{
			RPID:            cfg.Passkey.RPID,
			RPName:          cfg.Passkey.RPName,
			RPOrigins:       cfg.Passkey.RPOrigins,
			Timeout:         cfg.Passkey.Timeout,
			RequireUV:       cfg.Passkey.RequireUV,
			AttestationType: cfg.Passkey.AttestationType,
		},
		PasskeyRepo:  passkeyRepo,
		UserRepo:     userRepo,
		TokenService: tokenService,
	})

	guildService := service.NewGuildService(service.GuildServiceConfig{
		GuildRepo:  guildRepo,
		MemberRepo: memberRepo,
		UserRepo:   userRepo,
	})

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
		GuildRepo:     guildRepo,
	})

	voteService := service.NewVoteService(service.VoteServiceConfig{
		VoteRepo:  voteRepo,
		GuildRepo: guildRepo,
	})

	adventureService := service.NewAdventureService(service.AdventureServiceConfig{
		AdventureRepo: adventureRepo,
		AdmissionRepo: adventureAdmissionRepo,
		GuildRepo:     guildRepo,
	})

	eventRoleService := service.NewEventRoleService(eventRoleRepo, interestService)

	eventService := service.NewEventService(eventRepo, compatibilityService, questionnaireService, eventRoleService)

	// TODO: Implement Rideshare service (renamed from Commute)
	// commuteService := service.NewCommuteService(commuteRepo, trustService)

	poolService := service.NewPoolService(service.PoolServiceConfig{
		PoolRepo:      poolRepo,
		GuildRepo:     guildRepo,
		MemberRepo:    memberRepo,
		Compatibility: compatibilityService,
	})

	discoveryService := service.NewDiscoveryService(service.DiscoveryServiceConfig{
		AvailabilityRepo:  availabilityRepo,
		CompatibilityRepo: questionnaireRepo,
		InterestRepo:      interestRepo,
		ProfileRepo:       profileRepo,
	})

	// Initialize seeder service for admin tools
	seederService := service.NewSeederService(db)

	// Initialize admin actions service (will be connected to eventHub after it's created)
	var adminActionsService *service.AdminActionsService

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

	// Initialize admin actions service (now that eventHub exists)
	adminActionsService = service.NewAdminActionsService(db, eventHub)

	// Initialize moderation service
	moderationService := service.NewModerationService(moderationRepo, eventHub)

	// Initialize admin users service
	adminUsersService := service.NewAdminUsersService(db, userRepo, profileRepo, moderationService)

	// Initialize admin discovery service
	adminDiscoveryService := service.NewAdminDiscoveryService(db, discoveryService, compatibilityService)

	poolMatcher := jobs.NewPoolMatcher(poolService, 1*time.Hour)
	poolMatcher.Start()
	defer poolMatcher.Stop()

	// Initialize push notification service
	pushService, err := service.NewPushService(service.PushServiceConfig{
		DeviceRepo:         deviceTokenRepo,
		Enabled:            cfg.Push.Enabled,
		FCMCredentialsPath: cfg.Push.FCMCredentialsPath,
	})
	if err != nil {
		slog.Error("Failed to initialize push service", "error", err)
		// Continue without push - it's optional
		pushService = nil
	}

	// Initialize nudge service and processor
	nudgeService := service.NewNudgeService(service.NudgeServiceConfig{
		AvailabilityRepo: availabilityRepo,
		PoolRepo:         poolRepo,
		EventHub:         eventHub,
		PushService:      pushService,
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
	guildHandler := handler.NewGuildHandler(guildService)
	// TODO: Implement Person, Activity, Timer handlers
	// personHandler := handler.NewPersonHandler(guildService, eventHub)
	// activityHandler := handler.NewActivityHandler(guildService, eventHub)
	// timerHandler := handler.NewTimerHandler(guildService, eventHub)
	eventsHandler := handler.NewEventsHandler(eventHub)
	profileHandler := handler.NewProfileHandler(profileService)
	interestHandler := handler.NewInterestHandler(interestService)
	questionnaireHandler := handler.NewQuestionnaireHandler(questionnaireService, compatibilityService)
	availabilityHandler := handler.NewAvailabilityHandler(availabilityService, profileService)
	resonanceHandler := handler.NewResonanceHandler(resonanceService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	eventHandler := handler.NewEventHandler(eventService)
	eventRoleHandler := handler.NewEventRoleHandler(eventRoleService)
	trustHandler := handler.NewTrustHandler(trustService)
	trustRatingHandler := handler.NewTrustRatingHandler(trustRatingService)
	roleCatalogHandler := handler.NewRoleCatalogHandler(roleCatalogService)
	voteHandler := handler.NewVoteHandler(voteService)
	adventureHandler := handler.NewAdventureHandler(adventureService)
	// TODO: Implement Rideshare handler (renamed from Commute)
	// commuteHandler := handler.NewCommuteHandler(commuteService)
	poolHandler := handler.NewPoolHandler(poolService, guildService)
	discoveryHandler := handler.NewDiscoveryHandler(discoveryService)
	moderationHandler := handler.NewModerationHandler(moderationService, userRepo)
	deviceHandler := handler.NewDeviceHandler(deviceTokenRepo)
	adminSeederHandler := handler.NewAdminSeederHandler(seederService)
	adminActionsHandler := handler.NewAdminActionsHandler(adminActionsService)
	adminUsersHandler := handler.NewAdminUsersHandler(adminUsersService)
	adminDiscoveryHandler := handler.NewAdminDiscoveryHandler(adminDiscoveryService)

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
	adminMiddleware := middleware.AdminAuth(tokenService)
	mux.Handle("POST /v1/auth/logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /v1/auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))

	// Passkey registration endpoints (protected - user must be logged in)
	mux.Handle("POST /v1/auth/passkey/register/start", authMiddleware(http.HandlerFunc(passkeyHandler.RegisterStart)))
	mux.Handle("POST /v1/auth/passkey/register/finish", authMiddleware(http.HandlerFunc(passkeyHandler.RegisterFinish)))
	mux.Handle("DELETE /v1/auth/passkey/", authMiddleware(http.HandlerFunc(passkeyHandler.Delete)))

	// Guild endpoints
	mux.Handle("GET /v1/guilds", authMiddleware(http.HandlerFunc(guildHandler.List)))
	mux.Handle("POST /v1/guilds", authMiddleware(http.HandlerFunc(guildHandler.Create)))
	mux.Handle("GET /v1/guilds/{guildId}", authMiddleware(http.HandlerFunc(guildHandler.Get)))
	mux.Handle("PATCH /v1/guilds/{guildId}", authMiddleware(http.HandlerFunc(guildHandler.Update)))
	mux.Handle("DELETE /v1/guilds/{guildId}", authMiddleware(http.HandlerFunc(guildHandler.Delete)))
	mux.Handle("POST /v1/guilds/{guildId}/join", authMiddleware(http.HandlerFunc(guildHandler.Join)))
	mux.Handle("POST /v1/guilds/{guildId}/leave", authMiddleware(http.HandlerFunc(guildHandler.Leave)))
	mux.Handle("GET /v1/guilds/{guildId}/members", authMiddleware(http.HandlerFunc(guildHandler.GetMembers)))
	mux.Handle("GET /v1/guilds/{guildId}/members/{userId}/role", authMiddleware(http.HandlerFunc(guildHandler.GetMemberRole)))
	mux.Handle("PATCH /v1/guilds/{guildId}/members/{userId}/role", authMiddleware(http.HandlerFunc(guildHandler.UpdateMemberRole)))

	// SSE events endpoint - simplified without guild access for now
	mux.Handle("GET /v1/events/stream", authMiddleware(http.HandlerFunc(eventsHandler.Stream)))
	_ = eventsHandler

	// Profile endpoints (auth required)
	mux.Handle("GET /v1/profile", authMiddleware(http.HandlerFunc(profileHandler.Get)))
	mux.Handle("PATCH /v1/profile", authMiddleware(http.HandlerFunc(profileHandler.Update)))
	mux.Handle("GET /v1/users/{userId}/profile", authMiddleware(http.HandlerFunc(profileHandler.GetUser)))
	mux.Handle("GET /v1/profiles/nearby", authMiddleware(http.HandlerFunc(profileHandler.GetNearby)))

	// Device token endpoints (for push notifications)
	mux.Handle("POST /v1/devices", authMiddleware(http.HandlerFunc(deviceHandler.Register)))
	mux.Handle("GET /v1/devices", authMiddleware(http.HandlerFunc(deviceHandler.List)))
	mux.Handle("DELETE /v1/devices/{deviceId}", authMiddleware(http.HandlerFunc(deviceHandler.Delete)))

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

	// Questionnaire endpoints (public)
	mux.HandleFunc("GET /v1/questions", questionnaireHandler.ListQuestions)
	mux.HandleFunc("GET /v1/questions/categories", questionnaireHandler.GetCategories)

	// Questionnaire endpoints (auth required)
	mux.Handle("GET /v1/questions/{questionId}", authMiddleware(http.HandlerFunc(questionnaireHandler.GetQuestion)))
	mux.Handle("GET /v1/profile/answers", authMiddleware(http.HandlerFunc(questionnaireHandler.GetUserAnswers)))
	mux.Handle("GET /v1/profile/answers/detailed", authMiddleware(http.HandlerFunc(questionnaireHandler.GetUserAnswersWithQuestions)))
	mux.Handle("GET /v1/profile/questions/progress", authMiddleware(http.HandlerFunc(questionnaireHandler.GetQuestionProgress)))
	mux.Handle("POST /v1/questions/{questionId}/answer", authMiddleware(http.HandlerFunc(questionnaireHandler.AnswerQuestion)))
	mux.Handle("PATCH /v1/questions/{questionId}/answer", authMiddleware(http.HandlerFunc(questionnaireHandler.UpdateAnswer)))
	mux.Handle("DELETE /v1/questions/{questionId}/answer", authMiddleware(http.HandlerFunc(questionnaireHandler.DeleteAnswer)))
	mux.Handle("GET /v1/compatibility/{userId}", authMiddleware(http.HandlerFunc(questionnaireHandler.GetCompatibility)))
	mux.Handle("GET /v1/compatibility/{userId}/yikes", authMiddleware(http.HandlerFunc(questionnaireHandler.GetYikesSummary)))

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

	// Event endpoints
	mux.Handle("POST /v1/events", authMiddleware(http.HandlerFunc(eventHandler.CreateEvent)))
	mux.Handle("GET /v1/events/{eventId}", authMiddleware(http.HandlerFunc(eventHandler.GetEvent)))
	mux.Handle("PATCH /v1/events/{eventId}", authMiddleware(http.HandlerFunc(eventHandler.UpdateEvent)))
	mux.Handle("POST /v1/events/{eventId}/cancel", authMiddleware(http.HandlerFunc(eventHandler.CancelEvent)))
	mux.Handle("POST /v1/events/{eventId}/rsvp", authMiddleware(http.HandlerFunc(eventHandler.RSVP)))
	mux.Handle("DELETE /v1/events/{eventId}/rsvp", authMiddleware(http.HandlerFunc(eventHandler.CancelRSVP)))
	mux.Handle("GET /v1/events/{eventId}/pending-rsvps", authMiddleware(http.HandlerFunc(eventHandler.GetPendingRSVPs)))
	mux.Handle("POST /v1/events/{eventId}/rsvps/{rsvpUserId}/respond", authMiddleware(http.HandlerFunc(eventHandler.RespondToRSVP)))
	mux.Handle("POST /v1/events/{eventId}/hosts", authMiddleware(http.HandlerFunc(eventHandler.AddHost)))
	mux.Handle("POST /v1/events/{eventId}/completion", authMiddleware(http.HandlerFunc(eventHandler.ConfirmCompletion)))
	mux.Handle("POST /v1/events/{eventId}/checkin", authMiddleware(http.HandlerFunc(eventHandler.Checkin)))
	mux.Handle("POST /v1/events/{eventId}/feedback", authMiddleware(http.HandlerFunc(eventHandler.SubmitFeedback)))
	mux.Handle("GET /v1/discover/events", authMiddleware(http.HandlerFunc(eventHandler.GetPublicEvents)))
	mux.Handle("GET /v1/guilds/{guildId}/events", authMiddleware(http.HandlerFunc(eventHandler.GetGuildEvents)))

	// Event role endpoints
	mux.Handle("POST /v1/events/{eventId}/roles", authMiddleware(http.HandlerFunc(eventRoleHandler.CreateRole)))
	mux.Handle("GET /v1/events/{eventId}/roles", authMiddleware(http.HandlerFunc(eventRoleHandler.GetRoles)))
	mux.Handle("GET /v1/events/{eventId}/roles/overview", authMiddleware(http.HandlerFunc(eventRoleHandler.GetRolesOverview)))
	mux.Handle("PATCH /v1/events/{eventId}/roles/{roleId}", authMiddleware(http.HandlerFunc(eventRoleHandler.UpdateRole)))
	mux.Handle("DELETE /v1/events/{eventId}/roles/{roleId}", authMiddleware(http.HandlerFunc(eventRoleHandler.DeleteRole)))
	mux.Handle("POST /v1/events/{eventId}/roles/assign", authMiddleware(http.HandlerFunc(eventRoleHandler.AssignRole)))
	mux.Handle("GET /v1/events/{eventId}/roles/mine", authMiddleware(http.HandlerFunc(eventRoleHandler.GetMyRoles)))
	mux.Handle("GET /v1/events/{eventId}/roles/suggestions", authMiddleware(http.HandlerFunc(eventRoleHandler.GetRoleSuggestions)))
	mux.Handle("DELETE /v1/events/{eventId}/roles/assignments/{assignmentId}", authMiddleware(http.HandlerFunc(eventRoleHandler.CancelAssignment)))

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

	// Pool endpoints (guild-scoped)
	mux.Handle("GET /v1/guilds/{guildId}/pools", authMiddleware(http.HandlerFunc(poolHandler.ListPools)))
	mux.Handle("POST /v1/guilds/{guildId}/pools", authMiddleware(http.HandlerFunc(poolHandler.CreatePool)))
	mux.Handle("GET /v1/guilds/{guildId}/pools/{poolId}", authMiddleware(http.HandlerFunc(poolHandler.GetPool)))
	mux.Handle("PATCH /v1/guilds/{guildId}/pools/{poolId}", authMiddleware(http.HandlerFunc(poolHandler.UpdatePool)))
	mux.Handle("DELETE /v1/guilds/{guildId}/pools/{poolId}", authMiddleware(http.HandlerFunc(poolHandler.DeletePool)))
	mux.Handle("POST /v1/guilds/{guildId}/pools/{poolId}/join", authMiddleware(http.HandlerFunc(poolHandler.JoinPool)))
	mux.Handle("POST /v1/guilds/{guildId}/pools/{poolId}/leave", authMiddleware(http.HandlerFunc(poolHandler.LeavePool)))
	mux.Handle("GET /v1/guilds/{guildId}/pools/{poolId}/members", authMiddleware(http.HandlerFunc(poolHandler.GetPoolMembers)))
	mux.Handle("PATCH /v1/guilds/{guildId}/pools/{poolId}/membership", authMiddleware(http.HandlerFunc(poolHandler.UpdateMembership)))
	mux.Handle("GET /v1/guilds/{guildId}/pools/{poolId}/stats", authMiddleware(http.HandlerFunc(poolHandler.GetPoolStats)))
	mux.Handle("GET /v1/guilds/{guildId}/pools/{poolId}/matches", authMiddleware(http.HandlerFunc(poolHandler.GetMatchHistory)))

	// Pool matching endpoints (user-scoped)
	mux.Handle("GET /v1/profile/matches/pending", authMiddleware(http.HandlerFunc(poolHandler.GetPendingMatches)))
	mux.Handle("PATCH /v1/matches/{matchId}", authMiddleware(http.HandlerFunc(poolHandler.UpdateMatch)))

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
	mux.Handle("GET /v1/admin/distrust-signals", adminMiddleware(http.HandlerFunc(trustRatingHandler.GetDistrustSignals)))

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
	mux.Handle("GET /v1/guilds/{guildId}/adventures", authMiddleware(http.HandlerFunc(adventureHandler.ListGuildAdventures)))
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

	// Admin seeder endpoints (for development/testing) - requires admin role
	mux.Handle("GET /v1/admin/seed/scenarios", adminMiddleware(http.HandlerFunc(adminSeederHandler.ListScenarios)))
	mux.Handle("POST /v1/admin/seed/users", adminMiddleware(http.HandlerFunc(adminSeederHandler.SeedUsers)))
	mux.Handle("POST /v1/admin/seed/guilds", adminMiddleware(http.HandlerFunc(adminSeederHandler.SeedGuilds)))
	mux.Handle("POST /v1/admin/seed/events", adminMiddleware(http.HandlerFunc(adminSeederHandler.SeedEvents)))
	mux.Handle("POST /v1/admin/seed/scenario", adminMiddleware(http.HandlerFunc(adminSeederHandler.SeedScenario)))
	mux.Handle("DELETE /v1/admin/seed/cleanup", adminMiddleware(http.HandlerFunc(adminSeederHandler.Cleanup)))

	// Admin user management endpoints - requires admin role
	mux.Handle("GET /v1/admin/users", adminMiddleware(http.HandlerFunc(adminUsersHandler.ListUsers)))
	mux.Handle("GET /v1/admin/users/{userId}", adminMiddleware(http.HandlerFunc(adminUsersHandler.GetUser)))
	mux.Handle("PATCH /v1/admin/users/{userId}/role", adminMiddleware(http.HandlerFunc(adminUsersHandler.UpdateRole)))
	mux.Handle("DELETE /v1/admin/users/{userId}", adminMiddleware(http.HandlerFunc(adminUsersHandler.DeleteUser)))

	// Admin discovery lab endpoints - requires admin role
	mux.Handle("GET /v1/admin/discovery/users", adminMiddleware(http.HandlerFunc(adminDiscoveryHandler.GetUsersWithLocations)))
	mux.Handle("POST /v1/admin/discovery/simulate", adminMiddleware(http.HandlerFunc(adminDiscoveryHandler.SimulateDiscovery)))
	mux.Handle("GET /v1/admin/discovery/compatibility/{userAId}/{userBId}", adminMiddleware(http.HandlerFunc(adminDiscoveryHandler.GetCompatibility)))

	// Admin action endpoints (for triggering events as users) - requires admin role
	mux.Handle("GET /v1/admin/actions/users", adminMiddleware(http.HandlerFunc(adminActionsHandler.GetUsers)))
	mux.Handle("GET /v1/admin/actions/guilds", adminMiddleware(http.HandlerFunc(adminActionsHandler.GetGuilds)))
	mux.Handle("GET /v1/admin/actions/events", adminMiddleware(http.HandlerFunc(adminActionsHandler.GetEvents)))
	mux.Handle("POST /v1/admin/actions/location", adminMiddleware(http.HandlerFunc(adminActionsHandler.UpdateLocation)))
	mux.Handle("POST /v1/admin/actions/trust-rating", adminMiddleware(http.HandlerFunc(adminActionsHandler.CreateTrustRating)))
	mux.Handle("POST /v1/admin/actions/guild-join", adminMiddleware(http.HandlerFunc(adminActionsHandler.JoinGuild)))
	mux.Handle("POST /v1/admin/actions/rsvp", adminMiddleware(http.HandlerFunc(adminActionsHandler.RSVP)))
	mux.Handle("POST /v1/admin/actions/event-create", adminMiddleware(http.HandlerFunc(adminActionsHandler.CreateEvent)))

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
