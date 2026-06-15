package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"financial-dashboard/internal/charts"
	"financial-dashboard/internal/lua"
	"financial-dashboard/internal/models"
	"financial-dashboard/internal/pages"
)

type Handler struct {
	store   *Store
	luaEng  *lua.Engine
	funcMap template.FuncMap
}

func New(store *Store, eng *lua.Engine) *Handler {
	return &Handler{store: store, luaEng: eng, funcMap: pages.FuncMap()}
}

// render executes layout.html + the named page template.
func (h *Handler) render(w http.ResponseWriter, page string, data any) {
	t, err := template.New("layout.html").Funcs(h.funcMap).ParseFiles(
		"web/templates/layout.html",
		"web/templates/"+page+".html",
	)
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) basePage(r *http.Request, tab, title string) pages.BasePage {
	f := r.URL.Query().Get("flash")
	return pages.BasePage{
		Tab:         tab,
		PageTitle:   title,
		PageDate:    time.Now().Format("Monday, 02 de January de 2006"),
		LuaInsights: h.computeInsights(),
		Flash:       pages.FlashMsg(f),
		FlashType:   pages.FlashType(f),
	}
}

func (h *Handler) computeInsights() string {
	currentExp := h.store.GetCurrentMonthExpenses()
	budget := h.store.GetBudget()
	totalExp := 0.0
	for _, e := range currentExp {
		totalExp += e.Amount
	}
	depositTotal := 0.0
	for _, d := range h.store.GetCurrentMonthDeposits() {
		depositTotal += d.Amount
	}
	return h.luaEng.RunInsights(models.DashboardSummary{
		TotalIncome:   budget.Income + depositTotal,
		TotalExpenses: totalExp,
		Goals:         h.store.GetGoals(),
	})
}

func segmentFromPath(path string) (id int, action string, ok bool) {
	// path like /expenses/42/edit or /expenses/42/delete
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, "", false
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, "", false
	}
	if len(parts) >= 3 {
		return id, parts[2], true
	}
	return id, "", true
}

// ── MAIN PAGE DISPATCHER ──────────────────────────────────────────────────

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "overview"
	}
	switch tab {
	case "overview":
		h.renderOverview(w, r)
	case "expenses":
		h.renderExpenses(w, r)
	case "goals":
		h.renderGoals(w, r)
	case "calendar":
		h.renderCalendar(w, r)
	case "forecast":
		h.renderForecast(w, r)
	case "lua":
		h.renderLua(w, r)
	default:
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// ── OVERVIEW ─────────────────────────────────────────────────────────────

func (h *Handler) renderOverview(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	currentExp := h.store.GetCurrentMonthExpenses()
	budget := h.store.GetBudget()

	byCategory := make(map[string]float64)
	totalExp := 0.0
	for _, e := range currentExp {
		byCategory[string(e.Category)] += e.Amount
		totalExp += e.Amount
	}
	depositTotal := 0.0
	for _, d := range h.store.GetCurrentMonthDeposits() {
		depositTotal += d.Amount
	}
	income := budget.Income + depositTotal
	balance := income - totalExp

	expPct := 0.0
	if income > 0 {
		expPct = (totalExp / income) * 100
	}
	savPct := 0.0
	if income > 0 {
		savPct = ((income - totalExp) / income) * 100
	}

	budgetItems := make([]pages.BudgetItem, 0, len(budget.Budgets))
	for cat, lim := range budget.Budgets {
		spent := byCategory[string(cat)]
		pct := 0.0
		if lim > 0 {
			pct = (spent / lim) * 100
		}
		budgetItems = append(budgetItems, pages.BudgetItem{
			Category: string(cat),
			Icon:     pages.CatIcon(string(cat)),
			Spent:    spent,
			Limit:    lim,
			Pct:      pct,
			Class:    pages.BudgetClass(pct),
		})
	}
	sort.Slice(budgetItems, func(i, j int) bool {
		return budgetItems[i].Category < budgetItems[j].Category
	})

	monthlyTotals := h.store.GetMonthlyTotals()
	var forecasts []models.Forecast
	var historyAmts []float64
	for i := 5; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		key := t.Format("2006-01")
		actual := monthlyTotals[key]
		historyAmts = append(historyAmts, actual)
		forecasts = append(forecasts, models.Forecast{
			Month: t.Format("Jan/06"), Actual: actual, Income: income,
		})
	}
	predicted := h.luaEng.RunForecast(historyAmts)
	forecasts = append(forecasts, models.Forecast{
		Month: now.AddDate(0, 1, 0).Format("Jan/06") + " ▲", Predicted: predicted, Income: income,
	})

	allExp := h.store.GetAllExpenses()
	sort.Slice(allExp, func(i, j int) bool { return allExp[i].Date.After(allExp[j].Date) })
	if len(allExp) > 10 {
		allExp = allExp[:10]
	}

	goals := h.store.GetGoals()
	goalsComplete := 0
	for _, g := range goals {
		if g.TargetAmount > 0 && g.CurrentAmount >= g.TargetAmount {
			goalsComplete++
		}
	}

	data := pages.OverviewPage{
		BasePage:       h.basePage(r, "overview", "Visão Geral"),
		Income:         income,
		DepositTotal:   depositTotal,
		TotalExpenses:  totalExp,
		Balance:        balance,
		ExpensePct:     expPct,
		SavingsPct:     savPct,
		GoalsCount:     len(goals),
		GoalsComplete:  goalsComplete,
		DonutSVG:       charts.RenderDonut(byCategory),
		BarSVG:         charts.RenderBar(forecasts),
		BudgetItems:    budgetItems,
		RecentExpenses: allExp,
	}
	h.render(w, "overview", data)
}

