package ui

import (
	"sort"
	"strconv"

	"github.com/LaPingvino/goloco/pkg/i18n"
)

// The industry list (Ui/Windows/IndustryList.cpp): its list tab — Name,
// Status, Production transported, Production last month — on the shared
// table window, 759 wide (kWindowSize). PARTIAL: the second tab (fund new
// industries) and the resizing.

// IndustryListRow is one industry: IndustryList.cpp drawScroll's cells.
type IndustryListRow struct {
	ID          int
	Name        string
	Status      string
	Produces    bool   // canProduceCargo: the two production cells are drawn
	Transported uint8  // percent (getAverageTransportedCargo)
	LastMonth   uint32 // getProductionLastMonth's amount
	Unit        string // its cargo's unit name ({COMMA32} …)
}

func (r IndustryListRow) cells() []string {
	c := []string{r.Name, r.Status, "", ""}
	if r.Produces {
		c[2] = FillArgs(i18n.Label(i18n.ProductionTransportedPercent), strconv.Itoa(int(r.Transported)))
		c[3] = FillArgs(ObjectString(r.Unit), Comma32(int64(r.LastMonth)))
	}
	return c
}

// industryOrder is IndustryList.cpp getOrder: name and status by string,
// production transported and last month descending.
func industryOrder(mode int, a, b IndustryListRow) bool {
	switch mode {
	case 0:
		return a.Name < b.Name
	case 1:
		return a.Status < b.Status
	case 2:
		return b.Transported < a.Transported
	case 3:
		return b.LastMonth < a.LastMonth
	}
	return false
}

// NewIndustryList opens the industry list at (sw - 759, 30), as
// IndustryList::open; rows open their industry's window via OnRow.
func NewIndustryList(sw int, industries []IndustryListRow, titleColour, colour uint8) *TownListWindow {
	const w759 = 759
	status := countLabel(len(industries), i18n.StatusNumIndustriesSingular, i18n.StatusNumIndustriesPlural)
	w := &TownListWindow{X: sw - w759, Y: 30, TitleColour: titleColour, Colour: colour, Caption: i18n.Label(i18n.TitleIndustries), Tabs: 2,
		Columns: []TableColumn{{4, 200, i18n.Label(i18n.IndustryTableHeader), 0, 198}, {204, 240, i18n.Label(i18n.IndustryTableHeaderStatus), 200, 238},
			{444, 160, i18n.Label(i18n.IndustryTableHeaderProduction), 440, 138}, {604, 151, i18n.Label(i18n.IndustryTableHeaderProductionLastMonth), 600, 138}},
		Status: status, Hover: -1}
	w.setTableGeometry()
	w.W, w.List = w759, [4]int{3, 56, w759 - 6, 126} // onResize: the scroll view to the frame
	w.Rows = make([][]string, len(industries))
	w.Sort = func(mode int) {
		order := append([]IndustryListRow(nil), industries...)
		sort.SliceStable(order, func(i, j int) bool { return industryOrder(mode, order[i], order[j]) })
		w.RowIDs = w.RowIDs[:0]
		for i, r := range order {
			w.Rows[i] = r.cells()
			w.RowIDs = append(w.RowIDs, r.ID)
		}
	}
	w.Sort(0)
	return w
}
