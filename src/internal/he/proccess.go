package he

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"log"
)

// Tracker keeps track of financial stats for a single game (income, expense, ROI, HE).
type Tracker struct {
	Income  float64
	Expense float64
	ROI     float64
	HE      float64
}

// NewTracker creates and returns a new Tracker instance.
func NewTracker() *Tracker {
	return &Tracker{}
}

// AddIncome increments total income (player deposits, bets, etc.).
func (t *Tracker) AddIncome(amount float64) {
	t.Income += amount
}

// AddExpense increments total expense (payouts, rewards, etc.).
func (t *Tracker) AddExpense(amount float64) {
	t.Expense += amount
}

// calRatio calculates ROI (Return on Investment).
// ROI = (income / expense) * 100
// If expense = 0, special cases are handled.
func (t *Tracker) calRatio() {
	if t.Expense == 0 {
		if t.Income == 0 {
			t.ROI = 100 // Neutral: no income, no expense
			return
		}
		t.ROI = 999999 // Special case: division by zero
		return
	}
	t.ROI = (t.Income / t.Expense) * 100
}

// CalHouseEdge calculates the House Edge.
// HE = (income - expense) / income * 100
func (t *Tracker) CalHouseEdge() {
	if t.Income == 0 {
		t.HE = 0
		return
	}
	t.HE = (t.Income - t.Expense) / t.Income * 100
}

// Save computes ROI and HE, then persists all values to the database.
func (t *Tracker) Save(gameTable string, gameID int) {
	t.calRatio()
	t.CalHouseEdge()

	query := fmt.Sprintf(
		`UPDATE %s SET income=%.2f, expense=%.2f, roi=%.2f, he=%.2f WHERE id=%d`,
		gameTable,
		t.Income,
		t.Expense,
		t.ROI,
		t.HE,
		gameID,
	)
	log.Println(query)

	grpcclient.SendQuery(query)

}

func _Example() {

	tracker := NewTracker()

	tracker.AddIncome(1000)
	tracker.AddExpense(930)

	tracker.Save("g1_games", 123)

}