// ── EXPENSES ─────────────────────────────────────────────────────────────

func (h *Handler) renderExpenses(w http.ResponseWriter, r *http.Request) {
	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	cat := r.URL.Query().Get("cat")

	all := h.store.GetAllExpenses()
	sort.Slice(all, func(i, j int) bool { return all[i].Date.After(all[j].Date) })

	var filtered []models.Expense
	for _, e := range all {
		if cat != "" && string(e.Category) != cat {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(e.Description), search) {
			continue
		}
		filtered = append(filtered, e)
	}

	h.render(w, "expenses", pages.ExpensesPage{
		BasePage:   h.basePage(r, "expenses", "Gastos"),
		Search:     r.URL.Query().Get("search"),
		Category:   cat,
		Expenses:   filtered,
		Categories: pages.AllCategories,
	})
}

// ── GOALS ────────────────────────────────────────────────────────────────

func (h *Handler) renderGoals(w http.ResponseWriter, r *http.Request) {
	goals := h.store.GetGoals()
	var vms []pages.GoalVM
	for _, g := range goals {
		pct := 0.0
		if g.TargetAmount > 0 {
			pct = (g.CurrentAmount / g.TargetAmount) * 100
		}
		vms = append(vms, pages.GoalVM{
			Goal:        g,
			Pct:         pct,
			PctClamped:  pages.ClampPct(pct),
			DeadlineFmt: g.Deadline.Format("Jan/2006"),
		})
	}
	h.render(w, "goals", pages.GoalsPage{
		BasePage: h.basePage(r, "goals", "Metas"),
		Goals:    vms,
	})
}

// ── CALENDAR ─────────────────────────────────────────────────────────────

