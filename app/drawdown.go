package drawdown

import (
	"fmt"
	"math"
)

type Parameters struct {
	Years                    int
	Year0AnnualIncome        int
	PlatformChargesOnBalance float64
	AnnualInflationRate      float64
	TaxBandAnnualPctIncrease float64
	Sources                  []*Source
	DrawSequence             []*Source
	TaxPaymentSequence       []*Source
	TaxAccounts              map[*Source]*TaxAccount
	TaxRegimes               []*TaxRegime
	Actions                  []func(year int, need int64)
	InflationLinkedVariables []*int64
}

type Transaction struct {
	Year      int
	Source    string
	Amount    int64 // inc tax
	Tax       int64
	TaxRaised int64
	Balance   int64
}

func Iterate(p Parameters) []Transaction {

	transactions := []Transaction{}

	var unpaidTax int64 = 0
	for year := 0; year < p.Years; year++ {
		var need int64 = int64(float64(p.Year0AnnualIncome) * math.Pow(1+p.AnnualInflationRate/100, float64(year)))
		need += unpaidTax
		unpaidTax = 0
		//fmt.Println("year", year, "need", need)

		// Start of year.
		for _, source := range p.Sources {
			source.StartYear(year)
		}
		for _, ta := range p.TaxAccounts {
			ta.Reset(year)
		}
		// Actions
		for _, a := range p.Actions {
			a(year, need)
		}

		// Platform charges
		balance := int64(0)
		for _, source := range p.Sources {
			balance += source.Balance()
		}
		platformCharges := int64(float64(balance) * p.PlatformChargesOnBalance / 100)
		need += platformCharges
		//fmt.Println("year", year, "balance", balance, "charges", platformCharges)

		// Withdrawals
		withdrawn := make(map[*Source]int64) // Amount withdrawn from each source this year.
		for _, source := range p.DrawSequence {
			iss := source.Withdraw(need) // Source might split withdrawal between multiple sub-sources.
			for _, is := range iss {
				need -= is.Amount
				need = max(0, need) // Some sources, such as the State Pension, may return more than needed.
				withdrawn[is.Source] += is.Amount
			}
		}
		// Tax
		taxRaised := make(map[*Source]int64) // Tax amount raised from each source this year.
		taxToPay := int64(0)
		for is, w := range withdrawn {
			ta, taxable := p.TaxAccounts[is]
			if taxable {
				tax := ta.TaxOn(w)
				taxToPay += tax
				taxRaised[is] += tax
				//fmt.Println(is.Name, "withdrawn", w, "taxable", taxable, "ta", ta.Name, "tax", tax)
			}
		}
		// Pay tax
		taxWithdrawn := make(map[*Source]int64) // Amount withdrawn from each source this year to pay tax.
		for _, source := range p.TaxPaymentSequence {
			sas := source.Withdraw(taxToPay) // Source might split withdrawal between multiple sub-sources.
			for _, sa := range sas {
				taxToPay -= sa.Amount
				withdrawn[sa.Source] += sa.Amount
				taxWithdrawn[sa.Source] += sa.Amount
				// If paying tax from a taxable source, calculate the additional tax raised on that withdrawal.
				ta, taxable := p.TaxAccounts[sa.Source]
				if taxable {
					unpaidTax += ta.TaxDue(sa.Amount)
				}
			}
		}
		if taxToPay > 0 {
			panic("some tax unpaid")
		}

		// End of year.
		for _, source := range p.Sources {
			t := Transaction{
				Year:      year,
				Source:    source.Name,
				Amount:    withdrawn[source],    // inc tax
				Tax:       taxWithdrawn[source], // tax
				TaxRaised: taxRaised[source],    // tax raised
				Balance:   source.Balance(),
			}
			transactions = append(transactions, t)
			source.EndYear(year)
		}
		for _, tr := range p.TaxRegimes {
			tr.ScaleOneYear(p.TaxBandAnnualPctIncrease)
		}
		for _, iv := range p.InflationLinkedVariables {
			*iv = int64(float64(*iv) * (1 + p.AnnualInflationRate/100))
		}

		if need > 0 {
			panic(fmt.Sprint("Not enough funds in year ", year, " need ", need))
		}
	}
	return transactions
}
