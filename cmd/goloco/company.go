package main

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
)

// company.go — the Company window (CompanyWindow.cpp) and the company list
// (CompanyList.cpp): status, details, colours, finances, cargo, challenge.

// companyFace is the owner's image of a company: its competitor object's
// images[ownerEmotion], recoloured to mainColours.primary (0 if unknown).
func (g *Game) companyFace(company int) (int, uint8) {
	if g.r.ObjMgr == nil || company >= len(g.w.CompanyFaces) {
		return 0, 0
	}
	f := g.w.CompanyFaces[company]
	if int(f.Competitor) >= len(g.r.ObjMgr.CompetitorObjects) || g.r.ObjMgr.CompetitorObjects[f.Competitor] == nil || f.Emotion >= 9 {
		return 0, 0
	}
	var colour uint8
	if company < len(g.w.CompanyColours) {
		colour = g.w.CompanyColours[company]
	}
	return int(g.r.ObjMgr.CompetitorObjects[f.Competitor].Images[f.Emotion]), colour
}

// playerCompanyDropdown is PlayerInfoPanel's playerMouseDown: every company
// by performanceIndex (highest first), each its position, face, name and
// performance (dropdown_company_performance), then the companies list item;
// showBelow (25 high), the player's company highlighted. playerDropdownClick:
// a company opens its Company window (CompanyWindow::open, the Status tab).
// PARTIAL: the companies list item (CompanyList isn't ported).
func (g *Game) playerCompanyDropdown(x, y, w, h int) {
	m := &g.w.Map
	type co struct {
		id   int
		perf int16
	}
	var cs []co
	for id := range m.CompanyNames {
		if !m.CompanyEmpty(id) && id < len(m.CompanyFaces) {
			cs = append(cs, co{id, m.CompanyFaces[id].Performance})
		}
	}
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].perf > cs[j].perf })
	ctrl, _ := g.w.ControllingCompany()
	var items []ui.DropdownItem
	highlight := -1
	for i, c := range cs {
		face, col := g.companyFace(c.id)
		rating := i18n.CorporateRatingPlatelayer + min(9, max(0, int(c.perf)/100))
		perf := ui.FillArgs("{INT16_1DP}% “{STRINGID}”", strconv.Itoa(int(c.perf)/10)+"."+strconv.Itoa(int(c.perf)%10), i18n.Str(rating))
		items = append(items, ui.DropdownItem{Text: i18n.Str(i18n.Position1st + min(i, 14)), Sprite: face, SpriteColour: col,
			Lines: []string{m.CompanyName(c.id), perf}, CompanyPerformance: true})
		if c.id == ctrl {
			highlight = i
		}
	}
	listIcon := 2309                                 // ImageIds::company_list_dropdown_icon
	listText := i18n.Str(i18n.DropdownCompaniesList) // its text after the layout codes
	if k := strings.LastIndexByte(listText, '}'); k >= 0 {
		listText = listText[k+1:]
	}
	items = append(items, ui.DropdownItem{Text: listText, Sprite: listIcon, CompaniesListItem: true})
	var colour uint8 = ui.DropdownColour
	if sk := g.skin(); sk != nil {
		colour = sk.PlayerInfoToolbarColour
	}
	d := ui.NewDropdownShowText(x, y, w, h, colour, items, 25, g.sw, g.sh)
	d.SetToolbarColour(colour)
	d.Highlighted = highlight
	d.OnSelect = func(i int) {
		if i >= 0 && i < len(cs) {
			g.openCompanyWindow(cs[i].id)
		} else if i == len(cs) {
			g.openCompanyList() // the companies list item: CompanyList::open
		}
	}
	g.dropdown = d
}

// openCompanyWindow is CompanyWindow::open: the Company window on its Status
// tab.
func (g *Game) openCompanyWindow(company int) {
	if cw := g.newCompanyWindow(company); cw != nil {
		cw.SwitchTab(ui.CompanyTabStatus)
		g.showAuto(cw, 0, 0) // create: Status::kWindowSize, the tab it opens on
	}
}

// openCompanyChallenge is CompanyWindow::openChallenge: the Company window on
// its Challenge tab.
func (g *Game) openCompanyChallenge(company int) {
	if cw := g.newCompanyWindow(company); cw != nil {
		cw.SwitchTab(ui.CompanyTabChallenge)
		g.showAuto(cw, 0, 0)
	}
}

