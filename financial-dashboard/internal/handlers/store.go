package handlers

import (
	"math/rand"
	"time"

	"financial-dashboard/internal/models"
)

// Store holds in-memory data (replace with SQLite/Postgres in production)
type Store struct {
	Expenses []models.Expense
	Goals    []models.Goal
	Budget   models.MonthlyBudget
}

func NewStore() *Store {
	s := &Store{}
	s.seedData()
	return s
}

func (s *Store) seedData() {
	now := time.Now()

	// --- Expenses ---
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
			descList := descs[cat]
			desc := descList[r.Intn(len(descList))]

			var amount float64
			switch cat {
			case models.CategoryHousing:
				amount = 400 + r.Float64()*600
			case models.CategoryHealth:
				amount = 30 + r.Float64()*200
			case models.CategoryFood:
				amount = 15 + r.Float64()*150
			case models.CategoryTransport:
				amount = 10 + r.Float64()*100
			case models.CategoryLeisure:
				amount = 20 + r.Float64()*200
			case models.CategoryEducation:
				amount = 50 + r.Float64()*300
			default:
				amount = 20 + r.Float64()*150
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

	// Previous months (for trend)
	for m := -5; m < 0; m++ {
		month := now.AddDate(0, m, 0)
		numExp := r.Intn(20) + 25
		for i := 0; i < numExp; i++ {
			cat := categories[r.Intn(len(categories))]
			descList := descs[cat]
			desc := descList[r.Intn(len(descList))]
			amount := 20 + r.Float64()*500
			amount = float64(int(amount*100)) / 100
			day := r.Intn(28) + 1
			date := time.Date(month.Year(), month.Month(), day, 12, 0, 0, 0, time.Local)
			s.Expenses = append(s.Expenses, models.Expense{
				ID: id, Description: desc, Amount: amount,
				Category: cat, Date: date,
			})
			id++
		}
	}

	// --- Goals ---
	s.Goals = []models.Goal{
		{ID: 1, Name: "Reserva de Emergência", TargetAmount: 15000, CurrentAmount: 8700, Deadline: now.AddDate(0, 6, 0), Color: "#4ade80", Icon: "🛡️"},
		{ID: 2, Name: "Viagem Europa", TargetAmount: 12000, CurrentAmount: 3200, Deadline: now.AddDate(1, 0, 0), Color: "#60a5fa", Icon: "✈️"},
		{ID: 3, Name: "Notebook Novo", TargetAmount: 4500, CurrentAmount: 4200, Deadline: now.AddDate(0, 1, 0), Color: "#f472b6", Icon: "💻"},
		{ID: 4, Name: "Entrada Apartamento", TargetAmount: 80000, CurrentAmount: 22000, Deadline: now.AddDate(3, 0, 0), Color: "#fb923c", Icon: "🏠"},
	}

	// --- Budget ---
	s.Budget = models.MonthlyBudget{
		Month:  int(now.Month()),
		Year:   now.Year(),
		Income: 8500,
		Budgets: map[models.Category]float64{
			models.CategoryFood:      1200,
			models.CategoryTransport: 400,
			models.CategoryHealth:    300,
			models.CategoryLeisure:   500,
			models.CategoryEducation: 600,
			models.CategoryHousing:   2500,
			models.CategoryOther:     300,
		},
	}
}

// GetCurrentMonthExpenses returns expenses for the current month.
func (s *Store) GetCurrentMonthExpenses() []models.Expense {
	now := time.Now()
	var result []models.Expense
	for _, e := range s.Expenses {
		if e.Date.Month() == now.Month() && e.Date.Year() == now.Year() {
			result = append(result, e)
		}
	}
	return result
}

// GetMonthlyTotals returns total expenses grouped by month (last 6).
func (s *Store) GetMonthlyTotals() map[string]float64 {
	totals := make(map[string]float64)
	for _, e := range s.Expenses {
		key := e.Date.Format("2006-01")
		totals[key] += e.Amount
	}
	return totals
}

// AddExpense adds a new expense to the store.
func (s *Store) AddExpense(e models.Expense) models.Expense {
	e.ID = len(s.Expenses) + 1
	e.Date = time.Now()
	s.Expenses = append(s.Expenses, e)
	return e
}
