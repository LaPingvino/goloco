package sim

import "testing"

type fakeNews struct {
	m        *Map
	open     bool
	normal   bool
	settings [MessageCriticalityCount]NewsType
	log      []string
	sceneAge int
}

func (f *fakeNews) NewsWindowOpen() bool { return f.open }
func (f *fakeNews) CloseNews()           { f.open = false; f.log = append(f.log, "close") }
func (f *fakeNews) SpeedNormal() bool    { return f.normal }
func (f *fakeNews) NewsSetting(c uint8) NewsType {
	return f.settings[c]
}
func (f *fakeNews) OpenNews(id uint16) {
	o, ok := f.m.OpenNews(id, f.sceneAge, f.NewsSetting, true)
	if ok && o.Kind != NewsKindNone {
		f.open = true
	}
	f.log = append(f.log, [...]string{"none", "ticker", "advice", "paper"}[o.Kind])
}

func newsMap(types ...MessageType) (*Map, *fakeNews) {
	m := &Map{ActiveMessageIndex: MessageIDNull, ControllingCompany: 0}
	for _, t := range types {
		m.Messages = append(m.Messages, Message{Type: t, CompanyID: 0})
	}
	f := &fakeNews{m: m, normal: true}
	for i := range f.settings {
		f.settings[i] = NewsWindow
	}
	m.CloseNewsWindow = f.CloseNews
	return m, f
}

// getActiveHighestPriority: the highest priority, the first of equals, a
// message the player opened before all.
func TestActiveHighestPriorityMessage(t *testing.T) {
	m, _ := newsMap(MsgCargoNowAccepted, MsgNewVehicle, MsgCompanyCheated, MsgNewVehicle)
	if got := m.ActiveHighestPriorityMessage(); got != 2 {
		t.Errorf("highest: %d, want 2 (companyCheated, priority 7)", got)
	}
	m.Messages[2].SetActive(false)
	if got := m.ActiveHighestPriorityMessage(); got != 1 {
		t.Errorf("highest: %d, want 1 (the first newVehicle)", got)
	}
	m.Messages[0].SetUserSelected()
	if got := m.ActiveHighestPriorityMessage(); got != 0 {
		t.Errorf("highest: %d, want 0 (user selected)", got)
	}
	for i := range m.Messages {
		m.Messages[i].SetActive(false)
	}
	if got := m.ActiveHighestPriorityMessage(); got != MessageIDNull {
		t.Errorf("highest: %d, want none", got)
	}
}

// NewsWindow::open: the kind by the type's flags and the news settings; a
// competitor's company news takes the competitor setting; old news never
// tickers.
func TestOpenNewsKind(t *testing.T) {
	all := func(v NewsType) func(uint8) NewsType { return func(uint8) NewsType { return v } }
	m, _ := newsMap(MsgCargoNowAccepted, MsgNewVehicle, MsgCompanyPromoted)
	if o, _ := m.OpenNews(0, 0, all(NewsWindow), true); o.Kind != NewsKindAdvice || o.Old || o.Sound != soundNotification {
		t.Errorf("advice: %+v", o)
	}
	if o, _ := m.OpenNews(1, 0, all(NewsWindow), true); o.Kind != NewsKindPaper || o.Sound != 24 {
		t.Errorf("paper: %+v", o)
	}
	if o, _ := m.OpenNews(1, 0, all(NewsWindow), false); o.Sound != SoundIDNull {
		t.Errorf("sounds off: %+v", o)
	}
	if o, _ := m.OpenNews(1, 0, all(NewsTicker), true); o.Kind != NewsKindTicker || m.ActiveMessageIndex != 1 {
		t.Errorf("ticker: %+v, active %d", o, m.ActiveMessageIndex)
	}
	// The player's promotion by majorCompany, a competitor's by majorCompetitor.
	only := func(c uint8) func(uint8) NewsType {
		return func(k uint8) NewsType {
			if k == c {
				return NewsTicker
			}
			return NewsWindow
		}
	}
	if o, _ := m.OpenNews(2, 0, only(MsgCritMajorCompany), true); o.Kind != NewsKindTicker {
		t.Errorf("own major news: %+v", o)
	}
	m.Messages[2].CompanyID = 3
	if o, _ := m.OpenNews(2, 0, only(MsgCritMajorCompany), true); o.Kind != NewsKindPaper || o.Sound != soundNotification {
		t.Errorf("a competitor's major news: %+v", o)
	}
	// none: inactive, the active index left on it (as upstream).
	if o, _ := m.OpenNews(0, 0, all(NewsNone), true); o.Kind != NewsKindNone || m.Messages[0].IsActive() || m.ActiveMessageIndex != 0 {
		t.Errorf("none: %+v, active %v, index %d", o, m.Messages[0].IsActive(), m.ActiveMessageIndex)
	}
	// Old news: seen before and the scene 10 updates old — the window, whatever the setting.
	m.Messages[1].TimeActive = 5
	if o, _ := m.OpenNews(1, 9, all(NewsTicker), true); o.Old || o.Kind != NewsKindTicker {
		t.Errorf("a young scene: %+v", o)
	}
	if o, _ := m.OpenNews(1, 10, all(NewsNone), true); !o.Old || o.Kind != NewsKindPaper || o.Sound != SoundIDNull {
		t.Errorf("old news: %+v", o)
	}
	if _, ok := m.OpenNews(9, 0, all(NewsWindow), true); ok {
		t.Error("a message past the list opened")
	}
}

