package userdb

func setupGamePortfolio(u *User) {
	u.Portfolio = Portfolio{CashUSD: Amount{Units: 100_000},
		Positions: map[string]Amount{
			"BTC": {Units: 0, Nanos: 0},
		}}
}
