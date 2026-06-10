package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"financial-dashboard/internal/lua"
	"financial-dashboard/internal/models"
)

// Handler holds all HTTP handler dependencies.
type Handler struct {
	store  *Store
	luaEng *lua.Engine
}

// New creates a new Handler.
func New(store *Store, eng *lua.Engine) *Handler {
	return &Handler{store: store, luaEng: eng}
}

func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// Dashboard returns the full dashboard summary.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	currentExpenses := h.store.GetCurrentMonthExpenses()

	// Sum by category
	byCategory := make(map[string]float64)
	totalExpenses := 0.0
	for _, e := range currentExpenses {
		byCategory[string(e.Category)] += e.Amount
		totalExpenses += e.Amount
	}

	income := h.store.Budget.Income

	// Budget usage %
	budgetUsage := make(map[string]float64)
	for cat, budget := range h.store.Budget.Budgets {
		if budget > 0 {
			budgetUsage[string(cat)] = (byCategory[string(cat)] / budget) * 100
		}
	}

	// Monthly trend (last 6 months)
	monthlyTotals := h.store.GetMonthlyTotals()
	var forecasts []models.Forecast
	var historyAmounts []float64

	for i := 5; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		key := t.Format("2006-01")
		actual := monthlyTotals[key]
		historyAmounts = append(historyAmounts, actual)
		forecasts = append(forecasts, models.Forecast{
			Month:     t.Format("Jan/06"),
			Actual:    actual,
			Predicted: 0,
			Income:    income,
		})
	}

	// Lua forecast for next month
	predicted := h.luaEng.RunForecast(historyAmounts)
	nextMonth := now.AddDate(0, 1, 0)
	forecasts = append(forecasts, models.Forecast{
		Month:     nextMonth.Format("Jan/06") + " (prev.)",
		Predicted: predicted,
		Actual:    0,
		Income:    income,
	})

	// Recent expenses (last 10)
	sort.Slice(h.store.Expenses, func(i, j int) bool {
		return h.store.Expenses[i].Date.After(h.store.Expenses[j].Date)
	})
	recent := h.store.Expenses
	if len(recent) > 10 {
		recent = recent[:10]
	}

	summary := models.DashboardSummary{
		TotalExpenses:      totalExpenses,
		TotalIncome:        income,
		Balance:            income - totalExpenses,
		ExpensesByCategory: byCategory,
		MonthlyTrend:       forecasts,
		Goals:              h.store.Goals,
		RecentExpenses:     recent,
		BudgetUsage:        budgetUsage,
	}

	summary.LuaInsights = h.luaEng.RunInsights(summary)

	jsonResponse(w, http.StatusOK, summary)
}

// Expenses returns all current month expenses.
func (h *Handler) Expenses(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var e models.Expense
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		created := h.store.AddExpense(e)
		jsonResponse(w, http.StatusCreated, created)
		return
	}
	jsonResponse(w, http.StatusOK, h.store.GetCurrentMonthExpenses())
}

// Goals returns all financial goals.
func (h *Handler) Goals(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, h.store.Goals)
}

// Calendar returns calendar events for the current month.
func (h *Handler) Calendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	var events []models.CalendarEvent

	for _, e := range h.store.Expenses {
		if e.Date.Month() == now.Month() && e.Date.Year() == now.Year() {
			events = append(events, models.CalendarEvent{
				ID:       e.ID,
				Title:    e.Description,
				Amount:   e.Amount,
				Date:     e.Date,
				Type:     "expense",
				Category: e.Category,
			})
		}
	}

	// Add goal deadlines
	for _, g := range h.store.Goals {
		if g.Deadline.Month() == now.Month() && g.Deadline.Year() == now.Year() {
			events = append(events, models.CalendarEvent{
				ID:    g.ID + 1000,
				Title: "Meta: " + g.Name,
				Date:  g.Deadline,
				Type:  "goal",
			})
		}
	}

	jsonResponse(w, http.StatusOK, events)
}

// LuaScripts returns the Lua script sources for transparency.
func (h *Handler) LuaScripts(w http.ResponseWriter, r *http.Request) {
	scripts := map[string]string{
		"forecast": h.luaEng.ScriptSource("forecast"),
		"insights": h.luaEng.ScriptSource("insights"),
		"budget":   h.luaEng.ScriptSource("budget_check"),
	}
	jsonResponse(w, http.StatusOK, scripts)
}
