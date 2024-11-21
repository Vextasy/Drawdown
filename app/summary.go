package drawdown

// Summary is an aggregation of a set of transactions.
type DrawSummary struct {
	TotalWithdrawn int64
	TotalTaxPaid   int64
	FinalBalance   int64
	YearsFunded    int
}

// Summary returns a summary of the given DrawHistory transactions.
// Summary relies on DrawHistory being sorted by increasing year.
func (h DrawHistory) Summary() DrawSummary {
	s := DrawSummary{}
	balanceByYear := map[int]int64{}
	for _, t := range h {
		s.TotalWithdrawn += t.Amount
		s.TotalTaxPaid += t.Tax
		balanceByYear[t.Year] += t.Balance
		if t.Year > s.YearsFunded {
			s.YearsFunded = t.Year
		}
	}
	s.FinalBalance = balanceByYear[s.YearsFunded]
	return s
}
