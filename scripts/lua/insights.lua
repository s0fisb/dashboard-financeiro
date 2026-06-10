-- insights.lua
-- Análise de saúde financeira e geração de insights
-- Executado pelo gopher-lua engine no backend Go

function analyze_health(income, expenses, goals)
  local insights = {}
  local savings_rate = 0

  if income > 0 then
    savings_rate = (income - expenses) / income * 100
  end

  -- Regra 50/30/20
  local needs_limit   = income * 0.50
  local wants_limit   = income * 0.30
  local savings_limit = income * 0.20

  if savings_rate < 0 then
    table.insert(insights, {
      level = "critical",
      msg   = "Gastos superam a renda. Reveja o orçamento imediatamente."
    })
  elseif savings_rate < 10 then
    table.insert(insights, {
      level = "warning",
      msg   = string.format("Taxa de poupança: %.1f%%. Tente reduzir gastos variáveis.", savings_rate)
    })
  elseif savings_rate >= 20 then
    table.insert(insights, {
      level = "success",
      msg   = string.format("Ótima taxa de poupança: %.1f%%!", savings_rate)
    })
  end

  -- Metas
  for _, g in ipairs(goals) do
    local pct = 0
    if g.target > 0 then
      pct = g.current / g.target * 100
    end
    if pct >= 100 then
      table.insert(insights, {level = "success", msg = "Meta '" .. g.name .. "' concluída!"})
    elseif pct >= 75 then
      table.insert(insights, {level = "info", msg = string.format("Meta '%s': %.0f%% — quase lá!", g.name, pct)})
    end
  end

  return insights
end

-- check_budget: verifica status do orçamento por categoria
function check_budget(spent, limit)
  if limit <= 0 then return "sem_limite" end
  local pct = spent / limit * 100
  if pct >= 100 then return "excedido"
  elseif pct >= 90 then return "critico"
  elseif pct >= 70 then return "atencao"
  else return "ok"
  end
end
