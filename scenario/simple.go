package scenario

import (
	drawdown "github.com/vextasy/drawdown/app"
)

func NewSimpleDrawScenario() *drawdown.DrawScenario {
	s := &drawdown.DrawScenario{
		Rates: drawdown.DrawRates{},
	}

	// State Pension
	const (
		StatePensionYear0Amount       = 10000
		StatePensionStartingYear      = 1
		StatePensionAnnualPctIncrease = 2.5

		// Savings
		SavingsInitialBalance = 40000
		IsaInitialBalance     = 40000

		// Pension
		Pension1InitialBalance = 350000
		Pension2InitialBalance = 150000

		// Investments
		GiaInitialBalance = 50000
	)

	// Sources
	is_state_pension_1 := drawdown.NewStatePension("State Pension 1", StatePensionYear0Amount, StatePensionAnnualPctIncrease, 0)
	is_state_pension_2 := drawdown.NewStatePension("State Pension 2", StatePensionYear0Amount, StatePensionAnnualPctIncrease, StatePensionStartingYear)

	is_savings := drawdown.NewSavingsAccount("Savings", SavingsInitialBalance, &s.Rates.SavingsGrowthRate)
	is_isa := drawdown.NewInvestmentAccount("ISA", IsaInitialBalance, &s.Rates.InvestmentGrowthRate)

	is_pension_1 := drawdown.NewInvestmentAccount("Pension 1", 0.75*Pension1InitialBalance, &s.Rates.InvestmentGrowthRate)
	is_pension_1_tfls := drawdown.NewSavingsAccount("TFLS 1", 0.25*Pension1InitialBalance, &s.Rates.SavingsGrowthRate)

	is_pension_2 := drawdown.NewInvestmentAccount("Pension 2", 0.75*Pension2InitialBalance, &s.Rates.InvestmentGrowthRate)
	is_pension_2_tfls := drawdown.NewSavingsAccount("TFLS 2", 0.25*Pension2InitialBalance, &s.Rates.SavingsGrowthRate)

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
		is_pension_2:       incomeTaxAccount2,
		is_gia:             capitalGainsTaxAccount,
	}

	// Inflation linked variables
	incomeTaxAllowance := incomeTaxRegime.TaxFreeAllowance()
	capitalGainsTaxAllowance := capitalGainsTaxRegime.TaxFreeAllowance()
	annualMaximumIsaContribution := int64(20000)
	v1000 := int64(1000)

	allInflationLinkedVariables := []*int64{
		&annualMaximumIsaContribution,
		&incomeTaxAllowance,
		&capitalGainsTaxAllowance,
		&v1000,
	}

	// The full set of sources
	allSources := []*drawdown.Source{
		is_state_pension_1,
		is_state_pension_2,
		is_pension_1,
		is_pension_1_tfls,
		is_pension_2,
		is_pension_2_tfls,
		is_savings,
		is_isa,
		is_gia,
	}

	// The order in which to draw from the sources
	// drawdown.Seq draws upto an amount from one or more sources in order.
	// drawdown.Split draws from two sources using a percentage split.
	drawSequence := []*drawdown.Source{
		is_state_pension_1,
		is_state_pension_2,
		drawdown.Seq(&capitalGainsTaxAllowance, is_gia),
		drawdown.Seq(&v1000, is_pension_1_tfls, is_savings, is_pension_2_tfls),
		is_isa,
		is_pension_1_tfls,
		is_pension_2_tfls,
		is_pension_1,
		is_pension_2,
		is_savings,
		is_gia,
	}

	// The order in which to consider sources for the payment of tax.
	taxPaymentSequence := []*drawdown.Source{
		is_isa,
		is_pension_1_tfls,
		is_pension_2_tfls,
		is_savings,
		is_gia,
		is_pension_1,
		is_pension_2,
	}

	// Actions are performed at the start of the year
	// after the sources and tax accounts have been initialised
	// and before any withdrawals are made.
	actions := []func(year int, need int64, s *drawdown.DrawScenario){
		func(year int, need int64, s *drawdown.DrawScenario) {
			//fmt.Println("year", year, "need", need)
		},
		func(year int, need int64, s *drawdown.DrawScenario) {
			// Every year make the maximum annual ISA contribution from the savings accounts.
			drawdown.Transfer(&annualMaximumIsaContribution, is_isa, is_savings, is_pension_1_tfls, is_pension_2_tfls)
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
