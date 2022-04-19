package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kudarap/ghsearch"
	"github.com/kudarap/ghsearch/github"
	"github.com/kudarap/ghsearch/http"
	"github.com/kudarap/ghsearch/logging"
	"github.com/kudarap/ghsearch/redis"
)

func main() {
	app := newApp()
	if err := app.conf.loadFromEnv(); err != nil {
		app.log.Errorf("could not load config: %s", err)
		return
	}

	if err := app.setup(); err != nil {
		app.log.Errorf("could not setup app: %s", err)
		return
	}
	defer app.close()

	if err := app.run(); err != nil {
		app.log.Errorf("could not run app: %s", err)
	}
}

type Application struct {
	conf    Config
	log     *logging.Logger
	closeFn func() error

	server *http.Server
}

func (app *Application) setup() error {
	// Initialize dependencies
	githubClient, err := github.NewClient(app.conf.GithubToken)
	if err != nil {
		return fmt.Errorf("could not setup github: %s", err)
	}
	redisClient, err := redis.NewClient(app.conf.RedisURL)
	if err != nil {
		return fmt.Errorf("could not setup redis: %s", err)
	}

	userSourceCache := redis.NewUserSource(redisClient, githubClient)
	userService := ghsearch.NewUserService(userSourceCache)

	restHandler := http.NewRestHandler(userService)
	srv := http.NewServer(app.conf.Addr, restHandler, app.log)
	if err = srv.Run(); err != nil {
		return err
	}

	app.closeFn = func() error { return redisClient.Close() }
	return nil
}

func (app *Application) run() error {
	return app.server.Run()
}

func (app *Application) close() {
	if err := app.closeFn(); err != nil {
		app.log.Errorln("error when closing app: %s", err)
	}
}

func newApp() *Application {
	var a Application
	a.log = logging.New()
	return &a
}

type Config struct {
	Addr        string
	RedisURL    string
	GithubToken string
}

func (c *Config) loadFromEnv() error {
	// Loads env vars from .env file.
	err := godotenv.Load()
	if err != nil {
		return err
	}

	c.Addr = os.Getenv("ADDR")
	c.RedisURL = os.Getenv("REDIS_URL")
	c.GithubToken = os.Getenv("GITHUB_TOKEN")
	return nil
}