// openCompanyFinances is CompanyWindow::openFinances: the Company window on
// its Finances tab (636 × 319), the owner's colours.
func (g *Game) openCompanyFinances(company int) {
	if cw := g.newCompanyWindow(company); cw != nil {
		g.showAuto(cw, 270, 182) // create: Status::kWindowSize
	}
}

// newCompanyWindow builds a company's window (WindowType::company, numbered
// by the company) on its Finances tab; nil: no such company.
func (g *Game) newCompanyWindow(company int) *ui.CompanyFinancesWindow {
	if company < 0 || company >= len(g.w.Map.Companies) {
		return nil
	}
	tc, wc := g.windowColours(true)
	if company < len(g.w.CompanyColours) { // the company's own colour, as its windows are
		tc = g.w.CompanyColours[company]
	}
	w, h := ui.CompanyFinancesSize[0], ui.CompanyFinancesSize[1]
	cw := &ui.CompanyFinancesWindow{WinBase: ui.WinBase{WinType: ui.WindowTypeCompany, Number: company, W: w, H: h}, TitleColour: tc, Colour: wc,
		Tab:   ui.CompanyTabFinances, // openFinances: currentTab = tab_finances
		Title: g.w.Map.CompanyName(company), Data: g.companyFinancesData(company),
		OnLoan:        func(dir, step int) { g.changeLoan(company, dir, step) },
		OnRename:      func() { g.renameCompany(company, tc, wc) },
		OnRenameOwner: func() { g.renameOwner(company, tc, wc) },
		OnColour: func(colourType int, primary bool, x, y, w, h int) {
			g.colourSwatchOpen(company, colourType, primary, x, y, h)
		},
		OnColourMode: func(colourType int, on bool) { // ColourScheme onMouseUp check_…
			g.changeCompanyColour(company, sim.ColourSchemeChange{Company: uint8(company), ColourType: uint8(colourType), Value: b2u8(on), SetMode: true})
		},
		OnAutopay: func() { // onMouseUp loan_autopay: challengeFlags ^= autopayLoan
			if company < len(g.w.Map.CompanyFaces) {
				g.w.Map.CompanyFaces[company].ChallengeFlags ^= sim.CompanyFlagAutopayLoan
			}
		}}
	cw.ScrollToLatest() // openFinances: scrollToLatestData
	cw.Face, cw.FaceColour = g.companyFace(company)
	cw.CompanyID, cw.ScreenW, cw.ScreenH = company, g.sw, g.sh
	for id := range g.w.CompanyNames { // populateCompanySelect: every non-empty slot
		if g.w.CompanyNames[id] == 0 || id >= len(g.w.CompanyFaces) {
			continue
		}
		face, col := g.companyFace(id)
		cw.Companies = append(cw.Companies, ui.CompanyChoice{ID: id, Name: g.w.CompanyName(id), Performance: int(g.w.CompanyFaces[id].Performance), Face: face, Colour: col})
	}
	cw.OpenDropdown = func(d *ui.DropdownMenu) { g.dropdown = d }
	cw.OnCompany = func(id int) { // switchCompany: the window turns to that company, on its place and tab
		if id == company {
			return
		}
		if open := g.companyWin(id); open != nil { // that company's window is open: it comes to the front
			g.wins.BringToFront(open)
			return
		}
		if n := g.newCompanyWindow(id); n != nil {
			n.SwitchTab(cw.Tab)
			g.wins.Replace(cw, n)
		}
	}
	return cw
}

// changeLoan is Finances' loan_decrease / loan_increase: the loan ∓ £1,000 ×
// getClickRepeatStepSize (1 for a click, growing while held), as ChangeLoan with apply; a refusal opens the error window
// (cant_pay_back_loan / cant_borrow_any_more_money and the command's reason).
func (g *Game) changeLoan(company, dir, step int) {
	m := &g.w.Map
	if company < 0 || company >= len(m.Companies) {
		return
	}
	cur := m.Companies[company].CurrentLoan
	bound := func(v int64) int32 { return int32(min(max(v, math.MinInt32), math.MaxInt32)) } // Math::Bound
	newLoan, title := bound(int64(cur)+int64(step)*1000), i18n.CantBorrowAnyMoreMoney
	if dir < 0 {
		if cur == 0 {
			return
		}
		newLoan, title = max(bound(int64(cur)-int64(step)*1000), 0), i18n.CantPayBackLoan
	}
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureLoanInterest, func(f uint8) (int32, bool) { return m.ChangeLoan(c, f, newLoan) }); !ok {
		g.openError(i18n.Label(title), cmdErrorText(c))
	}
}