func (h *Handler) renderCalendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	year, month := now.Year(), int(now.Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}
	selectedDay := 0
	if d := r.URL.Query().Get("day"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			selectedDay = parsed
		}
	}

	expenses := h.store.GetExpensesByMonth(year, month)
	goals := h.store.GetGoals()

	eventsByDay := make(map[int][]models.CalendarEvent)
	var allEvents []models.CalendarEvent
	for _, e := range expenses {
		day := e.Date.Day()
		ev := models.CalendarEvent{
			ID: e.ID, Title: e.Description, Amount: e.Amount,
			Date: e.Date, Type: "expense", Category: e.Category,
		}
		eventsByDay[day] = append(eventsByDay[day], ev)
		allEvents = append(allEvents, ev)
	}
	for _, g := range goals {
		if g.Deadline.Year() == year && int(g.Deadline.Month()) == month {
			day := g.Deadline.Day()
			ev := models.CalendarEvent{
				ID: g.ID + 10000, Title: "Meta: " + g.Name,
				Date: g.Deadline, Type: "goal",
			}
			eventsByDay[day] = append(eventsByDay[day], ev)
			allEvents = append(allEvents, ev)
		}
	}
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Date.Before(allEvents[j].Date)
	})

	// Events to show in the list below calendar
	var shownEvents []pages.CalEventVM
	if selectedDay > 0 {
		for _, ev := range eventsByDay[selectedDay] {
			shownEvents = append(shownEvents, pages.CalEventVM{
				CalendarEvent: ev,
				DateFmt:       ev.Date.Format("02/01"),
			})
		}
	} else {
		for _, ev := range allEvents {
			shownEvents = append(shownEvents, pages.CalEventVM{
				CalendarEvent: ev,
				DateFmt:       ev.Date.Format("02/01"),
			})
		}
	}

	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local).Weekday()
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local).Day()

	prevY, prevM := year, month-1
	if prevM < 1 {
		prevM = 12; prevY--
	}
	nextY, nextM := year, month+1
	if nextM > 12 {
		nextM = 1; nextY++
	}
	todayDay := 0
	if now.Year() == year && int(now.Month()) == month {
		todayDay = now.Day()
	}

	h.render(w, "calendar", pages.CalendarPage{
		BasePage:     h.basePage(r, "calendar", "Calendário"),
		Year:         year,
		Month:        month,
		MonthLabel:   pages.FmtMonth(year, month),
		FirstWeekday: int(firstDay),
		DaysInMonth:  daysInMonth,
		Today:        todayDay,
		SelectedDay:  selectedDay,
		EventsByDay:  eventsByDay,
		AllEvents:    allEvents,
		ShownEvents:  shownEvents,
		PrevYear:     prevY,
		PrevMonth:    prevM,
		NextYear:     nextY,
		NextMonth:    nextM,
	})
}

// ── FORECAST ─────────────────────────────────────────────────────────────

func (h *Handler) renderForecast(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	budget := h.store.GetBudget()
	depositTotal := 0.0
	for _, d := range h.store.GetCurrentMonthDeposits() {
		depositTotal += d.Amount
	}
	income := budget.Income + depositTotal

	monthlyTotals := h.store.GetMonthlyTotals()
	var forecasts []models.Forecast
	var historyAmts []float64
	for i := 5; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		key := t.Format("2006-01")
		actual := monthlyTotals[key]
		historyAmts = append(historyAmts, actual)
		forecasts = append(forecasts, models.Forecast{
			Month: t.Format("Jan/06"), Actual: actual, Income: income,
		})
	}
	predicted := h.luaEng.RunForecast(historyAmts)
	forecasts = append(forecasts, models.Forecast{
		Month: now.AddDate(0, 1, 0).Format("Jan/06") + " ▲", Predicted: predicted, Income: income,
	})

	h.render(w, "forecast", pages.ForecastPage{
		BasePage:    h.basePage(r, "forecast", "Previsões"),
		ForecastSVG: charts.RenderForecastLine(forecasts),
		MonthCards:  forecasts,
	})
}

// ── LUA SCRIPTS ──────────────────────────────────────────────────────────

func (h *Handler) renderLua(w http.ResponseWriter, r *http.Request) {
	h.render(w, "lua", pages.LuaPage{
		BasePage:    h.basePage(r, "lua", "Scripts Lua"),
		ForecastSrc: h.luaEng.ScriptSource("forecast"),
		InsightsSrc: h.luaEng.ScriptSource("insights"),
	})
}

// ── EXPENSE FORM ─────────────────────────────────────────────────────────

func (h *Handler) ExpenseForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "expense_form", pages.ExpenseFormPage{
		BasePage:   h.basePage(r, "expenses", "Novo Gasto"),
		IsEdit:     false,
		Expense:    models.Expense{Date: time.Now()},
		Categories: pages.AllCategories,
	})
}

