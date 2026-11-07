package sim

import "github.com/LaPingvino/goloco/pkg/i18n"

// ExpenditureLoanInterest is ExpenditureType::LoanInterest.
const ExpenditureLoanInterest = 15

// ChangeLoan is GameCommands::changeLoan (ChangeLoan.cpp 0x0046DE88): the
// updating company's loan set to newLoan — paying back needs the cash for
// the difference (not_enough_cash_available), borrowing stops at the
// inflation-adjusted maximum loan (bank_refuses_to_increase_loan); applied,
// the cash moves by the difference. It costs nothing.
func (m *Map) ChangeLoan(c *Cmd, flags uint8, newLoan int32) (int32, bool) {
	if int(c.Company) >= len(m.Companies) || newLoan < 0 {
		return 0, false
	}
	co := &m.Companies[c.Company]
	diff := co.CurrentLoan - newLoan
	if co.CurrentLoan > newLoan {
		if co.Cash < int64(diff) {
			c.ErrorText = i18n.NotEnoughCashAvailable
			return 0, false
		}
	} else if maxLoan := m.InflationAdjustedCost(int16(m.MaxLoanSize), 0, 8); newLoan > maxLoan {
		c.ErrorText = i18n.BankRefusesToIncreaseLoan
		return 0, false
	}
	if flags&CmdApply != 0 {
		co.CurrentLoan = newLoan
		co.Cash -= int64(diff)
	}
	return 0, true
}
