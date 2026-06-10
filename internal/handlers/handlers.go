package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"financial-dashboard/internal/lua"
	"financial-dashboard/internal/models"
)

type Handler struct {
	store  *Store
	luaEng *lua.Engine
}

func New(store *Store, eng *lua.Engine) *Handler {
	return &Handler{store: store, luaEng: eng}
}

func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errResponse(w http.ResponseWriter, status int, msg string) {
	jsonResponse(w, status, map[string]string{"error": msg})
}

func idFromPath(r *http.Request) (int, bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		return 0, false
	}
	id, err := strconv.Atoi(parts[len(parts)-1])
	return id, err == nil
}

// ── DASHBOARD ─────────────────────────────────

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	currentExpenses := h.store.GetCurrentMonthExpenses()
	budget := h.store.GetBudget()

	byCategory := make(map[string]float64)
	totalExpenses := 0.0
	for _, e := range currentExpenses {
		byCategory[string(e.Category)] += e.Amount
		totalExpenses += e.Amount
	}

	income := budget.Income
	budgetUsage := make(map[string]float64)
	for cat, lim := range budget.Budgets {
		if lim > 0 {
			budgetUsage[string(cat)] = (byCategory[string(cat)] / lim) * 100
		}
	}

	monthlyTotals := h.store.GetMonthlyTotals()
	var forecasts []models.Forecast
	var historyAmounts []float64
	for i := 5; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		key := t.Format("2006-01")
		actual := monthlyTotals[key]
		historyAmounts = append(historyAmounts, actual)
		forecasts = append(forecasts, models.Forecast{Month: t.Format("Jan/06"), Actual: actual, Income: income})
	}
	predicted := h.luaEng.RunForecast(historyAmounts)
	forecasts = append(forecasts, models.Forecast{
		Month: now.AddDate(0, 1, 0).Format("Jan/06") + " ▲", Predicted: predicted, Income: income,
	})

	allExp := h.store.GetAllExpenses()
	sort.Slice(allExp, func(i, j int) bool { return allExp[i].Date.After(allExp[j].Date) })
	recent := allExp
	if len(recent) > 10 {
		recent = recent[:10]
	}

	goals := h.store.GetGoals()
	summary := models.DashboardSummary{
		TotalExpenses: totalExpenses, TotalIncome: income,
		Balance: income - totalExpenses, ExpensesByCategory: byCategory,
		MonthlyTrend: forecasts, Goals: goals, RecentExpenses: recent, BudgetUsage: budgetUsage,
	}
	summary.LuaInsights = h.luaEng.RunInsights(summary)
	jsonResponse(w, http.StatusOK, summary)
}

// ── EXPENSES ─────────────────────────────────

func (h *Handler) Expenses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		all := h.store.GetAllExpenses()
		sort.Slice(all, func(i, j int) bool { return all[i].Date.After(all[j].Date) })
		jsonResponse(w, http.StatusOK, all)

	case http.MethodPost:
		var e models.Expense
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			errResponse(w, 400, err.Error())
			return
		}
		jsonResponse(w, http.StatusCreated, h.store.AddExpense(e))
	}
}

func (h *Handler) ExpenseByID(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromPath(r)
	if !ok {
		errResponse(w, 400, "id inválido")
		return
	}

	switch r.Method {
	case http.MethodPut:
		var e models.Expense
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			errResponse(w, 400, err.Error())
			return
		}
		e.ID = id
		updated, found := h.store.UpdateExpense(e)
		if !found {
			errResponse(w, 404, "gasto não encontrado")
			return
		}
		jsonResponse(w, http.StatusOK, updated)

	case http.MethodDelete:
		if !h.store.DeleteExpense(id) {
			errResponse(w, 404, "gasto não encontrado")
			return
		}
		jsonResponse(w, http.StatusOK, map[string]bool{"deleted": true})
	}
}

// ── GOALS ─────────────────────────────────────

func (h *Handler) Goals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonResponse(w, http.StatusOK, h.store.GetGoals())
	case http.MethodPost:
		var g models.Goal
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			errResponse(w, 400, err.Error())
			return
		}
		jsonResponse(w, http.StatusCreated, h.store.AddGoal(g))
	}
}

func (h *Handler) GoalByID(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromPath(r)
	if !ok {
		errResponse(w, 400, "id inválido")
		return
	}

	switch r.Method {
	case http.MethodPut:
		var g models.Goal
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			errResponse(w, 400, err.Error())
			return
		}
		g.ID = id
		updated, found := h.store.UpdateGoal(g)
		if !found {
			errResponse(w, 404, "meta não encontrada")
			return
		}
		jsonResponse(w, http.StatusOK, updated)

	case http.MethodDelete:
		if !h.store.DeleteGoal(id) {
			errResponse(w, 404, "meta não encontrada")
			return
		}
		jsonResponse(w, http.StatusOK, map[string]bool{"deleted": true})
	}
}

// ── BUDGET ────────────────────────────────────

func (h *Handler) Budget(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonResponse(w, http.StatusOK, h.store.GetBudget())
	case http.MethodPut:
		var b models.MonthlyBudget
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			errResponse(w, 400, err.Error())
			return
		}
		h.store.UpdateBudget(b)
		jsonResponse(w, http.StatusOK, b)
	}
}

// ── CALENDAR ──────────────────────────────────

func (h *Handler) Calendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	var events []models.CalendarEvent
	for _, e := range h.store.GetCurrentMonthExpenses() {
		events = append(events, models.CalendarEvent{
			ID: e.ID, Title: e.Description, Amount: e.Amount,
			Date: e.Date, Type: "expense", Category: e.Category,
		})
	}
	for _, g := range h.store.GetGoals() {
		if g.Deadline.Month() == now.Month() && g.Deadline.Year() == now.Year() {
			events = append(events, models.CalendarEvent{
				ID: g.ID + 10000, Title: "Meta: " + g.Name, Date: g.Deadline, Type: "goal",
			})
		}
	}
	jsonResponse(w, http.StatusOK, events)
}

// ── LUA SCRIPTS ───────────────────────────────

func (h *Handler) LuaScripts(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]string{
		"forecast": h.luaEng.ScriptSource("forecast"),
		"insights": h.luaEng.ScriptSource("insights"),
		"budget":   h.luaEng.ScriptSource("budget_check"),
	})
}

func (h *Handler) GoalDeposit(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromPath(r)
	if !ok {
		errResponse(w, 400, "id inválido")
		return
	}
	var body struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errResponse(w, 400, err.Error())
		return
	}
	h.store.mu.Lock()
	for i, g := range h.store.Goals {
		if g.ID == id {
			h.store.Goals[i].CurrentAmount += body.Amount
			jsonResponse(w, http.StatusOK, h.store.Goals[i])
			h.store.mu.Unlock()
			return
		}
	}
	h.store.mu.Unlock()
	errResponse(w, 404, "meta não encontrada")
}
