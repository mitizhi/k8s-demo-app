package main

import (
	"net/http"
	"strings"
	log "github.com/sirupsen/logrus"
	"k8s-demo-app/internal/handlers"
	"k8s-demo-app/internal/config"
)

const listening_host = "0.0.0.0"

func main() {

	log.Infof("\"" + config.GetAppName() + "\" version " + config.GetAppVersion() + " starting...")
	port := config.GetEnvDefault("PORT", "8080")
	/*
	  We assume prefix is of the form:

	   [ <directory-name> [ "/" <directory-name> ]* ]

	  So it consists of zero or more <directory-name> components
	  separated by slash "/". Specifically we do not expect slashes
	  at the beginning nor at the end of the prefix definition.
	   Note that in msys2: if given PREFIX="/" we get prefix === "/C:/msys65/" !!!

	  Just in case, we trim any trailing slashes...
	*/
	prefix := "/" + strings.TrimSuffix(config.GetEnvDefault("PREFIX", ""), "/") + "/"
	if prefix == "//" {
		prefix = "/"
	}

	http.HandleFunc("/", handlers.MakeHandler(prefix))
	log.Infof("Listening on %s:%s with prefix \"%s\" (Base URL: http://*:%s%s)",
		listening_host, port, prefix, port, prefix)
	if err := http.ListenAndServe(listening_host + ":" + port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
