package main

import (
	"flag"
	"github.com/bernardo-bruning/ollama-copilot/internal"
	"github.com/bernardo-bruning/ollama-copilot/internal/handlers"
	"github.com/bernardo-bruning/ollama-copilot/internal/middleware"
	"log"
	"net/http"
	"text/template"

	"github.com/ollama/ollama/api"
)

var (
	port        = flag.String("port", ":11437", "Port to listen on")
	portSSL     = flag.String("port-ssl", ":11436", "Port to listen on")
	proxyPort   = flag.String("proxy-port", ":11435", "Proxy port to listen on")
	cert        = flag.String("cert", "/etc/ollama-copilot/server.crt", "Certificate file path *.crt")
	key         = flag.String("key", "/etc/ollama-copilot/server.key", "Key file path *.key")
	model       = flag.String("model", "codellama:code", "LLM model to use")
	numPredict  = flag.Int("num-predict", 50, "Number of predictions to return")
	templateStr = flag.String("template", "<PRE> {{.Prefix}} <SUF> {{.Suffix}} <MID>", "Fill-in-middle template to apply in prompt")
)

// main is the entrypoint for the program.
func main() {
	flag.Parse()
	api, err := api.ClientFromEnvironment()

	if err != nil {
		log.Fatalf("error initialize api: %s", err.Error())
		return
	}

	templ, err := template.New("prompt").Parse(*templateStr)
	if err != nil {
		log.Fatalf("error parsing template: %s", err.Error())
		return
	}

	mux := http.NewServeMux()

	mux.Handle("/health", handlers.NewHealthHandler())
	mux.Handle("/copilot_internal/v2/token", handlers.NewTokenHandler())
	mux.Handle("/v1/engines/copilot-codex/completions", handlers.NewCompletionHandler(api, *model, templ, *numPredict))

	go internal.Proxy(*proxyPort, *port)

	go func() {
		err = http.ListenAndServeTLS(*portSSL, *cert, *key, middleware.LogMiddleware(mux))
		if err != nil {
			log.Fatalf("error listening: %s", err.Error())
		}
	}()

	err = http.ListenAndServe(*port, middleware.LogMiddleware(mux))
	if err != nil {
		log.Fatalf("error listening: %s", err.Error())
	}
}
