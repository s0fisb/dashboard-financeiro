package models

import "time"

type Category string

const (
	CategoryFood        Category = "alimentação"
	CategoryTransport   Category = "transporte"
	CategoryHealth      Category = "saúde"
	CategoryLeisure     Category = "lazer"
	CategoryEducation   Category = "educação"
	CategoryHousing     Category = "moradia"
	CategoryOther       Category = "outros"
)

type Expense struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    Category  `json:"category"`
	Date        time.Time `json:"date"`
	IsRecurring bool      `json:"is_recurring"`
}

type Goal struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	TargetAmount float64  `json:"target_amount"`
	CurrentAmount float64 `json:"current_amount"`
	Deadline    time.Time `json:"deadline"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
}

type MonthlyBudget struct {
	Month    int            `json:"month"`
	Year     int            `json:"year"`
	Income   float64        `json:"income"`
	Budgets  map[Category]float64 `json:"budgets"`
}

type Forecast struct {
	Month       string  `json:"month"`
	Predicted   float64 `json:"predicted"`
	Actual      float64 `json:"actual"`
	Income      float64 `json:"income"`
}

type CalendarEvent struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Type        string    `json:"type"` // "expense", "income", "goal"
	Category    Category  `json:"category"`
}

type DashboardSummary struct {
	TotalExpenses    float64            `json:"total_expenses"`
	TotalIncome      float64            `json:"total_income"`
	Balance          float64            `json:"balance"`
	ExpensesByCategory map[string]float64 `json:"expenses_by_category"`
	MonthlyTrend     []Forecast         `json:"monthly_trend"`
	Goals            []Goal             `json:"goals"`
	RecentExpenses   []Expense          `json:"recent_expenses"`
	BudgetUsage      map[string]float64 `json:"budget_usage"`
	LuaInsights      string             `json:"lua_insights"`
}
