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
	if port == "" { port = "8080" }

	luaEng := lua.NewEngine()
	store := handlers.NewStore()
	h := handlers.New(store, luaEng)

	mux := http.NewServeMux()

	// Dashboard
	mux.HandleFunc("/api/dashboard",      cors(h.Dashboard))
	// Expenses collection + individual
	mux.HandleFunc("/api/expenses",       cors(h.Expenses))
	mux.HandleFunc("/api/expenses/",      cors(h.ExpenseByID))
	// Goals collection + individual
	mux.HandleFunc("/api/goals",          cors(h.Goals))
	mux.HandleFunc("/api/goals/",         cors(h.GoalByID))
	// Budget / income
	mux.HandleFunc("/api/budget",         cors(h.Budget))
	// Calendar & Lua
	mux.HandleFunc("/api/calendar",       cors(h.Calendar))
	mux.HandleFunc("/api/lua-scripts",    cors(h.LuaScripts))

	// Static
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	srv := &http.Server{Addr: ":" + port, Handler: logMW(mux), ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second}
	log.Printf("🚀 FinançasLua — http://localhost:%s", port)
	if err := srv.ListenAndServe(); err != nil { log.Fatalf("server: %v", err) }
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusNoContent); return }
		next(w, r)
	}
}

func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("[%s] %s %s %v\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path, time.Since(start))
	})
}
