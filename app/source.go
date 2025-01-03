package drawdown

import (
	"math"
	"strings"
)

// Type SourceAmount represents an amount of money associated with a given source.
type SourceAmount struct {
	Source *Source
	Amount int64
}

// Source represents something from which income can be drawn.
// This might be a savings account, an investment account, or a pension, for example.
type Source struct {
	Name              string
	balance           int64                             // The amount of money currently in the source.
	year              int                               // The current year (origin one) - decisions might be based on this.
	hasPlatformCharge bool                              // the balance counts towards the platform charge.
	startYear         func(year int)                    // Called at the beginning of each year typically to set the opening balance (year origin is zero).
	endYear           func(year int)                    // Called at the end of each year.
	makeWithdrawal    func(amount int64) []SourceAmount // nil, else it returns the amount withdrawn from the source.
}

// setBalance sets the source's balance to a given value.
func (is *Source) setBalance(amount int64) {
	if amount < 0 {
		panic("Cannot set a negative balance")
	}
	is.balance = amount
}

func (is *Source) Balance() int64 {
	return is.balance
}

func (is *Source) PlatformChargeBalance() int64 {
	if is.hasPlatformCharge {
		return is.balance
	}
	return 0
}

func (is *Source) reduceBalance(amount int64) []SourceAmount {
	if is.balance < amount {
		amount = is.balance
		is.balance = 0
	} else {
		is.setBalance(is.balance - amount)
	}
	return []SourceAmount{{is, amount}}
}

func (is *Source) increaseBalance(amount int64) {
	is.setBalance(is.balance + amount)
}

// Withdraw removes the given amount from the source's balance and returns the amount withdrawn.
// If the source balance is less than the requested amount, the requested amount will be reduced to the balance
// and the balance will be set to 0.
// Otherwise the the balance will be reduced by the requested amount.
// The returned SourceAmount contains the actual amount withdrawn.
func (is *Source) Withdraw(amount int64) []SourceAmount {
	if amount < 0 {
		//fmt.Println("amount", amount, "account", is.Name)
		panic("Cannot withdraw negative amount")
	}
	if is.makeWithdrawal != nil {
		return is.makeWithdrawal(amount)
	}
	return is.reduceBalance(amount)
}

func (is *Source) Deposit(amount int64) {
	if amount < 0 {
		//fmt.Println("deposit", amount, "account", is.Name)
		panic("Cannot deposit negative amount")
	}
	is.increaseBalance(amount)
}

// NewYear is called at the beginning of a year, typically to set the opening balance.
// The year origin is one.
func (is *Source) StartYear(year int) {
	is.year = year
	if is.startYear == nil {
		return
	}
	is.startYear(year)
}

// EndYear is called at the end of each year.
// The year origin is one.
func (is *Source) EndYear(year int) {
	if is.endYear == nil {
		return
	}
	is.endYear(year)
}

// IsEmpty returns true if the source's balance is 0.
func (is *Source) IsEmpty() bool {
	return is.balance == 0
}

// NewStatePension creates a state pension source.
// Year1AnnualAmount is the amount paid in year 1.
// AnnualPctIncrease is the percentage increase per year. (For example, 2.0 for 2% increase per year)
// StartingYear is the year that the state pension starts.
func NewStatePension(name string, year1AnnualAmount int64, annualPctIncrease float64, startingYear int) *Source {
	is := &Source{
		Name:              name,
		hasPlatformCharge: false,
	}
	// Set a new opening balance each year which is scaled up by the annual percentage increase.
	is.startYear = func(year int) {
		var newBalance int64
		initialAnnualStatePension := year1AnnualAmount
		increasePct := annualPctIncrease / 100
		newBalance = int64(math.Pow((1+increasePct), float64(year-1)) * float64(initialAnnualStatePension))
		if is.year < startingYear {
			is.setBalance(0)
		} else {
			is.setBalance(newBalance)
		}
	}
	is.makeWithdrawal = func(amount int64) []SourceAmount {
		/*
			if is.year < startingYear {
				return []SourceAmount{{is, 0}}
			}
		*/
		return is.reduceBalance(is.balance)
	}
	return is
}

