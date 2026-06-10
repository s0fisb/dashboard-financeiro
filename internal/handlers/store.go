package handlers

import (
	"math/rand"
	"sync"
	"time"

	"financial-dashboard/internal/models"
)

// Store holds in-memory data with full CRUD support.
// Replace with SQLite/Postgres in production.
type Store struct {
	mu       sync.RWMutex
	Expenses []models.Expense
	Goals    []models.Goal
	Budget   models.MonthlyBudget
	Deposits []models.Deposit
	nextExpID int
	nextGoalID int
	nextDepID  int
}

func NewStore() *Store {
	s := &Store{}
	s.seedData()
	return s
}

func (s *Store) seedData() {
	now := time.Now()
	categories := []models.Category{
		models.CategoryFood, models.CategoryTransport, models.CategoryHealth,
		models.CategoryLeisure, models.CategoryEducation, models.CategoryHousing, models.CategoryOther,
	}
	descs := map[models.Category][]string{
		models.CategoryFood:      {"Supermercado Extra", "iFood", "Restaurante Madero", "Padaria Dom Pão", "Café Amazon"},
		models.CategoryTransport: {"Uber", "Gasolina Posto Shell", "Bilhete TRI", "Manutenção carro", "Estacionamento"},
		models.CategoryHealth:    {"Farmácia São João", "Consulta médica", "Academia Smart Fit", "Plano de saúde", "Dentista"},
		models.CategoryLeisure:   {"Netflix", "Spotify", "Cinema", "Shopping", "Viagem curta"},
		models.CategoryEducation: {"Udemy", "Livros Amazon", "Curso inglês", "MBA parcela", "Material escolar"},
		models.CategoryHousing:   {"Aluguel", "Condomínio", "Luz CEMIG", "Internet Vivo", "Água SAAE"},
		models.CategoryOther:     {"Presente aniversário", "Roupas", "Eletrônicos", "Doação", "Pets"},
	}

	id := 1
	r := rand.New(rand.NewSource(42))

	for day := 1; day <= now.Day(); day++ {
		numTx := r.Intn(3) + 1
		for t := 0; t < numTx; t++ {
			cat := categories[r.Intn(len(categories))]
			desc := descs[cat][r.Intn(len(descs[cat]))]
			var amount float64
			switch cat {
			case models.CategoryHousing:    amount = 400 + r.Float64()*600
			case models.CategoryHealth:     amount = 30 + r.Float64()*200
			case models.CategoryFood:       amount = 15 + r.Float64()*150
			case models.CategoryTransport:  amount = 10 + r.Float64()*100
			case models.CategoryLeisure:    amount = 20 + r.Float64()*200
			case models.CategoryEducation:  amount = 50 + r.Float64()*300
			default:                        amount = 20 + r.Float64()*150
			}
			amount = float64(int(amount*100)) / 100
			date := time.Date(now.Year(), now.Month(), day, r.Intn(18)+7, r.Intn(60), 0, 0, time.Local)
			s.Expenses = append(s.Expenses, models.Expense{
				ID: id, Description: desc, Amount: amount,
				Category: cat, Date: date, IsRecurring: r.Float64() > 0.8,
			})
			id++
		}
	}
	for m := -5; m < 0; m++ {
		month := now.AddDate(0, m, 0)
		for i := 0; i < r.Intn(20)+25; i++ {
			cat := categories[r.Intn(len(categories))]
			desc := descs[cat][r.Intn(len(descs[cat]))]
			amount := float64(int((20+r.Float64()*500)*100)) / 100
			date := time.Date(month.Year(), month.Month(), r.Intn(28)+1, 12, 0, 0, 0, time.Local)
			s.Expenses = append(s.Expenses, models.Expense{ID: id, Description: desc, Amount: amount, Category: cat, Date: date})
			id++
		}
	}
	s.nextExpID = id

	s.Goals = []models.Goal{
		{ID: 1, Name: "Reserva de Emergência", TargetAmount: 15000, CurrentAmount: 8700, Deadline: now.AddDate(0, 6, 0), Color: "#4ade80", Icon: "🛡️"},
		{ID: 2, Name: "Viagem Europa",         TargetAmount: 12000, CurrentAmount: 3200, Deadline: now.AddDate(1, 0, 0), Color: "#60a5fa", Icon: "✈️"},
		{ID: 3, Name: "Notebook Novo",         TargetAmount: 4500,  CurrentAmount: 4200, Deadline: now.AddDate(0, 1, 0), Color: "#f472b6", Icon: "💻"},
		{ID: 4, Name: "Entrada Apartamento",   TargetAmount: 80000, CurrentAmount: 22000, Deadline: now.AddDate(3, 0, 0), Color: "#fb923c", Icon: "🏠"},
	}
	s.nextGoalID = 5
	s.nextDepID = 1

	s.Budget = models.MonthlyBudget{
		Month: int(now.Month()), Year: now.Year(), Income: 8500,
		Budgets: map[models.Category]float64{
			models.CategoryFood: 1200, models.CategoryTransport: 400, models.CategoryHealth: 300,
			models.CategoryLeisure: 500, models.CategoryEducation: 600, models.CategoryHousing: 2500,
			models.CategoryOther: 300,
		},
	}
}

