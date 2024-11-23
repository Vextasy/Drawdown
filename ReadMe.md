Drawdown is a Go command line program that runs an iteration of a pension drawdown strategy to discover how it will unfold over time given a set of sources (pensions, investments, savings accounts) and growth rates (assumed rates of savings and investment growth and inflation) and a period of years over which drawdown will take place. The output of the program shows the balance remaining, the amount withdrawn, and the amount of tax paid at the end of each of the years. 

When run with the "-s" command line flag the program will, instead, run the same strategy over each combination of several values of the growth rates to see the impact on the final balance, the amount withdrawn and the amount of tax paid.

*Usage*
```sh
./drawdown [-s]
```

Output is in CSV format to drawdown.csv and summary.csv.