// companyFinancesData is what Finances::draw reads of a company: its
// expenditure columns (numExpenditureYears, oldest first, the current year
// last), the loan and interest rate, cash (bankrupt), the latest company
// value and the vehicles' profit.
func (g *Game) companyFinancesData(company int) ui.CompanyFinances {
	m := &g.w.Map
	if company < 0 || company >= len(m.Companies) {
		return ui.CompanyFinances{}
	}
	c := &m.Companies[company]
	cur := int(m.CurrentYear)
	years := min(int(c.NumExpenditureYears), len(c.Expenditures))
	f := ui.CompanyFinances{CurrentYear: cur, NumYears: int(c.NumExpenditureYears), Loan: int64(c.CurrentLoan), InterestRate: int(m.LoanInterestRate),
		Cash: c.Cash, Value: c.CompanyValueHistory[0], Profit: c.VehicleProfit}
	ctrl, _ := g.w.ControllingCompany()
	f.Own = company == ctrl
	if company < len(m.CompanyFaces) {
		f.Bankrupt = m.CompanyFaces[company].ChallengeFlags&(1<<9) != 0 // CompanyFlags::bankrupt
		f.Autopay = m.CompanyFaces[company].ChallengeFlags&sim.CompanyFlagAutopayLoan != 0
	}
	for i := 0; i < years; i++ {
		f.Years = append(f.Years, cur-(years-i)+1)
		f.Expenditures = append(f.Expenditures, c.Expenditures[years-i-1])
	}
	return f
}

// companyCargoLines is CargoDelivered::draw's list: for each cargo object
// the company has delivered, its unit name (singular for one) with the count.
func (g *Game) companyCargoLines(company int) []string {
	m := &g.w.Map
	if company < 0 || company >= len(m.Companies) || g.r.ObjMgr == nil {
		return nil
	}
	var out []string
	for i, n := range m.Companies[company].CargoDelivered {
		if n == 0 || i >= len(g.r.ObjMgr.CargoObjects) || g.r.ObjMgr.CargoObjects[i] == nil {
			continue
		}
		c := g.r.ObjMgr.CargoObjects[i]
		unit := c.UnitNamePlural
		if n == 1 {
			unit = c.UnitNameSingular
		}
		out = append(out, ui.FillArgs(ui.ObjectString(unit), ui.Comma32(int64(n))))
	}
	return out
}

// companyChallenge is what Challenge::draw reads: the scenario's details, the
// objective (formatChallengeArguments) and the player company's progress.
func (g *Game) companyChallenge(company int) ui.CompanyChallenge {
	m := &g.w.Map
	o := m.Objective
	co := ui.ChallengeObjective{Type: o.Type, Flags: o.Flags, CompanyValue: o.CompanyValue, MonthlyProfit: o.MonthlyVehicleProfit,
		PerformanceIndex: o.PerformanceIndex, CargoAmount: o.DeliveredCargoAmount, TimeLimitYears: o.TimeLimitYears,
		UntilYear: int(m.ObjectiveUntilYear), TitleOrEditor: g.inTitleScreen}
	if om := g.r.ObjMgr; om != nil && int(o.DeliveredCargoType) < len(om.CargoObjects) && om.CargoObjects[o.DeliveredCargoType] != nil {
		co.CargoUnitPlural = ui.ObjectString(om.CargoObjects[o.DeliveredCargoType].UnitNamePlural)
	}
	c := ui.CompanyChallenge{Details: g.scenarioDetails, Objective: ui.ChallengeObjectiveText(co)}
	if company >= 0 && company < len(m.Companies) && company < len(m.CompanyFaces) {
		c.Progress = int(m.Companies[company].ChallengeProgress)
		flags := m.CompanyFaces[company].ChallengeFlags
		c.Completed, c.Failed = flags&sim.CompanyFlagChallengeCompleted != 0, flags&sim.CompanyFlagChallengeFailed != 0
	}
	c.CompletedMonths = int(m.ObjectiveCompletedInMonths)
	c.TimeLimited = o.Flags&4 != 0
	c.MonthsLeft = int(o.TimeLimitYears)*12 - int(m.ObjectiveMonthsInChallenge)
	return c
}

