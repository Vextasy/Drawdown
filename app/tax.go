package drawdown

import (
	"math"
)

// A TaxAccount records the amount of a currency on which tax has already been paid
// and also the total amount of tax due on that amount.
type TaxAccount struct {
	Name        string
	regime      TaxRegime
	taxedamount int64 // The amount of currency on which tax has already been paid.
	tax         int64 // The total amount of tax that has been paid on that taxed amount.
}

func NewTaxAccount(name string, taxRegime TaxRegime) *TaxAccount {
	return &TaxAccount{
		Name:        name,
		regime:      taxRegime,
		taxedamount: 0,
		tax:         0,
	}
}

func (ta *TaxAccount) Reset(year int) {
	ta.taxedamount = 0
	ta.tax = 0
}

// TaxOn calculates the tax due on the given amount and records both the taxed amount and the tax due in the tax account.
// TaxOn returns the amount of tax due on the amount, given what has already been taxed in the tax account.
func (ta *TaxAccount) TaxOn(amount int64) int64 {
	newTax := ta.regime.TaxDue(amount, ta.taxedamount)
	ta.taxedamount += amount
	ta.tax += newTax
	return newTax
}

// TaxDue calculates the tax that is due on the withdrawal of the given amount.
// Unlike TaxOn this does not assume that the tax has been paid.
func (ta *TaxAccount) TaxDue(amount int64) int64 {
	return ta.regime.TaxDue(amount, ta.taxedamount)
}

// TaxRegime describes the rates of tax that are charged on increasing amounts.
// Commonly the first rateBound in the slice might represents a tax-free allowance.
type TaxRegime struct {
	Rates []RateBound
}

func NewTaxRegime(rates []RateBound) TaxRegime {
	return TaxRegime{Rates: rates}
}

func (tr TaxRegime) ScaleOneYear(annualPctIncrease float64) {
	for i := range tr.Rates {
		u := tr.Rates[i].upper
		if u == HighUpperBound {
			continue
		}
		u = int64(float64(u) * (1 + annualPctIncrease/100))
		tr.Rates[i].upper = u
	}
}

func (tr TaxRegime) TaxFreeAllowance() int64 {
	if len(tr.Rates) == 0 || tr.Rates[0].rate != 0 {
		return 0
	}
	return tr.Rates[0].upper
}

// RateBound contains a rate and an upper bound on the amount for which the rate applies.
// Normally a slice of RateBound is used to describe a tax regime.
// In such a slice, subsequent upper values must be strictly increasing.
// Rates are represented as a decimal percentage. For example, 10.1 for 10.1%.
const HighUpperBound = math.MaxInt64

type RateBound struct {
	upper int64
	rate  float64
}

func NewRateBound(upper int64, rate float64) RateBound {
	return RateBound{upper: upper, rate: rate}
}

// TaxDue returns the amount of tax due for the additional amount 'a'
// over an above the amount 'already' (which is assumed to have already been taxed).
func (tr TaxRegime) TaxDue(a int64, already int64) int64 {
	return int64(tr.taxDue(a+already) - tr.taxDue(already))
}

// taxDue returns the amount of tax due on an amount in the given tax regime.
func (tr TaxRegime) taxDue(a int64) int64 {
	remaining := a
	due := int64(0)
	lastUpper := int64(0)
	for _, rb := range tr.Rates {
		if remaining == 0 {
			break
		}
		taxable := min(remaining, rb.upper-lastUpper)
		remaining -= taxable
		due += int64(float64(taxable) * rb.rate / 100)
		lastUpper = rb.upper
	}
	if remaining > 0 {
		panic("tax regime failed - tax remaining")
	}
	return due
}
