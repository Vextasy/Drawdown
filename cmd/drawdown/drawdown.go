package main

import (
	"flag"
	"fmt"
	"os"

	drawdown "github.com/vextasy/drawdown/app"
	"github.com/vextasy/drawdown/scenario"
)

const (
	Years             = 30
	Year0AnnualIncome = 35000

	InvestmentGrowthRate     = 3.5  // %
	SavingsGrowthRate        = 3.5  // %
	AnnualInflationRate      = 2.5  // %
	PlatformChargeRate       = 0.25 // The % charge for using a platform as a percentage of the balance.
	TaxBandAnnualPctIncrease = 0.5  // %
)

func main() {
	summary := flag.Bool("s", false, "produce a summary")
	flag.Parse()
	if *summary {
		doSummary()
	} else {
		doDrawdown()
	}
}
func doDrawdown() {

	s := scenario.NewSimpleDrawScenario().WithRates(drawdown.DrawRates{
		InvestmentGrowthRate:     InvestmentGrowthRate,
		SavingsGrowthRate:        SavingsGrowthRate,
		AnnualInflationRate:      AnnualInflationRate,
		PlatformChargeRate:       PlatformChargeRate,
		TaxBandAnnualPctIncrease: TaxBandAnnualPctIncrease,
	})

	transactions := s.Iterate(Years, Year0AnnualIncome)

	if len(transactions) > 0 {

		file, err := os.Create("drawdown.csv")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		fmt.Fprintf(file, "Year,Source,Amount,Tax,Tax Raised,Balance\n")
		for _, t := range transactions {
			fmt.Fprintf(file, "%d,\"%s\",%v,%v,%v,%v\n", t.Year, t.Source, t.Amount, t.Tax, t.TaxRaised, t.Balance)
		}
	}

}

func doSummary() {
	file, err := os.Create("summary.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fmt.Fprintf(file, "Investment Growth Rate,Savings Growth Rate,Annual Inflation Rate,Platform Charge Rate,Tax Band Annual Percentage Increase,Total Withdrawn,Tax Paid,Final Balance,Final Year\n")

	for _, igr := range []float64{0.0, 0.5, 1.0, 2.0, 3.0, 4.0, 5.0, 8.0} { // Investment Growth Rate
		for _, sgr := range []float64{0.0, 0.5, 1.0, 2.0, 3.0, 4.0, 5.0, 8.0} { // Savings Growth Rate
			for _, air := range []float64{2.0, 2.5, 3, 4, 5} { // Annual Inflation Rate
				for _, pcr := range []float64{0.1, 0.25, 0.5} { // Platform Charge Rate
					for _, tbi := range []float64{0.0, 0.5, 1.0, 2.0} { // Tax Band Annual Percentage Increase
						s := scenario.NewSimpleDrawScenario().WithRates(drawdown.DrawRates{
							InvestmentGrowthRate:     igr,
							SavingsGrowthRate:        sgr,
							AnnualInflationRate:      air,
							PlatformChargeRate:       pcr,
							TaxBandAnnualPctIncrease: tbi,
						})
						transactions := s.Iterate(Years, Year0AnnualIncome)
						summary := transactions.Summary()
						fmt.Fprintf(file, "igr_%.2f, sgr_%.2f, air_%.2f, pcr_%.2f, tbi_%.2f, %d, %d, %d, %d\n", igr, sgr, air, pcr, tbi, summary.TotalWithdrawn, summary.TotalTaxPaid, summary.FinalBalance, summary.FinalYear)
					}
				}
			}
		}
	}

}
