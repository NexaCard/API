package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/NexaCard/API/internal/admincmd"
	"github.com/NexaCard/API/internal/app"
	"github.com/NexaCard/API/internal/config"
	"github.com/NexaCard/API/internal/logger"
	"github.com/NexaCard/API/internal/models"
	"github.com/NexaCard/API/internal/version"
	"github.com/NexaCard/API/internal/web"

	"github.com/gin-gonic/gin"
)

const (
	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiDim       = "\033[2m"
	ansiGreen     = "\033[32m"
	ansiBlue      = "\033[34m"
	ansiCyan      = "\033[36m"
	ansiBrightMag = "\033[95m"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "admin" {
		runAdminSubcommand(os.Args[2:])
		return
	}

	printStartupBanner()

	cfg := config.Load()
	logger.Init(cfg.Server.Mode, cfg.Log.ToLoggerOptions())
	stdLog := logger.StdLogger()

	if cfg.Server.Mode == "release" {
		if err := validateProductionConfig(cfg); err != nil {
			stdLog.Fatalf("production configuration error: %v", err)
		}
	} else if weakName, ok := firstWeakProductionSecret(cfg); ok {
		stdLog.Printf("Warning: %s is weak or still uses a default value; replace it before production use", weakName)
	}

	if web.Enabled() {
		fmt.Println(ansiGreen + "Embedded SPAs: admin (" + cfg.Web.AdminPath + "), user (/)" + ansiReset)
	}

	if web.Enabled() && cfg.Server.Mode == "release" && cfg.Web.AdminPath == "/admin" {
		stdLog.Printf("Warning: web.admin_path is still /admin in release mode; use a less predictable admin path to reduce automated scan risk")
	}

	for _, dir := range []string{"db", "uploads", "logs"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			stdLog.Printf("Warning: create directory %s failed: %v", dir, err)
		}
	}

	if err := models.InitDB(cfg.Database.Driver, cfg.Database.DSN, models.DBPoolConfig{
		MaxOpenConns:           cfg.Database.Pool.MaxOpenConns,
		MaxIdleConns:           cfg.Database.Pool.MaxIdleConns,
		ConnMaxLifetimeSeconds: cfg.Database.Pool.ConnMaxLifetimeSeconds,
		ConnMaxIdleTimeSeconds: cfg.Database.Pool.ConnMaxIdleTimeSeconds,
	}); err != nil {
		stdLog.Fatalf("database initialization failed: %v", err)
	}

	if err := models.AutoMigrate(); err != nil {
		stdLog.Fatalf("database migration failed: %v", err)
	}

	defaultAdminUser, defaultAdminPass := resolveDefaultAdminCredentials(cfg)
	if cfg.Server.Mode == "release" && defaultAdminPass == "" {
		stdLog.Printf("Warning: NEXACARD_DEFAULT_ADMIN_PASSWORD and bootstrap.default_admin_password are empty in release mode; default admin initialization has been skipped")
	} else if err := models.InitDefaultAdmin(defaultAdminUser, defaultAdminPass); err != nil {
		stdLog.Printf("Warning: initialize default admin failed: %v", err)
	}

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	var mode string
	flag.StringVar(&mode, "mode", app.ModeAll, "startup mode: all (default), api, worker")
	flag.Parse()

	if err := app.Run(app.Options{
		Config:  cfg,
		Logger:  logger.S(),
		Signals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		Mode:    mode,
	}); err != nil {
		stdLog.Fatalf("service runtime failed: %v", err)
	}
}

func printStartupBanner() {
	fmt.Println(ansiBrightMag + "==============================================================" + ansiReset)
	fmt.Println(ansiBrightMag + "                     NexaCard API starting                    " + ansiReset)
	fmt.Println(ansiBrightMag + "==============================================================" + ansiReset)
	fmt.Println(ansiCyan + "Brand:   NexaCard" + ansiReset)
	fmt.Println(ansiCyan + "API:     https://github.com/NexaCard/API" + ansiReset)
	fmt.Println(ansiCyan + "User:    https://github.com/NexaCard/user" + ansiReset)
	fmt.Println(ansiCyan + "Admin:   https://github.com/NexaCard/admin" + ansiReset)
	fmt.Println(ansiGreen + "Version: " + version.Version + ansiReset)
	fmt.Println(ansiDim + "--------------------------------------------------------------" + ansiReset)
}

func validateProductionConfig(cfg *config.Config) error {
	if weakName, ok := firstWeakProductionSecret(cfg); ok {
		return fmt.Errorf("%s is weak or still uses a default value; configure a strong random secret before running in production", weakName)
	}
	if hasWildcardOrigin(cfg.CORS.AllowedOrigins) {
		return fmt.Errorf("cors.allowed_origins contains '*'; set explicit storefront/admin origins before running in release mode")
	}
	return nil
}

func isWeakSecret(secret string) bool {
	if len(secret) < 32 {
		return true
	}
	normalized := strings.ToLower(secret)
	if strings.Contains(normalized, "change-me") ||
		strings.Contains(normalized, "change-in-production") ||
		strings.Contains(normalized, "your-secret-key") {
		return true
	}
	return false
}

func firstWeakProductionSecret(cfg *config.Config) (string, bool) {
	if cfg == nil {
		return "config", true
	}
	secrets := []struct {
		name  string
		value string
	}{
		{name: "app.secret_key", value: cfg.App.SecretKey},
		{name: "jwt.secret", value: cfg.JWT.SecretKey},
		{name: "user_jwt.secret", value: cfg.UserJWT.SecretKey},
	}
	for _, secret := range secrets {
		if isWeakSecret(secret.value) {
			return secret.name, true
		}
	}
	return "", false
}

func hasWildcardOrigin(origins []string) bool {
	for _, origin := range origins {
		if strings.TrimSpace(origin) == "*" {
			return true
		}
	}
	return false
}

func runAdminSubcommand(args []string) {
	cfg := config.Load()
	if err := models.InitDB(cfg.Database.Driver, cfg.Database.DSN, models.DBPoolConfig{
		MaxOpenConns:           cfg.Database.Pool.MaxOpenConns,
		MaxIdleConns:           cfg.Database.Pool.MaxIdleConns,
		ConnMaxLifetimeSeconds: cfg.Database.Pool.ConnMaxLifetimeSeconds,
		ConnMaxIdleTimeSeconds: cfg.Database.Pool.ConnMaxIdleTimeSeconds,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "init db: %v\n", err)
		os.Exit(1)
	}
	admincmd.Run(args)
}

func resolveDefaultAdminCredentials(cfg *config.Config) (string, string) {
	user := strings.TrimSpace(os.Getenv("NEXACARD_DEFAULT_ADMIN_USERNAME"))
	pass := strings.TrimSpace(os.Getenv("NEXACARD_DEFAULT_ADMIN_PASSWORD"))
	if user == "" {
		user = strings.TrimSpace(os.Getenv("DJ_DEFAULT_ADMIN_USERNAME"))
	}
	if pass == "" {
		pass = strings.TrimSpace(os.Getenv("DJ_DEFAULT_ADMIN_PASSWORD"))
	}
	if cfg == nil {
		return user, pass
	}
	if user == "" {
		user = strings.TrimSpace(cfg.Bootstrap.DefaultAdminUsername)
	}
	if pass == "" {
		pass = strings.TrimSpace(cfg.Bootstrap.DefaultAdminPassword)
	}
	return user, pass
}
