package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"financial-dashboard/internal/handlers"
	"financial-dashboard/internal/lua"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	luaEng, err := lua.NewEngine()
	if err != nil {
		log.Fatalf("lua engine: %v", err)
	}

	store := handlers.NewStore()
	h := handlers.New(store, luaEng)

	mux := http.NewServeMux()

	// Main page dispatcher (all tabs via ?tab=)
	mux.HandleFunc("/", h.Index)

	// Expense routes
	mux.HandleFunc("/expenses/new", h.ExpenseForm)
	mux.HandleFunc("/expenses", h.ExpensePost)
	mux.HandleFunc("/expenses/", h.ExpenseRouter)

	// Goal routes
	mux.HandleFunc("/goals/new", h.GoalForm)
	mux.HandleFunc("/goals", h.GoalPost)
	mux.HandleFunc("/goals/", h.GoalRouter)

	// Budget routes
	mux.HandleFunc("/budget/edit", h.BudgetForm)
	mux.HandleFunc("/budget", h.BudgetPost)

	// Deposit routes
	mux.HandleFunc("/deposits/new", h.DepositForm)
	mux.HandleFunc("/deposits", h.DepositPost)

	// Reset
	mux.HandleFunc("/reset/confirm", h.ResetConfirm)
	mux.HandleFunc("/reset", h.Reset)

	// Static assets (CSS, fonts, images)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      logMW(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Printf("FinançasLua — http://localhost:%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("[%s] %s %s %v\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path, time.Since(start))
	})
}
