package lua

import (
	"fmt"
	"os"
	"strings"

	glua "github.com/yuin/gopher-lua"

	"financial-dashboard/internal/models"
)

type Engine struct {
	forecastSrc string
	insightsSrc string
}

// NewEngine loads the Lua scripts from disk. Returns an error if the files
// cannot be read (server startup should fail fast in that case).
func NewEngine() (*Engine, error) {
	fc, err := os.ReadFile("scripts/lua/forecast.lua")
	if err != nil {
		return nil, fmt.Errorf("forecast.lua: %w", err)
	}
	ins, err := os.ReadFile("scripts/lua/insights.lua")
	if err != nil {
		return nil, fmt.Errorf("insights.lua: %w", err)
	}
	return &Engine{forecastSrc: string(fc), insightsSrc: string(ins)}, nil
}

// RunForecast executes weighted_forecast(expenses) from forecast.lua.
// A fresh Lua state is created per call because gopher-lua is not goroutine-safe.
func (e *Engine) RunForecast(history []float64) float64 {
	L := glua.NewState()
	defer L.Close()
	if err := L.DoString(e.forecastSrc); err != nil {
		return 0
	}
	tbl := L.NewTable()
	for i, v := range history {
		L.RawSetInt(tbl, i+1, glua.LNumber(v))
	}
	if err := L.CallByParam(glua.P{
		Fn:      L.GetGlobal("weighted_forecast"),
		NRet:    1,
		Protect: true,
	}, tbl); err != nil {
		return 0
	}
	result := float64(L.ToNumber(-1))
	L.Pop(1)
	return result
}

// RunInsights executes analyze_health(income, expenses, goals) from insights.lua.
// Returns the insight messages joined with " | ".
func (e *Engine) RunInsights(summary models.DashboardSummary) string {
	L := glua.NewState()
	defer L.Close()
	if err := L.DoString(e.insightsSrc); err != nil {
		return ""
	}
	goalsTbl := L.NewTable()
	for i, g := range summary.Goals {
		gt := L.NewTable()
		L.SetField(gt, "name", glua.LString(g.Name))
		L.SetField(gt, "target", glua.LNumber(g.TargetAmount))
		L.SetField(gt, "current", glua.LNumber(g.CurrentAmount))
		L.RawSetInt(goalsTbl, i+1, gt)
	}
	if err := L.CallByParam(glua.P{
		Fn:      L.GetGlobal("analyze_health"),
		NRet:    1,
		Protect: true,
	}, glua.LNumber(summary.TotalIncome), glua.LNumber(summary.TotalExpenses), goalsTbl); err != nil {
		return ""
	}
	lv := L.Get(-1)
	L.Pop(1)
	result, ok := lv.(*glua.LTable)
	if !ok {
		return ""
	}
	var msgs []string
	result.ForEach(func(_, v glua.LValue) {
		tbl, ok := v.(*glua.LTable)
		if !ok {
			return
		}
		if s, ok := tbl.RawGetString("msg").(glua.LString); ok {
			msgs = append(msgs, string(s))
		}
	})
	return strings.Join(msgs, " | ")
}

func (e *Engine) ScriptSource(name string) string {
	switch name {
	case "forecast":
		return e.forecastSrc
	case "insights":
		return e.insightsSrc
	}
	return ""
}
