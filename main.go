package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	h "shortenerUrl/api"
	mr "shortenerUrl/repository/mongo"
	rr "shortenerUrl/repository/redis"
	"shortenerUrl/shortener"
)

/* 	https://www.google.com -> 98sj1-293
   	http://localhost:8000/98sj1-293 -> https://www.google.com

 	repo <- service -> serializer  -> http
*/



func main() {
	repo := chooseRepo()
	service := shortener.NewRedirectService(repo)
	handler := h.NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{code}", handler.Get)
	r.Post("/", handler.Post)

	errs := make(chan error, 2)
	go func() {

		fmt.Printf("Listening on port %s\n", httpPort())
		errs <- http.ListenAndServe(httpPort(), r)

	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("Terminated %s", <-errs)
}

func httpPort() string {
	port := goDotEnvVariable("PORT")
	if  port == "" {
		port = "8000"
	}

	return fmt.Sprintf(":%s", port)
}

func chooseRepo() shortener.RedirectRepository {

	typeDB := goDotEnvVariable("TYPE_DB")
	switch typeDB {
	case "redis":
		redisURL := goDotEnvVariable("REDIS_URL")
		repo, err := rr.NewRedisRepository(redisURL)
		if err != nil {
			log.Fatal(err)
		}
		return repo
	case "mongo":
		mongoURL := goDotEnvVariable("MONGO_URL")
		mongodb := goDotEnvVariable("MONGO_DB")
		mongoTimeout, _ := strconv.Atoi(goDotEnvVariable("MONGO_TIMEOUT"))
		repo, err := mr.NewMongoRepository(mongoURL, mongodb, mongoTimeout)
		if err != nil {
			log.Fatal(err)
		}
		return repo
	}
	return nil
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

