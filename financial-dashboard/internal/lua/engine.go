// Package lua provides a lightweight Lua script runner for financial analysis.
// In production, replace with gopher-lua (github.com/yuin/gopher-lua).
// This package simulates Lua execution and returns computed insights.
package lua

import (
	"fmt"
	"math"
	"strings"

	"financial-dashboard/internal/models"
)

// Engine simulates execution of Lua financial scripts.
// Production version would embed a real Lua VM via gopher-lua.
type Engine struct {
	scripts map[string]string
}

// NewEngine creates a new Lua engine with preloaded scripts.
func NewEngine() *Engine {
	e := &Engine{scripts: make(map[string]string)}
	e.loadDefaultScripts()
	return e
}

// loadDefaultScripts registers all Lua analysis scripts.
// These represent what would be .lua files loaded at runtime.
func (e *Engine) loadDefaultScripts() {
	e.scripts["forecast"] = `
-- forecast.lua: Weighted moving average forecast
function forecast(expenses)
  local weights = {0.1, 0.15, 0.25, 0.5}
  local total = 0
  local weight_sum = 0
  for i, v in ipairs(expenses) do
    local w = weights[i] or 0.1
    total = total + v * w
    weight_sum = weight_sum + w
  end
  return total / weight_sum
end
`

	e.scripts["insights"] = `
-- insights.lua: Generate financial health insights
function analyze(expenses, income, goals)
  local savings_rate = (income - expenses) / income * 100
  local insights = {}
  if savings_rate < 10 then
    table.insert(insights, "⚠️ Taxa de poupança abaixo de 10%")
  elseif savings_rate > 30 then
    table.insert(insights, "✅ Excelente taxa de poupança!")
  end
  return insights
end
`

	e.scripts["budget_check"] = `
-- budget_check.lua: Check budget limits per category
function check_budget(spent, limit)
  local pct = spent / limit * 100
  if pct > 90 then return "crítico"
  elseif pct > 70 then return "atenção"
  else return "ok" end
end
`
}

// RunForecast executes the Lua forecast script on expense history.
func (e *Engine) RunForecast(history []float64) float64 {
	// Weighted moving average — mirrors the Lua forecast.lua logic
	weights := []float64{0.1, 0.15, 0.25, 0.5}
	total, weightSum := 0.0, 0.0
	n := len(history)
	for i, v := range history {
		wi := 0.10
		idx := i - (n - len(weights))
		if idx >= 0 && idx < len(weights) {
			wi = weights[idx]
		}
		total += v * wi
		weightSum += wi
	}
	if weightSum == 0 {
		return 0
	}
	return math.Round(total/weightSum*100) / 100
}

// RunInsights executes insights.lua to produce financial health text.
func (e *Engine) RunInsights(expenses models.DashboardSummary) string {
	savingsRate := 0.0
	if expenses.TotalIncome > 0 {
		savingsRate = (expenses.TotalIncome - expenses.TotalExpenses) / expenses.TotalIncome * 100
	}

	var lines []string

	// Mirror insights.lua logic
	switch {
	case savingsRate < 0:
		lines = append(lines, "🚨 Gastos superam a renda este mês — reveja o orçamento imediatamente.")
	case savingsRate < 10:
		lines = append(lines, "⚠️ Taxa de poupança abaixo de 10% — considere cortar gastos variáveis.")
	case savingsRate < 20:
		lines = append(lines, "📊 Taxa de poupança razoável ("+fmt.Sprintf("%.1f", savingsRate)+"%). Tente chegar a 20%.")
	case savingsRate >= 30:
		lines = append(lines, "✅ Excelente! Taxa de poupança de "+fmt.Sprintf("%.1f", savingsRate)+"% — continue assim.")
	default:
		lines = append(lines, "👍 Boa taxa de poupança ("+fmt.Sprintf("%.1f", savingsRate)+"%). Metas no caminho certo.")
	}

	// Top spending category alert (budget_check.lua logic)
	for cat, amt := range expenses.ExpensesByCategory {
		if budget, ok := expenses.BudgetUsage[cat]; ok && budget > 0 {
			pct := amt / budget * 100
			if pct > 90 {
				lines = append(lines, "🔴 "+strings.Title(cat)+": orçamento crítico ("+fmt.Sprintf("%.0f", pct)+"% usado).")
			} else if pct > 70 {
				lines = append(lines, "🟡 "+strings.Title(cat)+": atenção ao orçamento ("+fmt.Sprintf("%.0f", pct)+"% usado).")
			}
		}
	}

	// Goal progress
	for _, g := range expenses.Goals {
		if g.TargetAmount > 0 {
			pct := g.CurrentAmount / g.TargetAmount * 100
			if pct >= 100 {
				lines = append(lines, "🏆 Meta \""+g.Name+"\" atingida!")
			} else if pct >= 75 {
				lines = append(lines, "🎯 Meta \""+g.Name+"\": "+fmt.Sprintf("%.0f", pct)+"% — quase lá!")
			}
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "📈 Finanças equilibradas. Continue monitorando seus gastos.")
	}

	return strings.Join(lines, " | ")
}

// ScriptSource returns the raw Lua source for display/documentation.
func (e *Engine) ScriptSource(name string) string {
	if s, ok := e.scripts[name]; ok {
		return s
	}
	return ""
}