// NewSavingsAccount creates a savings account source.
// InitialBalance is the balance at the start of the first year.
// AnnualPctIncrease is the percentage increase per year. (For example, 2.0 for 2% increase per year).
func NewSavingsAccount(name string, initialBalance int64, annualPctIncrease *float64) *Source {
	is := &Source{
		Name:              name,
		hasPlatformCharge: true,
	}
	is.setBalance(initialBalance)
	// Scale the balance up each year (apart for the first) by the annual percentage increase.
	is.startYear = func(year int) {
		if year == 1 {
			return
		}
		increasePct := *annualPctIncrease / 100
		is.setBalance(int64((1 + increasePct) * float64(is.balance)))
	}
	return is
}

// NewInvestmentAccount creates a fund or share investment account source.
// InitialBalance is the balance at the start of the first year.
// AnnualPctIncrease is the percentage growth per year. (For example, 2.0 for 2% growth per year).
func NewInvestmentAccount(name string, initialBalance int64, annualPctIncrease *float64) *Source {
	is := &Source{
		Name:              name,
		hasPlatformCharge: true,
	}
	is.setBalance(initialBalance)
	// Scale the balance up each year (apart for the first) by the annual percentage increase.
	//lastOpeningBalance := initialBalance
	is.startYear = func(year int) {
		if year == 1 {
			return
		}

		// Scale the balance up each year (apart for the first) by the annual percentage increase.
		increasePct := *annualPctIncrease / 100
		is.setBalance(int64((1 + increasePct) * float64(is.balance)))
	}
	return is
}

func Upto(is *Source, upto int64) *Source {
	nis := &Source{
		Name: is.Name,
	}
	nis.makeWithdrawal = func(amount int64) []SourceAmount {
		return is.reduceBalance(min(is.balance, upto))
	}
	return nis
}

func Seq(upto *int64, iss ...*Source) *Source {
	is := &Source{
		Name: "Seq " + strings.Join(incomeSourceNames(iss), " + "),
	}
	is.makeWithdrawal = func(amount int64) []SourceAmount {
		need := min(amount, *upto)
		sources := make([]SourceAmount, 0)
		for _, is := range iss {
			got := is.Withdraw(need)
			sources = append(sources, got...)
			need -= totalSourceAmount(got)
			if need == 0 {
				break
			}
		}
		return sources
	}
	return is
}

// Transfer can be used as an action to move money between sources.
func Transfer(upto *int64, to *Source, from ...*Source) *Source {
	sources := Seq(upto, from...).Withdraw(*upto)
	got := totalSourceAmount(sources)
	to.Deposit(got)
	//fmt.Println("Transfer", got, "to", to.Name, "from", strings.Join(incomeSourceNames(from), " + "))
	return &Source{
		Name: "Transfer to " + to.Name + " from " + strings.Join(incomeSourceNames(from), " + "),
	}
}

// Return a new Source which, on withdrawal, will draw from source1 and source2 in the given percentages.
// The percentages are expressed as, for example, 2.0 for 2%.
func Split(is1 *Source, is2 *Source, pct1 int64, pct2 int64) *Source {
	is := &Source{
		Name: is1.Name + " + " + is2.Name,
	}
	is.makeWithdrawal = func(amount int64) []SourceAmount {
		a1 := is1.Withdraw(amount * pct1 / 100)
		a2 := is2.Withdraw(amount * pct2 / 100)
		return append(a1, a2...)
	}
	return is
}

func totalSourceAmount(sources []SourceAmount) int64 {
	total := int64(0)
	for _, sa := range sources {
		total += sa.Amount
	}
	return total
}

func incomeSourceNames(iss []*Source) []string {
	names := make([]string, len(iss))
	for i, is := range iss {
		names[i] = is.Name
	}
	return names
}