func (h *Handler) ExpenseRouter(w http.ResponseWriter, r *http.Request) {
	id, action, ok := segmentFromPath(r.URL.Path)
	if !ok {
		http.Redirect(w, r, "/?tab=expenses", http.StatusFound)
		return
	}

	switch {
	case r.Method == http.MethodGet && action == "edit":
		all := h.store.GetAllExpenses()
		for _, e := range all {
			if e.ID == id {
				h.render(w, "expense_form", pages.ExpenseFormPage{
					BasePage:   h.basePage(r, "expenses", "Editar Gasto"),
					IsEdit:     true,
					Expense:    e,
					Categories: pages.AllCategories,
				})
				return
			}
		}
		http.Redirect(w, r, "/?tab=expenses", http.StatusFound)

	case r.Method == http.MethodGet && action == "confirm-delete":
		all := h.store.GetAllExpenses()
		for _, e := range all {
			if e.ID == id {
				h.render(w, "confirm_delete", pages.ConfirmDeletePage{
					BasePage:  h.basePage(r, "expenses", "Confirmar Exclusão"),
					Name:      e.Description,
					BackURL:   "/?tab=expenses",
					DeleteURL: fmt.Sprintf("/expenses/%d/delete", id),
				})
				return
			}
		}
		http.Redirect(w, r, "/?tab=expenses", http.StatusFound)

	case r.Method == http.MethodPost && action == "delete":
		h.store.DeleteExpense(id)
		http.Redirect(w, r, "/?tab=expenses&flash=deleted", http.StatusSeeOther)

	case r.Method == http.MethodPost && action == "":
		h.handleExpenseUpdate(w, r, id)

	default:
		http.Redirect(w, r, "/?tab=expenses", http.StatusFound)
	}
}

func (h *Handler) ExpensePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/expenses/new", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form error", http.StatusBadRequest)
		return
	}
	desc := strings.TrimSpace(r.FormValue("description"))
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 || desc == "" {
		h.render(w, "expense_form", pages.ExpenseFormPage{
			BasePage:   h.basePage(r, "expenses", "Novo Gasto"),
			IsEdit:     false,
			Expense:    buildExpenseFromForm(r),
			Categories: pages.AllCategories,
			Error:      "Preencha descrição e valor válido.",
		})
		return
	}
	date, _ := time.Parse("2006-01-02", r.FormValue("date"))
	if date.IsZero() {
		date = time.Now()
	}
	h.store.AddExpense(models.Expense{
		Description: desc,
		Amount:      amount,
		Category:    models.Category(r.FormValue("category")),
		Date:        date,
		IsRecurring: r.FormValue("is_recurring") == "on",
	})
	http.Redirect(w, r, "/?tab=expenses&flash=added", http.StatusSeeOther)
}

func (h *Handler) handleExpenseUpdate(w http.ResponseWriter, r *http.Request, id int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form error", http.StatusBadRequest)
		return
	}
	desc := strings.TrimSpace(r.FormValue("description"))
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 || desc == "" {
		exp := buildExpenseFromForm(r)
		exp.ID = id
		h.render(w, "expense_form", pages.ExpenseFormPage{
			BasePage:   h.basePage(r, "expenses", "Editar Gasto"),
			IsEdit:     true,
			Expense:    exp,
			Categories: pages.AllCategories,
			Error:      "Preencha descrição e valor válido.",
		})
		return
	}
	date, _ := time.Parse("2006-01-02", r.FormValue("date"))
	if date.IsZero() {
		date = time.Now()
	}
	h.store.UpdateExpense(models.Expense{
		ID:          id,
		Description: desc,
		Amount:      amount,
		Category:    models.Category(r.FormValue("category")),
		Date:        date,
		IsRecurring: r.FormValue("is_recurring") == "on",
	})
	http.Redirect(w, r, "/?tab=expenses&flash=updated", http.StatusSeeOther)
}

func buildExpenseFromForm(r *http.Request) models.Expense {
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	date, _ := time.Parse("2006-01-02", r.FormValue("date"))
	if date.IsZero() {
		date = time.Now()
	}
	return models.Expense{
		Description: r.FormValue("description"),
		Amount:      amount,
		Category:    models.Category(r.FormValue("category")),
		Date:        date,
		IsRecurring: r.FormValue("is_recurring") == "on",
	}
}

