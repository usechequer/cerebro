package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("There was a problem loading the environment variables")
	}

	router := mux.NewRouter()

	authSubRouter := router.PathPrefix("/auth").Subrouter().MatcherFunc(func(request *http.Request, requestMatcher *mux.RouteMatch) bool {
		match, _ := regexp.MatchString("/auth.*", request.URL.Path)
		return match
	})
	authSubRouter.HandlerFunc(proxy(os.Getenv("CARBON_API_URL")))

	userSubRouter := router.PathPrefix("/users").Subrouter().MatcherFunc(func(request *http.Request, requestMatcher *mux.RouteMatch) bool {
		match, _ := regexp.MatchString("/users.*", request.URL.Path)
		return match
	})
	userSubRouter.HandlerFunc(proxy(os.Getenv("CARBON_API_URL")))

	projectSubRouter := router.PathPrefix("/projects").Subrouter().MatcherFunc(func(request *http.Request, requestMatcher *mux.RouteMatch) bool {
		match, _ := regexp.MatchString("/projects.*", request.URL.Path)
		return match
	})
	projectSubRouter.HandlerFunc(proxy(os.Getenv("NITRO_API_URL")))

	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	fmt.Printf("Gateway is running on " + port)
	log.Fatal(http.ListenAndServe(port, router))
}

func proxy(target string) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, req *http.Request) {
		targetURL := target + req.URL.Path

		request, err := http.NewRequest(req.Method, targetURL, req.Body)

		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusBadGateway)
			return
		}

		request.Header = req.Header

		client := new(http.Client)
		response, err := client.Do(request)

		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusBadGateway)
			return
		}

		defer response.Body.Close()

		for key, values := range response.Header {
			for _, value := range values {
				responseWriter.Header().Add(key, value)
			}
		}

		responseWriter.WriteHeader(response.StatusCode)

		_, err = io.Copy(responseWriter, response.Body)

		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusBadGateway)
			return
		}
	}
}
