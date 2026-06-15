package pages

import (
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"

	"financial-dashboard/internal/models"
)

// sprint converts any value to string, handling named string types like models.Category.
func sprint(v interface{}) string { return fmt.Sprint(v) }

// ── HELPERS ────────────────────────────────────────────────────────────────

// FmtBRL formats a float as Brazilian Real: 1234.5 → "R$ 1.234,50"
func FmtBRL(v float64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	cents := int(math.Round(v * 100))
	reais := cents / 100
	frac := cents % 100

	s := fmt.Sprintf("%d", reais)
	// Insert thousands dot separators
	if len(s) > 3 {
		var parts []string
		for len(s) > 3 {
			parts = append([]string{s[len(s)-3:]}, parts...)
			s = s[:len(s)-3]
		}
		parts = append([]string{s}, parts...)
		s = strings.Join(parts, ".")
	}
	result := fmt.Sprintf("R$ %s,%02d", s, frac)
	if neg {
		return "-" + result
	}
	return result
}

func FmtDate(t time.Time) string {
	return t.Format("02/01")
}

func FmtDateInput(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func FmtMonth(year, month int) string {
	months := []string{"", "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
		"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}
	return fmt.Sprintf("%s %d", months[month], year)
}

var catIcons = map[string]string{
	"alimentação": "🍔",
	"transporte":  "🚗",
	"saúde":       "💊",
	"lazer":       "🎮",
	"educação":    "📚",
	"moradia":     "🏠",
	"outros":      "📦",
}

var catColors = map[string]string{
	"alimentação": "#f87171",
	"transporte":  "#60a5fa",
	"saúde":       "#34d399",
	"lazer":       "#a78bfa",
	"educação":    "#fbbf24",
	"moradia":     "#f5a623",
	"outros":      "#94a3b8",
}

func CatIcon(cat string) string {
	if ic, ok := catIcons[cat]; ok {
		return ic
	}
	return "📦"
}

func CatColor(cat string) string {
	if c, ok := catColors[cat]; ok {
		return c
	}
	return "#6b7280"
}

func BudgetClass(pct float64) string {
	switch {
	case pct >= 90:
		return "danger"
	case pct >= 70:
		return "warn"
	default:
		return "ok"
	}
}

func ClampPct(pct float64) float64 {
	if pct > 100 {
		return 100
	}
	if pct < 0 {
		return 0
	}
	return pct
}

func FlashMsg(f string) string {
	switch f {
	case "added":
		return "Adicionado com sucesso."
	case "updated":
		return "Atualizado com sucesso."
	case "deleted":
		return "Removido com sucesso."
	case "saved":
		return "Salvo com sucesso."
	case "deposited":
		return "Depósito realizado com sucesso."
	case "reset":
		return "Dados resetados com sucesso."
	}
	return ""
}

func FlashType(f string) string {
	if f == "deleted" || f == "reset" {
		return "error"
	}
	return "success"
}

// FuncMap returns the template function map for all pages.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"fmtBRL":      FmtBRL,
		"fmtDate":     FmtDate,
		"fmtDateInput": FmtDateInput,
		"fmtMonth":    FmtMonth,
		"catIcon":  func(cat interface{}) string { return CatIcon(sprint(cat)) },
		"catColor": func(cat interface{}) string { return CatColor(sprint(cat)) },
		"budgetClass": BudgetClass,
		"clampPct":    ClampPct,
		"flashMsg":    FlashMsg,
		"flashType":   FlashType,
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"add": func(a, b int) int { return a + b },
		"minF": func(a, b float64) float64 {
			if a < b {
				return a
			}
			return b
		},
		"sumCalEvents": func(evs []models.CalendarEvent) float64 {
			var t float64
			for _, e := range evs {
				t += e.Amount
			}
			return t
		},
		"pctStr": func(curr, target float64) string {
			if target <= 0 {
				return "0%"
			}
			p := curr / target * 100
			if p > 100 {
				p = 100
			}
			return fmt.Sprintf("%.0f%%", p)
		},
		"budgetVal": func(budgets map[models.Category]float64, cat string) float64 {
			return budgets[models.Category(cat)]
		},
	}
}