// sub_428E47: the message of the highest priority opens, runs its time and
// is cleared; then the next.
func TestUpdateNews(t *testing.T) {
	m, f := newsMap(MsgCargoNowAccepted, MsgNewVehicle)
	m.UpdateNews(f)
	if !f.open || m.ActiveMessageIndex != 1 || f.log[0] != "paper" {
		t.Fatalf("first update: open %v active %d log %v", f.open, m.ActiveMessageIndex, f.log)
	}
	// One message from the active index on: 384 updates.
	n := 0
	for f.open && n < 1000 {
		m.UpdateNews(f)
		n++
	}
	if n != 384 || m.Messages[1].IsActive() || m.ActiveMessageIndex != MessageIDNull {
		t.Errorf("shown for %d updates (want 384), active %v, index %d", n, m.Messages[1].IsActive(), m.ActiveMessageIndex)
	}
	m.UpdateNews(f)
	if !f.open || m.ActiveMessageIndex != 0 || f.log[len(f.log)-1] != "advice" {
		t.Errorf("the next: open %v active %d log %v", f.open, m.ActiveMessageIndex, f.log)
	}
	// Faster than normal with at most three from the index on: 256 updates.
	f.normal = false
	for n = 0; f.open && n < 1000; n++ {
		m.UpdateNews(f)
	}
	if n != 256 {
		t.Errorf("at speed: shown for %d updates, want 256", n)
	}
	m.UpdateNews(f)
	if f.open {
		t.Error("a window with no active message left")
	}
}

// A message of a higher priority takes the window from the one shown; a
// removed active message closes it; the title shows none.
func TestUpdateNewsPreemptAndRemove(t *testing.T) {
	m, f := newsMap(MsgCargoNowAccepted)
	m.UpdateNews(f)
	m.Messages = append(m.Messages, Message{Type: MsgCompanyCheated})
	m.UpdateNews(f)
	if m.ActiveMessageIndex != 1 || !f.open || len(f.log) != 3 || f.log[1] != "close" || f.log[2] != "paper" {
		t.Fatalf("pre-empt: active %d open %v log %v", m.ActiveMessageIndex, f.open, f.log)
	}
	m.removeMessage(0) // below the active one: its index follows
	if m.ActiveMessageIndex != 0 || !f.open {
		t.Errorf("after removing below: active %d open %v", m.ActiveMessageIndex, f.open)
	}
	m.removeMessage(0) // the active one: cleared, the window closed
	if m.ActiveMessageIndex != MessageIDNull || f.open {
		t.Errorf("after removing the active one: active %d open %v", m.ActiveMessageIndex, f.open)
	}
	m, f = newsMap(MsgCargoNowAccepted)
	m.TitleMode = true
	m.UpdateNews(f)
	if f.open {
		t.Error("news on the title screen")
	}
}

// The setting none: the message goes inactive without a window, and the
// next update takes the next one.
func TestUpdateNewsNone(t *testing.T) {
	m, f := newsMap(MsgNewVehicle, MsgCargoNowAccepted)
	f.settings[MsgCritGeneral] = NewsNone
	m.UpdateNews(f)
	if f.open || m.Messages[0].IsActive() {
		t.Fatalf("none: open %v active %v", f.open, m.Messages[0].IsActive())
	}
	m.UpdateNews(f)
	if !f.open || m.ActiveMessageIndex != 1 {
		t.Errorf("the next: open %v active %d", f.open, m.ActiveMessageIndex)
	}
}

// MessageWindow's row click / openLastMessage: the player's earlier choice
// leaves the queue, the new one is user selected with its time at 1 — old
// news for NewsWindow::open.
func TestSelectMessage(t *testing.T) {
	m, _ := newsMap(MsgCargoNowAccepted, MsgNewVehicle)
	if !m.SelectMessage(0) || !m.Messages[0].IsUserSelected() || m.Messages[0].TimeActive != 0x8001 {
		t.Fatalf("select: %#x", m.Messages[0].TimeActive)
	}
	o, _ := m.OpenNews(0, 100, func(uint8) NewsType { return NewsTicker }, true)
	if !o.Old || o.Kind != NewsKindAdvice {
		t.Errorf("a selected message: %+v", o)
	}
	m.SelectMessage(1)
	if m.Messages[0].IsActive() || !m.Messages[1].IsUserSelected() {
		t.Errorf("the earlier choice still active (%#x), or the new not selected (%#x)", m.Messages[0].TimeActive, m.Messages[1].TimeActive)
	}
	if m.SelectMessage(7) {
		t.Error("selected a message past the list")
	}
}

// The texts post() formats for a new company and a crash.
func TestNewCompanyAndCrashText(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(m.Towns) == 0 {
		t.Skip("no towns")
	}
	m.PostMessage(om, MsgNewCompany, 0, 0, uint16(m.Towns[0].ID), 0xFFFF)
	got := m.Messages[len(m.Messages)-1].Text
	t.Logf("%q", got)
	if !contains(got, "New Transport Company!\x05") || !contains(got, m.TownName(m.Towns[0].Name, int(m.Towns[0].ID))) {
		t.Errorf("new company text %q", got)
	}
}
