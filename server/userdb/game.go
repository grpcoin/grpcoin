package userdb

func setupGamePortfolio(u *User) {
	u.CashUSD = Amount{Units: 100_000}
	u.Positions = map[string]Amount{
		"BTC": {Units: 0, Nanos: 0},
	}
}
