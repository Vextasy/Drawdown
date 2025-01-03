package drawdown

import (
	"fmt"
	"math"
)

type DrawRates struct {
	InvestmentGrowthRate     float64
	SavingsGrowthRate        float64
	AnnualInflationRate      float64
	PlatformChargeRate       float64
	TaxBandAnnualPctIncrease float64
}

type DrawScenario struct {
	Sources                  []*Source
	DrawSequence             []*Source
	TaxPaymentSequence       []*Source
	TaxAccounts              map[*Source]*TaxAccount
	TaxRegimes               []*TaxRegime
	Actions                  []func(year int, need int64, s *DrawScenario)
	InflationLinkedVariables []*int64
	Rates                    DrawRates
}

func (s *DrawScenario) WithComponents(
	sources []*Source,
	drawSequence []*Source,
	taxPaymentSequence []*Source,
	taxAccounts map[*Source]*TaxAccount,
	taxRegimes []*TaxRegime,
	actions []func(year int, need int64, s *DrawScenario),
	inflationLinkedVariables []*int64,
) *DrawScenario {
	s.Sources = sources
	s.DrawSequence = drawSequence
	s.TaxPaymentSequence = taxPaymentSequence
	s.TaxAccounts = taxAccounts
	s.TaxRegimes = taxRegimes
	s.Actions = actions
	s.InflationLinkedVariables = inflationLinkedVariables
	return s
}

func (s *DrawScenario) WithRates(r DrawRates) *DrawScenario {
	s.Rates = r
	return s
}

// A Transaction represents the situation, for a given source, at the end of the year.
// Withdrawals from a source may cause tax to be raised.
// Tax raised may be paid by the same source, or a different source.
// The Amount of a transaction includes any amount withdrawn to pay tax.
type Transaction struct {
	Year      int
	Source    string
	Amount    int64 // The amount withdrawn from this source (including tax paid).
	Tax       int64 // (the amount of) Tax paid from this source.
	TaxRaised int64 // (the amount of) Tax raised as a result of withdrawing from this source.
	Balance   int64 // Remaining value in the source.
}

type DrawHistory []Transaction

// Iterate returns a transaction for each combination of Source and increasing Year.
func (s *DrawScenario) Iterate(years int, year1AnnualIncome int) DrawHistory {

	transactions := []Transaction{}

	var unpaidTax int64 = 0
	for year := 1; year <= years; year++ {
		var need int64 = int64(float64(year1AnnualIncome) * math.Pow(1+s.Rates.AnnualInflationRate/100, float64(year-1)))
		need += unpaidTax
		unpaidTax = 0
		//fmt.Println("year", year, "need", need)

		// Start of year.
		for _, source := range s.Sources {
			source.StartYear(year)
		}
		for _, ta := range s.TaxAccounts {
			ta.Reset(year)
		}
		// Actions
		for _, a := range s.Actions {
			a(year, need, s)
		}

		// Platform charges
		balance := int64(0)
		for _, source := range s.Sources {
			balance += source.PlatformChargeBalance()
		}
		platformCharges := int64(float64(balance) * s.Rates.PlatformChargeRate / 100)
		need += platformCharges
		//fmt.Println("year", year, "balance", balance, "charges", platformCharges)

		// Withdrawals
		withdrawn := make(map[*Source]int64) // Amount withdrawn from each source this year.
		for _, source := range s.DrawSequence {
			iss := source.Withdraw(need) // Source might split withdrawal between multiple sub-sources.
			for _, is := range iss {
				need -= is.Amount
				need = max(0, need) // Some sources, such as the State Pension, may return more than needed.
				withdrawn[is.Source] += is.Amount
				//fmt.Println("year", year, "source", is.Source.Name, "amount", is.Amount, "balance", is.Source.balance)
			}
		}
		// Tax
		taxRaised := make(map[*Source]int64) // Tax amount raised from each source this year.
		taxToPay := int64(0)
		for is, w := range withdrawn {
			ta, taxable := s.TaxAccounts[is]
			if taxable {
				tax := ta.TaxOn(w)
				taxToPay += tax
				taxRaised[is] += tax
				//fmt.Println(is.Name, "withdrawn", w, "taxable", taxable, "ta", ta.Name, "tax", tax)
			}
		}
		// Pay tax
		taxWithdrawn := make(map[*Source]int64) // Amount withdrawn from each source this year to pay tax.
		payTaxNextYear := true
		if payTaxNextYear {
			unpaidTax = taxToPay
		} else {

			for _, source := range s.TaxPaymentSequence {
				sas := source.Withdraw(taxToPay) // Source might split withdrawal between multiple sub-sources.
				for _, sa := range sas {
					taxToPay -= sa.Amount
					withdrawn[sa.Source] += sa.Amount
					taxWithdrawn[sa.Source] += sa.Amount
					// If paying tax from a taxable source, calculate the additional tax raised on that withdrawal.
					ta, taxable := s.TaxAccounts[sa.Source]
					if taxable {
						unpaidTax += ta.TaxDue(sa.Amount)
					}
				}
			}
			if taxToPay > 0 {
				fmt.Println("Some tax unpaid:", taxToPay)
			}
		}

		// End of year.
		for _, source := range s.Sources {
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
		for _, tr := range s.TaxRegimes {
			tr.ScaleOneYear(s.Rates.TaxBandAnnualPctIncrease)
		}
		for _, iv := range s.InflationLinkedVariables {
			*iv = int64(float64(*iv) * (1 + s.Rates.AnnualInflationRate/100))
		}

		if need > 0 {
			fmt.Println("Not enough funds in year ", year, " need ", need)
			break
		}
	}
	return transactions
}
