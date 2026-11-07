package ui

import "sort"

// CompanyChoice is one company for populateCompanySelect: id, name,
// performanceIndex, the owner's face and the company's primary colour.
type CompanyChoice struct {
	ID          int
	Name        string
	Performance int
	Face        int
	Colour      uint8
}

// companySelect is Dropdown::populateCompanySelect for the company select
// button (width − 28, 14, 26×26): the companies by performanceIndex (highest
// first; equal ones in slot order, as its repeated strict-max scan picks the
// first), each "{face} {name}" (dropdown_company_select), items 25 high,
// showText at the widget in the primary colour made translucent; the
// window's owner highlighted.
func (w *TownListWindow) companySelect() *DropdownMenu {
	return CompanySelectDropdown(w.X+w.W-28, w.Y+14, w.TitleColour, w.Companies, w.OwnerID, w.ScreenW, w.ScreenH, w.OnCompany)
}

// CompanySelectDropdown is populateCompanySelect for a company select
// button at (x, y) in a window of title colour tc.
func CompanySelectDropdown(x, y int, tc uint8, companies []CompanyChoice, owner, sw, sh int, onCompany func(id int)) *DropdownMenu {
	cs := append([]CompanyChoice(nil), companies...)
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].Performance > cs[j].Performance })
	items := make([]DropdownItem, len(cs))
	for i, c := range cs {
		items[i] = DropdownItem{Text: c.Name, Sprite: c.Face, SpriteColour: c.Colour, CompanySelect: true}
	}
	if sw <= 0 || sh <= 0 {
		sw, sh = 1<<20, 1<<20
	}
	d := NewDropdownShowText(x, y, 26, 26, tc, items, 25, sw, sh)
	d.SetToolbarColour(tc) // translucent: kDropdownColourTable (setColourAndInputFlags)
	for i, c := range cs {
		if c.ID == owner {
			d.Highlighted = i // setHighlightedItem
		}
	}
	d.OnSelect = func(i int) { // onCompanyDropdown
		if i >= 0 && i < len(cs) && onCompany != nil {
			onCompany(cs[i].ID)
		}
	}
	return d
}
