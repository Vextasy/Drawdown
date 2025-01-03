package scenario

import (
	drawdown "github.com/vextasy/drawdown/app"
)

func NewIvyDrawScenario() *drawdown.DrawScenario {
	s := &drawdown.DrawScenario{
		Rates: drawdown.DrawRates{},
	}

	// State Pension
	const (
		StatePensionYear0Amount       = 10000
		StatePensionStartingYear      = 4
		StatePensionAnnualPctIncrease = 2.5

		// Savings
		SavingsInitialBalance = 40000

		// Pension
		Pension1InitialBalance = 500000

		// Investments
		GiaInitialBalance = 50000
	)

	// Sources
	is_state_pension_1 := drawdown.NewStatePension("State Pension 1", StatePensionYear0Amount, StatePensionAnnualPctIncrease, 0)
	is_state_pension_2 := drawdown.NewStatePension("State Pension 2", StatePensionYear0Amount, StatePensionAnnualPctIncrease, StatePensionStartingYear)

	is_savings := drawdown.NewSavingsAccount("Savings", SavingsInitialBalance, &s.Rates.SavingsGrowthRate)

	is_pension_1 := drawdown.NewInvestmentAccount("Pension 1", Pension1InitialBalance, &s.Rates.InvestmentGrowthRate)

	is_gia := drawdown.NewInvestmentAccount("GIA", GiaInitialBalance, &s.Rates.InvestmentGrowthRate)

	// Tax Regimes
	incomeTaxRegime := drawdown.NewTaxRegime([]drawdown.RateBound{
		drawdown.NewRateBound(12540, 0.0),
		drawdown.NewRateBound(50270, 20.0),
		drawdown.NewRateBound(125140, 40.0),
		drawdown.NewRateBound(drawdown.HighUpperBound, 45.0),
	})
	capitalGainsTaxRegime := drawdown.NewTaxRegime([]drawdown.RateBound{
		drawdown.NewRateBound(3000, 0.0),
		drawdown.NewRateBound(drawdown.HighUpperBound, 18.0),
	})
	taxRegimes := []*drawdown.TaxRegime{&incomeTaxRegime, &capitalGainsTaxRegime}

	// Tax Accounts
	incomeTaxAccount1 := drawdown.NewTaxAccount("Income Tax 1", incomeTaxRegime)
	incomeTaxAccount2 := drawdown.NewTaxAccount("Income Tax 2", incomeTaxRegime)
	capitalGainsTaxAccount := drawdown.NewTaxAccount("Capital Gains Tax 1", capitalGainsTaxRegime)
	taxAccounts := map[*drawdown.Source]*drawdown.TaxAccount{
		is_state_pension_1: incomeTaxAccount1,
		is_state_pension_2: incomeTaxAccount2,
		is_pension_1:       incomeTaxAccount1,
		is_gia:             capitalGainsTaxAccount,
	}

	// Inflation linked variables

	allInflationLinkedVariables := []*int64{}

	// The full set of sources
	allSources := []*drawdown.Source{
		is_state_pension_1,
		is_state_pension_2,
		is_pension_1,
		is_savings,
		is_gia,
	}

	// The order in which to draw from the sources
	// drawdown.Seq draws upto an amount from one or more sources in order.
	// drawdown.Split draws from two sources using a percentage split.
	drawSequence := []*drawdown.Source{
		is_state_pension_1,
		is_state_pension_2,
		is_savings,
		is_pension_1,
		is_gia,
	}

	// The order in which to consider sources for the payment of tax.
	taxPaymentSequence := []*drawdown.Source{
		is_savings,
		is_pension_1,
		is_gia,
	}

	// Actions are performed at the start of the year
	// after the sources and tax accounts have been initialised
	// and before any withdrawals are made.
	actions := []func(year int, need int64, s *drawdown.DrawScenario){
		func(year int, need int64, s *drawdown.DrawScenario) {
			//fmt.Println("year", year, "need", need)
		},
	}

	return s.WithComponents(
		allSources,
		drawSequence,
		taxPaymentSequence,
		taxAccounts,
		taxRegimes,
		actions,
		allInflationLinkedVariables,
	)
}
