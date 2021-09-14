package tools

import "testing"

func TestMoney(t *testing.T) {
	stringAndMoney := map[string]MoneyT{
		"1.23":    MoneyT(123),
		"0.01":    MoneyT(1),
		"1122.33": MoneyT(112233),
		"0.11":    MoneyT(11),
		"0.00":    MoneyT(0),
		"-13.33":  MoneyT(-1333),
		"-0.01":   MoneyT(-1),
	}
	mStringAndMoney := map[string]MoneyT{
		"1.233":       MoneyT(123),
		"0.01001":     MoneyT(1),
		"1122.3":      MoneyT(112230),
		"":            MoneyT(0),
		"0":           MoneyT(0),
		"-13.3321412": MoneyT(-1333),
		"-1":          MoneyT(-100),
	}

	for str, money := range stringAndMoney {
		if money.String() != str {
			t.Error("money to string boom: " + str + " -- " + money.String())
		}
		rmoney, err := StringToMoneyT(str)
		if err != nil || rmoney != money {
			t.Error("string to money boom: " + str)
		}
	}
	for str, money := range mStringAndMoney {
		rmoney, err := StringToMoneyT(str)
		if err != nil || rmoney != money {
			t.Error("string to money boom: " + str)
		}
	}
}