// ── GOAL FORM ────────────────────────────────────────────────────────────

func (h *Handler) GoalForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "goal_form", pages.GoalFormPage{
		BasePage: h.basePage(r, "goals", "Nova Meta"),
		IsEdit:   false,
		Goal:     models.Goal{Color: pages.GoalColors[0], Icon: pages.GoalIcons[0], Deadline: time.Now().AddDate(1, 0, 0)},
		Icons:    pages.GoalIcons,
		Colors:   pages.GoalColors,
	})
}

func (h *Handler) GoalRouter(w http.ResponseWriter, r *http.Request) {
	id, action, ok := segmentFromPath(r.URL.Path)
	if !ok {
		http.Redirect(w, r, "/?tab=goals", http.StatusFound)
		return
	}

	switch {
	case r.Method == http.MethodGet && action == "edit":
		for _, g := range h.store.GetGoals() {
			if g.ID == id {
				h.render(w, "goal_form", pages.GoalFormPage{
					BasePage: h.basePage(r, "goals", "Editar Meta"),
					IsEdit:   true,
					Goal:     g,
					Icons:    pages.GoalIcons,
					Colors:   pages.GoalColors,
				})
				return
			}
		}
		http.Redirect(w, r, "/?tab=goals", http.StatusFound)

	case r.Method == http.MethodGet && action == "confirm-delete":
		for _, g := range h.store.GetGoals() {
			if g.ID == id {
				h.render(w, "confirm_delete", pages.ConfirmDeletePage{
					BasePage:  h.basePage(r, "goals", "Confirmar Exclusão"),
					Name:      g.Name,
					BackURL:   "/?tab=goals",
					DeleteURL: fmt.Sprintf("/goals/%d/delete", id),
				})
				return
			}
		}
		http.Redirect(w, r, "/?tab=goals", http.StatusFound)

	case r.Method == http.MethodGet && action == "deposit":
		for _, g := range h.store.GetGoals() {
			if g.ID == id {
				h.render(w, "goal_deposit_form", pages.GoalDepositFormPage{
					BasePage: h.basePage(r, "goals", "Depositar em Meta"),
					Goal:     g,
				})
				return
			}
		}
		http.Redirect(w, r, "/?tab=goals", http.StatusFound)

	case r.Method == http.MethodPost && action == "delete":
		h.store.DeleteGoal(id)
		http.Redirect(w, r, "/?tab=goals&flash=deleted", http.StatusSeeOther)

	case r.Method == http.MethodPost && action == "deposit":
		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, "/?tab=goals", http.StatusSeeOther)
			return
		}
		amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
		if err != nil || amount <= 0 {
			for _, g := range h.store.GetGoals() {
				if g.ID == id {
					h.render(w, "goal_deposit_form", pages.GoalDepositFormPage{
						BasePage: h.basePage(r, "goals", "Depositar em Meta"),
						Goal:     g,
						Error:    "Valor inválido.",
					})
					return
				}
			}
		}
		h.store.AddGoalDeposit(id, amount)
		http.Redirect(w, r, "/?tab=goals&flash=deposited", http.StatusSeeOther)

	case r.Method == http.MethodPost && action == "":
		h.handleGoalUpdate(w, r, id)

	default:
		http.Redirect(w, r, "/?tab=goals", http.StatusFound)
	}
}

func (h *Handler) GoalPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/goals/new", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form error", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	target, err := strconv.ParseFloat(r.FormValue("target_amount"), 64)
	if err != nil || target <= 0 || name == "" {
		g := buildGoalFromForm(r)
		h.render(w, "goal_form", pages.GoalFormPage{
			BasePage: h.basePage(r, "goals", "Nova Meta"),
			IsEdit:   false,
			Goal:     g,
			Icons:    pages.GoalIcons,
			Colors:   pages.GoalColors,
			Error:    "Preencha nome e valor alvo válido.",
		})
		return
	}
	h.store.AddGoal(buildGoalFromForm(r))
	http.Redirect(w, r, "/?tab=goals&flash=added", http.StatusSeeOther)
}