// ── EXPENSES ─────────────────────────────────

func (s *Store) GetCurrentMonthExpenses() []models.Expense {
	s.mu.RLock(); defer s.mu.RUnlock()
	now := time.Now()
	var result []models.Expense
	for _, e := range s.Expenses {
		if e.Date.Month() == now.Month() && e.Date.Year() == now.Year() {
			result = append(result, e)
		}
	}
	return result
}

func (s *Store) GetAllExpenses() []models.Expense {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]models.Expense, len(s.Expenses))
	copy(out, s.Expenses)
	return out
}

func (s *Store) AddExpense(e models.Expense) models.Expense {
	s.mu.Lock(); defer s.mu.Unlock()
	e.ID = s.nextExpID
	s.nextExpID++
	if e.Date.IsZero() { e.Date = time.Now() }
	s.Expenses = append(s.Expenses, e)
	return e
}

func (s *Store) UpdateExpense(updated models.Expense) (models.Expense, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	for i, e := range s.Expenses {
		if e.ID == updated.ID {
			if updated.Date.IsZero() { updated.Date = e.Date }
			s.Expenses[i] = updated
			return updated, true
		}
	}
	return models.Expense{}, false
}

func (s *Store) DeleteExpense(id int) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	for i, e := range s.Expenses {
		if e.ID == id {
			s.Expenses = append(s.Expenses[:i], s.Expenses[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) GetMonthlyTotals() map[string]float64 {
	s.mu.RLock(); defer s.mu.RUnlock()
	totals := make(map[string]float64)
	for _, e := range s.Expenses {
		totals[e.Date.Format("2006-01")] += e.Amount
	}
	return totals
}

// ── GOALS ─────────────────────────────────────

func (s *Store) GetGoals() []models.Goal {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]models.Goal, len(s.Goals))
	copy(out, s.Goals)
	return out
}

func (s *Store) AddGoal(g models.Goal) models.Goal {
	s.mu.Lock(); defer s.mu.Unlock()
	g.ID = s.nextGoalID
	s.nextGoalID++
	s.Goals = append(s.Goals, g)
	return g
}

func (s *Store) UpdateGoal(updated models.Goal) (models.Goal, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	for i, g := range s.Goals {
		if g.ID == updated.ID {
			s.Goals[i] = updated
			return updated, true
		}
	}
	return models.Goal{}, false
}

func (s *Store) DeleteGoal(id int) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	for i, g := range s.Goals {
		if g.ID == id {
			s.Goals = append(s.Goals[:i], s.Goals[i+1:]...)
			return true
		}
	}
	return false
}

// ── BUDGET / INCOME ───────────────────────────

func (s *Store) GetBudget() models.MonthlyBudget {
	s.mu.RLock(); defer s.mu.RUnlock()
	return s.Budget
}

func (s *Store) UpdateBudget(b models.MonthlyBudget) {
	s.mu.Lock(); defer s.mu.Unlock()
	s.Budget = b
}

// ── DEPOSITS ──────────────────────────────────

func (s *Store) AddDeposit(d models.Deposit) models.Deposit {
	s.mu.Lock(); defer s.mu.Unlock()
	d.ID = s.nextDepID
	s.nextDepID++
	if d.Date.IsZero() { d.Date = time.Now() }
	s.Deposits = append(s.Deposits, d)
	return d
}

func (s *Store) GetDeposits() []models.Deposit {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]models.Deposit, len(s.Deposits))
	copy(out, s.Deposits)
	return out
}

func (s *Store) GetCurrentMonthDeposits() []models.Deposit {
	s.mu.RLock(); defer s.mu.RUnlock()
	now := time.Now()
	var result []models.Deposit
	for _, d := range s.Deposits {
		if d.Date.Month() == now.Month() && d.Date.Year() == now.Year() {
			result = append(result, d)
		}
	}
	return result
}

// ── RESET ─────────────────────────────────────

func (s *Store) ResetAll() {
	s.mu.Lock(); defer s.mu.Unlock()
	now := time.Now()
	s.Expenses = nil
	s.Goals = nil
	s.Deposits = nil
	s.nextExpID = 1
	s.nextGoalID = 1
	s.nextDepID = 1
	s.Budget = models.MonthlyBudget{
		Month: int(now.Month()), Year: now.Year(), Income: 0,
		Budgets: map[models.Category]float64{
			models.CategoryFood: 0, models.CategoryTransport: 0, models.CategoryHealth: 0,
			models.CategoryLeisure: 0, models.CategoryEducation: 0, models.CategoryHousing: 0,
			models.CategoryOther: 0,
		},
	}
}
