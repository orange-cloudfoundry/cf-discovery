package main

import (
	"fmt"
	"github.com/cloudfoundry-community/gautocloud"
	"github.com/cloudfoundry-community/gautocloud/connectors/generic"
	"github.com/gorilla/mux"
	"github.com/vulcand/oxy/buffer"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"log"
	"net/http"
	"net/url"
	"os"
)

type GorouterRoutes map[string]interface{}

type IndexerConfig struct {
	GorouterPassword string   `cloud:"gorouter_password"`
	GorouterUsername string   `cloud:"gorouter_username"`
	GorouterEndpoint string   `cloud:"gorouter_endpoint,default=/routes"`
	GorouterPort     int      `cloud:"gorouter_port,default=8080"`
	GorouterProtocol string   `cloud:"gorouter_protocol,default=http"`
	GorouterBackends []string `cloud:"gorouter_backends"`
	FilteredDomains  []string `cloud:"filtered_domains"`
	Debug            bool     `cloud:"debug"`
}

func init() {
	gautocloud.RegisterConnector(generic.NewConfigGenericConnector(IndexerConfig{}))
}

func main() {
	var config IndexerConfig
	err := gautocloud.Inject(&config)
	exitOnError(err)
	converter := NewConverter(http.DefaultTransport, config.FilteredDomains, config.Debug)

	fwd, err := forward.New(forward.RoundTripper(converter))
	exitOnError(err)
	lb, err := roundrobin.New(fwd)
	exitOnError(err)
	bufferHandler, _ := buffer.New(lb, buffer.Retry(`IsNetworkError() && Attempts() < 2`))
	for _, backend := range config.GorouterBackends {
		lbUrl, _ := url.Parse(fmt.Sprintf(
			"%s://%s:%d%s",
			config.GorouterProtocol,
			backend,
			config.GorouterPort,
			config.GorouterEndpoint,
		))
		lb.UpsertServer(lbUrl)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := mux.NewRouter()
	router.Handle("/routes", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.SetBasicAuth(config.GorouterUsername, config.GorouterPassword)
		bufferHandler.ServeHTTP(w, req)
	}))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	s := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	s.ListenAndServe()
}
func exitOnError(err error) {
	if err == nil {
		return
	}
	log.Fatal(err.Error())
}
