package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	// Import the pq driver so that it can register itself with the database/sql
	// package. Note that we alias this import to the blank identifier, to stop the Go
	// compiler complaining that the package isn't being used.
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"movieapi.sanjayojha.dev/internal/data"
	"movieapi.sanjayojha.dev/internal/mailer"
	"movieapi.sanjayojha.dev/internal/vcs"
)

var (
	version = vcs.Version()
)

type config struct {
	port int
	env  string
	db   struct {
		user         string
		password     string
		name         string
		host         string
		port         string
		sslmode      string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	smtp struct {
		host     string
		port     string
		username string
		password string
		sender   string
	}

	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer *mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	sHandler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(sHandler)

	err := godotenv.Load()

	if err != nil {
		logger.Warn("No .env file found, reading from system environments")
	}

	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Enviornment (development|staging|production)")

	flag.StringVar(&cfg.db.user, "db-user", os.Getenv("DB_USER"), "Database Username")
	// password is myGreenP@ss. We have to URL encode special character like @, :, /,?. # etc. For @ we are using %40
	flag.StringVar(&cfg.db.password, "db-password", os.Getenv("DB_PASSWORD"), "Database Password")
	flag.StringVar(&cfg.db.name, "db-name", os.Getenv("DB_NAME"), "Database Name")
	flag.StringVar(&cfg.db.host, "db-host", os.Getenv("DB_HOST"), "Database Host")
	flag.StringVar(&cfg.db.port, "db-port", os.Getenv("DB_PORT"), "Database Port")
	flag.StringVar(&cfg.db.sslmode, "db-sslmode", os.Getenv("DB_SSLMODE"), "Database SSL mode (enable|disable)")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 20, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// limiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// smtp
	// We are using mailpit container
	flag.StringVar(&cfg.smtp.port, "smtp-port", os.Getenv("SMTP_PORT"), "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")

	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight API <no-reply@greenlight.sanjayojha.dev>", "SMTP sender")

	//setting cors origin
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	// Display version
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	// Basic validation to catch missing configuration early
	if cfg.db.user == "" || cfg.db.password == "" || cfg.db.host == "" {
		logger.Error("Critical database environment variables are missing")
		os.Exit(1)
	}

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the main() function exits.
	defer db.Close()

	logger.Info("database connection pool established")

	// mailer
	mailer, err := mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer,
	}

	// mux := http.NewServeMux()
	// mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {

	// Build the safe DSN string
	dsn := buildDSN(cfg)

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close() // close the connection if error
		return nil, err
	}

	return db, nil
}

func buildDSN(cfg config) string {
	// Format host and port correctly
	hostStr := fmt.Sprintf("%s:%s", cfg.db.host, cfg.db.port)

	// create URL object
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.db.user, cfg.db.password),
		Host:   hostStr,
		Path:   cfg.db.name,
	}

	// Add SSL mode query parameters (e.g., ?sslmode=disable)
	q := u.Query()
	q.Set("sslmode", cfg.db.sslmode)
	u.RawQuery = q.Encode()

	return u.String()
}