// ── PAGE DATA STRUCTS ──────────────────────────────────────────────────────

type BasePage struct {
	Tab         string
	PageTitle   string
	PageDate    string
	LuaInsights string
	Flash       string // flash message text (pre-resolved)
	FlashType   string // "success" or "error"
}

type BudgetItem struct {
	Category string
	Icon     string
	Spent    float64
	Limit    float64
	Pct      float64
	Class    string
}

type GoalVM struct {
	models.Goal
	Pct         float64
	PctClamped  float64
	DeadlineFmt string
}

type CalEventVM struct {
	models.CalendarEvent
	DateFmt string
}

// ── OVERVIEW ───────────────────────────────────────────────────────────────

type OverviewPage struct {
	BasePage
	Income        float64
	DepositTotal  float64
	TotalExpenses float64
	Balance       float64
	ExpensePct    float64
	SavingsPct    float64
	GoalsCount    int
	GoalsComplete int
	DonutSVG      template.HTML
	BarSVG        template.HTML
	BudgetItems   []BudgetItem
	RecentExpenses []models.Expense
}

// ── EXPENSES ───────────────────────────────────────────────────────────────

var AllCategories = []string{
	"alimentação", "transporte", "saúde", "lazer", "educação", "moradia", "outros",
}

type ExpensesPage struct {
	BasePage
	Search     string
	Category   string
	Expenses   []models.Expense
	Categories []string
}

// ── GOALS ──────────────────────────────────────────────────────────────────

type GoalsPage struct {
	BasePage
	Goals []GoalVM
}

// ── CALENDAR ───────────────────────────────────────────────────────────────

type CalendarPage struct {
	BasePage
	Year        int
	Month       int
	MonthLabel  string
	FirstWeekday int
	DaysInMonth  int
	Today        int
	SelectedDay  int
	EventsByDay  map[int][]models.CalendarEvent
	AllEvents    []models.CalendarEvent
	ShownEvents  []CalEventVM
	PrevYear    int
	PrevMonth   int
	NextYear    int
	NextMonth   int
}

// ── FORECAST ───────────────────────────────────────────────────────────────

type ForecastPage struct {
	BasePage
	ForecastSVG template.HTML
	MonthCards  []models.Forecast
}

// ── LUA SCRIPTS ────────────────────────────────────────────────────────────

type LuaPage struct {
	BasePage
	ForecastSrc string
	InsightsSrc string
}

// ── FORM PAGES ─────────────────────────────────────────────────────────────

type ExpenseFormPage struct {
	BasePage
	IsEdit     bool
	Expense    models.Expense
	Categories []string
	Error      string
}

type GoalFormPage struct {
	BasePage
	IsEdit bool
	Goal   models.Goal
	Icons  []string
	Colors []string
	Error  string
}

type BudgetFormPage struct {
	BasePage
	Budget     models.MonthlyBudget
	Categories []string
}

type DepositFormPage struct {
	BasePage
	Error string
}

type ConfirmDeletePage struct {
	BasePage
	Name      string
	BackURL   string
	DeleteURL string
}

type ResetConfirmPage struct {
	BasePage
}

type GoalDepositFormPage struct {
	BasePage
	Goal  models.Goal
	Error string
}

// GoalIcons lists the available emoji icons for goals.
var GoalIcons = []string{"🛡️", "✈️", "💻", "🏠", "🎯", "💰", "📱", "🚗", "📚", "🎓", "🏖️", "🎸"}

// GoalColors lists the available colors for goal cards.
var GoalColors = []string{"#4ade80", "#60a5fa", "#f472b6", "#fb923c", "#a78bfa", "#fbbf24", "#34d399", "#f87171"}
