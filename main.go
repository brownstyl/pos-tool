package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Initialize Datastore Framework
	if err := InitDB(); err != nil {
		fmt.Printf("⛔ Application startup crashed: %v\n", err)
		return
	}
	defer DB.Close()

	// Application Routing Definitions
	http.HandleFunc("/api/transfers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			LogTransactionHandler(w, r)
		case http.MethodGet:
			FetchAllTransactionsHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/escalate", CompileEscalationHandler)

	// Serve the interactive web interface dashboard
	http.HandleFunc("/", ServeDashboardHandler)

	fmt.Println("🚀 Enterprise Dispute Portal actively running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Network server initialization error: %v\n", err)
	}
}