// companyStatus is what Status::draw reads: the owner's face (the bigger
// image, competitor images[ownerEmotion] + 1), name and status
// (getOwnerStatus: bankrupt, else — the player's company having no
// observation — company_status_empty; PARTIAL: an AI's observation lines).
func (g *Game) companyStatus(company int) ui.CompanyStatus {
	face, col := g.companyFace(company)
	if face > 0 {
		face++
	}
	st := ui.CompanyStatus{Face: face, FaceColour: col, OwnerName: g.w.Map.OwnerName(company, g.r.ObjMgr), Status: i18n.Str(i18n.CompanyStatusEmpty)}
	if company < len(g.w.Map.CompanyFaces) && g.w.Map.CompanyFaces[company].ChallengeFlags&(1<<9) != 0 {
		st.Status = i18n.Str(i18n.CompanyStatusBankrupt)
	}
	ctrl, _ := g.w.ControllingCompany()
	st.Own = company == ctrl
	return st
}

// companyDetails is what Details::draw reads of a company.
func (g *Game) companyDetails(company int) ui.CompanyDetails {
	m := &g.w.Map
	if company < 0 || company >= len(m.Companies) || company >= len(m.CompanyFaces) {
		return ui.CompanyDetails{}
	}
	c, f := &m.Companies[company], &m.CompanyFaces[company]
	y, mo, d := uilayout.CalcDate(c.StartedDate)
	det := ui.CompanyDetails{Started: uilayout.DateDMY(d, i18n.Month(mo+1), y), Performance: f.Performance,
		OwnerName: m.OwnerName(company, g.r.ObjMgr), Counts: c.TransportTypeCount, Headquarters: c.HeadquartersX != -1}
	switch {
	case f.ChallengeFlags&(1<<5) != 0: // decreasedPerformance
		det.Trend = -1
	case f.ChallengeFlags&(1<<4) != 0: // increasedPerformance
		det.Trend = 1
	}
	return det
}

// changeCompanyColour is a ChangeCompanyColourScheme with apply (its error
// title error_cant_change_colour_scheme on a refusal).
func (g *Game) changeCompanyColour(company int, a sim.ColourSchemeChange) {
	m := &g.w.Map
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureMiscellaneous, func(f uint8) (int32, bool) { return m.ChangeCompanyColourScheme(c, f, a, g.r.ObjMgr) }); !ok {
		g.openError(i18n.Label(i18n.ErrorCantChangeColourScheme), cmdErrorText(c))
	}
}

func b2u8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// companyColours is what ColourScheme::prepareDraw reads of a company.
func (g *Game) companyColours(company int) ui.CompanyColours {
	m := &g.w.Map
	if company < 0 || company >= len(m.CompanyColourSchemes) {
		return ui.CompanyColours{}
	}
	cs := &m.CompanyColourSchemes[company]
	ctrl, _ := g.w.ControllingCompany()
	return ui.CompanyColours{Main: cs.Main, Vehicle: cs.Vehicle, Custom: cs.CustomSet, Own: company == ctrl}
}

// colourSwatchOpen is ColourScheme's onMouseDown on a colour swatch:
// showColour with the colour now set highlighted — every colour for a
// secondary, all but the competing companies' (competingColourMask) for a
// primary; onDropdown: ChangeCompanyColourScheme with the colour picked.
func (g *Game) colourSwatchOpen(company, colourType int, primary bool, x, y, h int) {
	m := &g.w.Map
	if company < 0 || company >= len(m.CompanyColourSchemes) {
		return
	}
	cs := &m.CompanyColourSchemes[company]
	pair := cs.Main
	if colourType > 0 {
		pair = cs.Vehicle[colourType-1]
	}
	selected := pair[1]
	available := uint32(0x7FFFFFFF)
	if primary {
		selected = pair[0]
		available &^= m.CompetingColourMask(company)
	}
	var colour uint8 = 1
	if w := g.companyWin(company); w != nil {
		colour = w.Colour
	}
	d := ui.NewColourDropdown(x, y, h, available, selected, colour, g.sw, g.sh)
	d.OnSelect = func(c uint8) {
		g.changeCompanyColour(company, sim.ColourSchemeChange{Company: uint8(company), ColourType: uint8(colourType), Value: c, IsPrimary: primary})
	}
	g.colourDropdown = d
}

// renameCompany is renameCompanyPrompt and renameCompany: the text input
// (title_name_company, prompt_enter_new_company_name, the current name), then
// — a non-empty name — RenameCompany with apply (cannot_rename_this_company
// and the reason on a refusal).
func (g *Game) renameCompany(company int, titleColour, colour uint8) {
	m := &g.w.Map
	g.textInput = ui.OpenTextInput(i18n.Str(i18n.TitleNameCompany), i18n.Str(i18n.PromptEnterNewCompanyName),
		m.CompanyName(company), 31, g.sw, g.sh, titleColour, colour, true, func(text string) {
			if text == "" {
				return
			}
			c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
			if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureMiscellaneous, func(f uint8) (int32, bool) { return m.RenameCompany(c, f, company, text) }); !ok {
				g.openError(i18n.Label(i18n.CannotRenameThisCompany), cmdErrorText(c))
				return
			}
			if w := g.companyWin(company); w != nil {
				w.Title = m.CompanyName(company)
			}
		})
}

