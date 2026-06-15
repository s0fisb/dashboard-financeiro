GO      := go
BINARY  := bin/server
PORT    := 8081

.PHONY: all build run clean dev

all: build

build:
	@echo "→ Compilando backend Go..."
	@mkdir -p bin
	$(GO) build -o $(BINARY) ./cmd/server/
	@echo "✅ Build completo: $(BINARY)"

run: build
	@echo "→ Iniciando FinançasLua na porta $(PORT)..."
	@PORT=$(PORT) ./$(BINARY)

dev:
	@echo "→ Modo desenvolvimento (rebuild automático)..."
	@which air > /dev/null 2>&1 && air || $(MAKE) run

clean:
	@rm -rf bin/
	@echo "→ Binários removidos."

test:
	$(GO) test ./...

# Mostra os scripts Lua carregados
lua-list:
	@ls scripts/lua/

help:
	@echo "FinançasLua — Dashboard Financeiro"
	@echo ""
	@echo "Comandos:"
	@echo "  make build    Compila o servidor Go"
	@echo "  make run      Compila e inicia o servidor"
	@echo "  make clean    Remove binários"
	@echo "  make test     Roda os testes"
	@echo ""
	@echo "Acesse: http://localhost:$(PORT)"
