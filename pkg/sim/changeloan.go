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

// CompanyFlagAutopayLoan is CompanyFlags::autopayLoan (new for OpenLoco):
// the Finances tab's "pay back the loan automatically" checkbox.
const CompanyFlagAutopayLoan = 1 << 31

// updateLoanAutorepay is Company::updateLoanAutorepay (the controlling
// company, daily): with autopay on, the loan paid back as far as the cash
// goes — ChangeLoan with apply, as that company.
func (m *Map) updateLoanAutorepay(id int) {
	if id >= len(m.Companies) || id >= len(m.CompanyFaces) {
		return
	}
	c := &m.Companies[id]
	if c.CurrentLoan <= 0 || c.Cash <= 0 || m.CompanyFaces[id].ChallengeFlags&CompanyFlagAutopayLoan == 0 {
		return
	}
	newLoan := c.CurrentLoan - int32(max(0, min(int64(c.CurrentLoan), c.Cash)))
	cmd := &Cmd{Company: uint8(id)}
	m.DoCommandFor(cmd, CmdApply, ExpenditureLoanInterest, func(f uint8) (int32, bool) { return m.ChangeLoan(cmd, f, newLoan) })
}