func (h *Handler) handleGoalUpdate(w http.ResponseWriter, r *http.Request, id int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form error", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	target, err := strconv.ParseFloat(r.FormValue("target_amount"), 64)
	if err != nil || target <= 0 || name == "" {
		g := buildGoalFromForm(r)
		g.ID = id
		h.render(w, "goal_form", pages.GoalFormPage{
			BasePage: h.basePage(r, "goals", "Editar Meta"),
			IsEdit:   true,
			Goal:     g,
			Icons:    pages.GoalIcons,
			Colors:   pages.GoalColors,
			Error:    "Preencha nome e valor alvo válido.",
		})
		return
	}
	g := buildGoalFromForm(r)
	g.ID = id
	h.store.UpdateGoal(g)
	http.Redirect(w, r, "/?tab=goals&flash=updated", http.StatusSeeOther)
}

func buildGoalFromForm(r *http.Request) models.Goal {
	target, _ := strconv.ParseFloat(r.FormValue("target_amount"), 64)
	current, _ := strconv.ParseFloat(r.FormValue("current_amount"), 64)
	deadline, _ := time.Parse("2006-01-02", r.FormValue("deadline"))
	if deadline.IsZero() {
		deadline = time.Now().AddDate(1, 0, 0)
	}
	color := r.FormValue("color")
	if color == "" {
		color = pages.GoalColors[0]
	}
	icon := r.FormValue("icon")
	if icon == "" {
		icon = pages.GoalIcons[0]
	}
	return models.Goal{
		Name:          strings.TrimSpace(r.FormValue("name")),
		TargetAmount:  target,
		CurrentAmount: current,
		Deadline:      deadline,
		Color:         color,
		Icon:          icon,
	}
}

// ── BUDGET ───────────────────────────────────────────────────────────────

func (h *Handler) BudgetForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "budget_form", pages.BudgetFormPage{
		BasePage:   h.basePage(r, "overview", "Orçamento"),
		Budget:     h.store.GetBudget(),
		Categories: pages.AllCategories,
	})
}

func (h *Handler) BudgetPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/budget/edit", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form error", http.StatusBadRequest)
		return
	}
	income, _ := strconv.ParseFloat(r.FormValue("income"), 64)
	now := time.Now()
	b := models.MonthlyBudget{
		Month:   int(now.Month()),
		Year:    now.Year(),
		Income:  income,
		Budgets: make(map[models.Category]float64),
	}
	for _, cat := range pages.AllCategories {
		v, _ := strconv.ParseFloat(r.FormValue("b_"+cat), 64)
		b.Budgets[models.Category(cat)] = v
	}
	h.store.UpdateBudget(b)
	http.Redirect(w, r, "/?tab=overview&flash=saved", http.StatusSeeOther)
}

// ── DEPOSIT ──────────────────────────────────────────────────────────────

func (h *Handler) DepositForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "deposit_form", pages.DepositFormPage{
		BasePage: h.basePage(r, "overview", "Nova Entrada"),
	})
}

func (h *Handler) DepositPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/deposits/new", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form error", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		h.render(w, "deposit_form", pages.DepositFormPage{
			BasePage: h.basePage(r, "overview", "Nova Entrada"),
			Error:    "Valor inválido.",
		})
		return
	}
	date, _ := time.Parse("2006-01-02", r.FormValue("date"))
	if date.IsZero() {
		date = time.Now()
	}
	h.store.AddDeposit(models.Deposit{
		Description: strings.TrimSpace(r.FormValue("description")),
		Amount:      amount,
		Date:        date,
	})
	http.Redirect(w, r, "/?tab=overview&flash=added", http.StatusSeeOther)
}

// ── RESET ────────────────────────────────────────────────────────────────

func (h *Handler) ResetConfirm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "reset_confirm", pages.ResetConfirmPage{
		BasePage: h.basePage(r, "overview", "Resetar Dados"),
	})
}

func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/reset/confirm", http.StatusFound)
		return
	}
	h.store.ResetAll()
	http.Redirect(w, r, "/?flash=reset", http.StatusSeeOther)
}