// openCompanyList is CompanyList::open: the list on its first tab, 640 × 272
// the skin's title bar and
// window colours, sorted by performance (sortMode 2); a row opens that
// company's window.
func (g *Game) openCompanyList() {
	m := &g.w.Map
	tc, wc := g.windowColours(false)
	var rows []ui.CompanyListRow
	for id := range m.CompanyNames {
		if m.CompanyEmpty(id) || id >= len(m.CompanyFaces) || id >= len(m.Companies) {
			continue
		}
		face, col := g.companyFace(id)
		f := m.CompanyFaces[id]
		r := ui.CompanyListRow{ID: id, Name: m.CompanyName(id), Face: face, FaceColour: col, Status: g.companyStatus(id).Status,
			Performance: f.Performance, Value: m.Companies[id].CompanyValueHistory[0],
			PerfHist: m.Companies[id].PerformanceIndexHist, HistorySize: int(m.Companies[id].HistorySize),
			CargoHist: m.Companies[id].CargoUnitsDeliveredHist, DistHist: m.Companies[id].CargoUnitsDistanceHist,
			ValueHist: m.Companies[id].CompanyValueHistory}
		if id < len(g.w.CompanyColours) {
			r.Colour = g.w.CompanyColours[id]
		}
		if f.ChallengeFlags&(1<<4) != 0 && f.ChallengeFlags&(1<<5) != 0 { // upstream's test: increased and decreased both
			r.Trend = 1
		}
		rows = append(rows, r)
	}
	w, h := ui.CompanyListSize[0], ui.CompanyListSize[1]
	cl := &ui.CompanyListWindow{WinBase: ui.WinBase{WinType: ui.WindowTypeCompanyList, W: w, H: h}, TitleColour: tc, Colour: wc, Rows: rows, SortMode: 2, Hover: -1,
		OnRow: g.openCompanyWindow}
	for i := range cl.Records { // CompanyManager::getRecords()
		rec := m.CompanyRecords
		r := ui.SpeedRecord{Speed: int(rec.Speed[i])}
		if c := int(rec.Company[i]); c != 0xFF && c < len(m.CompanyNames) && !m.CompanyEmpty(c) {
			y, mo, d := uilayout.CalcDate(rec.Date[i])
			r.Held, r.Name, r.Date = true, m.CompanyName(c), uilayout.DateDMY(d, i18n.Month(mo+1), y)
			r.Face, r.FaceColour = g.companyFace(c)
		}
		cl.Records[i] = r
	}
	if om := g.r.ObjMgr; om != nil { // buildDeliveredCargoPaymentsTable: 100 units, 10 blocks, 2 … 120 days
		for i, c := range om.CargoObjects {
			if c == nil || i >= len(ui.CargoLineColour) {
				continue
			}
			lc := ui.CompanyListCargo{Slot: i, Name: c.Name, Colour: ui.CargoLineColour[i]}
			for d := uint16(2); d < 122; d += 2 {
				lc.Payments[d/2-1] = m.DeliveredCargoPayment(c, 100, 10, d)
			}
			cl.Cargos = append(cl.Cargos, lc)
		}
	}
	cl.Sort()
	g.showAuto(cl, 0, 0)
}

// renameOwner is Status's change_owner_name and renameCompanyOwnerName: the
// text input (title_name_owner, prompt_enter_new_name_for_owner, the owner's
// name), then — a non-empty name — RenameCompanyOwner with apply
// (cannot_change_owner_name and the reason on a refusal).
func (g *Game) renameOwner(company int, titleColour, colour uint8) {
	m := &g.w.Map
	g.textInput = ui.OpenTextInput(i18n.Str(i18n.TitleNameOwner), i18n.Str(i18n.PromptEnterNewNameForOwner),
		m.OwnerName(company, g.r.ObjMgr), 31, g.sw, g.sh, titleColour, colour, true, func(text string) {
			if text == "" {
				return
			}
			c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
			if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureMiscellaneous, func(f uint8) (int32, bool) {
				return m.RenameCompanyOwner(c, f, company, text, g.r.ObjMgr)
			}); !ok {
				g.openError(i18n.Label(i18n.CannotChangeOwnerName), cmdErrorText(c))
			}
		})
}
