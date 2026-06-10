# FinançasLua — Dashboard Financeiro

Dashboard financeiro pessoal construído com **Go** no backend e **Lua** como motor de análise e scripting.

```
┌─────────────────────────────────────────────────────┐
│  FinançasLua                                        │
│  Backend: Go 1.22 (net/http stdlib)                 │
│  Scripts: Lua (gopher-lua em produção)              │
│  Frontend: HTML + CSS + Chart.js                    │
└─────────────────────────────────────────────────────┘
```

## Funcionalidades

| Módulo            | Descrição                                              |
|-------------------|--------------------------------------------------------|
| **Visão Geral**   | KPIs, gráfico de pizza por categoria, tendência mensal |
| **Gastos**        | Listagem e adição de gastos por categoria              |
| **Metas**         | Cards de progresso com prazo e valor atual             |
| **Calendário**    | Eventos financeiros do mês com navegação               |
| **Previsão**      | Gráfico de linha com previsão via algoritmo Lua        |
| **Scripts Lua**   | Visualização dos scripts rodando no backend            |

## Arquitetura

```
financial-dashboard/
├── cmd/server/
│   └── main.go              ← Servidor HTTP Go (stdlib net/http)
├── internal/
│   ├── handlers/
│   │   ├── handlers.go      ← Handlers HTTP (dashboard, expenses, goals…)
│   │   └── store.go         ← Store in-memory com dados de exemplo
│   ├── lua/
│   │   └── engine.go        ← Motor Lua (Go puro; usa gopher-lua em prod)
│   └── models/
│       └── models.go        ← Structs de domínio (Expense, Goal, Forecast…)
├── scripts/lua/
│   ├── forecast.lua         ← Algoritmo de previsão (média ponderada)
│   └── insights.lua         ← Análise de saúde financeira
├── web/
│   ├── index.html           ← Interface principal
│   └── static/
│       ├── css/dashboard.css
│       └── js/dashboard.js
├── go.mod
└── Makefile
```

## Como Rodar

### Pré-requisitos
- Go 1.22+

### Build e execução

```bash
# Clonar / entrar na pasta
cd financial-dashboard

# Compilar e rodar
make run

# Ou manualmente:
go build -o bin/server ./cmd/server/
./bin/server
```

Acesse: **http://localhost:8080**

## API REST

| Método | Rota               | Descrição                              |
|--------|--------------------|----------------------------------------|
| GET    | `/api/dashboard`   | Resumo completo + insights Lua         |
| GET    | `/api/expenses`    | Gastos do mês atual                    |
| POST   | `/api/expenses`    | Adicionar novo gasto                   |
| GET    | `/api/goals`       | Metas financeiras                      |
| GET    | `/api/calendar`    | Eventos do calendário                  |
| GET    | `/api/lua-scripts` | Código-fonte dos scripts Lua           |

### Exemplo — Adicionar gasto

```bash
curl -X POST http://localhost:8080/api/expenses \
  -H "Content-Type: application/json" \
  -d '{"description":"Mercado","amount":187.50,"category":"alimentação"}'
```

## Scripts Lua

Os scripts em `scripts/lua/` são executados pelo backend Go via **gopher-lua**.
Para produção, adicione a dependência:

```bash
go get github.com/yuin/gopher-lua
```

### forecast.lua — Previsão de Gastos

Implementa média ponderada com pesos `[0.10, 0.15, 0.25, 0.50]` sobre os
últimos 4 meses, dando mais peso aos dados mais recentes.

### insights.lua — Análise de Saúde Financeira

Calcula taxa de poupança, verifica regra 50/30/20, alertas de orçamento
por categoria e progresso de metas.

## Próximos Passos

- [ ] Persistência com SQLite (`database/sql`)
- [ ] Autenticação JWT
- [ ] Integração com gopher-lua para scripts dinâmicos
- [ ] Export CSV/PDF dos relatórios
- [ ] App mobile com Go + WebView
- [ ] Hot-reload de scripts Lua sem reiniciar o servidor

## Tecnologias

- **Go 1.22** — net/http, encoding/json, embed
- **Lua 5.1** — via gopher-lua (github.com/yuin/gopher-lua)
- **Chart.js 4** — gráficos interativos
- **Google Fonts** — Syne, DM Mono, DM Sans
