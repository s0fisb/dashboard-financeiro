-- forecast.lua
-- Previsão de gastos com média móvel ponderada
-- Executado pelo gopher-lua via internal/lua/engine.go

function weighted_forecast(expenses)
  local weights = {0.1, 0.15, 0.25, 0.5}
  local n = #expenses
  local total = 0
  local weight_sum = 0

  for i = 1, n do
    local offset = i - (n - #weights)
    local w = 0.1
    if offset >= 1 and offset <= #weights then
      w = weights[offset]
    end
    total = total + expenses[i] * w
    weight_sum = weight_sum + w
  end

  if weight_sum == 0 then return 0 end
  return math.floor(total / weight_sum * 100 + 0.5) / 100
end

-- Retorna tendência: "crescente", "estável", "decrescente"
function trend(expenses)
  if #expenses < 2 then return "estável" end
  local last = expenses[#expenses]
  local prev = expenses[#expenses - 1]
  local diff = (last - prev) / prev * 100
  if diff > 5 then return "crescente"
  elseif diff < -5 then return "decrescente"
  else return "estável"
  end
end
