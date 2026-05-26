package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"gitops-values-translator-go/translator"
)

//go:embed static/index.html
var staticFiles embed.FS

type TranslationRequest struct {
	YAML               string `json:"yaml"`
	Cluster            string `json:"cluster"`
	EnvOverride        string `json:"envOverride"`
	TeamOverride       string `json:"teamOverride"`
	UseCommonConfigmap bool   `json:"useCommonConfigmap"`
	JavaOpts           string `json:"javaOpts"`
}

type TranslationResponse struct {
	Translated string `json:"translated,omitempty"`
	TargetPath string `json:"targetPath,omitempty"`
	Error      string `json:"error,omitempty"`
}

func main() {
	// Parse HTML template from embedded files
	tmpl, err := template.ParseFS(staticFiles, "static/index.html")
	if err != nil {
		log.Fatalf("Fatal error parsing embedded static files: %v", err)
	}

	// Serve Frontend
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		
		// Dynamically inject settings from environment variables into frontend
		config := map[string]string{
			"DefaultCluster":  getEnv("DEFAULT_CLUSTER", "on-premise"),
			"JavaToolOptions": getEnv("JAVA_TOOL_OPTIONS", "-Xms256m -Xmx768m -XX:+UseG1GC"),
		}
		if err := tmpl.Execute(w, config); err != nil {
			log.Printf("Error serving template: %v", err)
		}
	})

	// Translation REST API Endpoint
	http.HandleFunc("/api/translate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(TranslationResponse{Error: "Method not allowed"})
			return
		}

		var req TranslationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(TranslationResponse{Error: "Invalid JSON request payload"})
			return
		}

		translated, targetPath, err := translator.TranslateValues(
			req.YAML,
			req.Cluster,
			req.EnvOverride,
			req.TeamOverride,
			req.JavaOpts,
			req.UseCommonConfigmap,
		)
		if err != nil {
			json.NewEncoder(w).Encode(TranslationResponse{Error: err.Error()})
			return
		}

		json.NewEncoder(w).Encode(TranslationResponse{
			Translated: translated,
			TargetPath: targetPath,
		})
	})

	portStr := getEnv("PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 8080
	}

	fmt.Printf("\n")
	fmt.Printf("  🔮  \x1b[36m\x1b[1mGitOps Values Translator (Go version)\x1b[0m\n")
	fmt.Printf("  🚀  Server is running at \x1b[32m\x1b[1mhttp://localhost:%d\x1b[0m\n", port)
	fmt.Printf("  ⚡  Serving premium Vanilla HTML5/CSS/JS frontend\n")
	fmt.Printf("  📂  Embedding all assets natively in a compiled binary\n\n")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Fatal error starting web server: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
