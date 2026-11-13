package main

import (
	"log"
	"strings"
	"unicode/utf8"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// news.go — the news as the game runs it: MessageManager's update once per
// interface update (sub_428E47, pkg/sim/news.go), NewsWindow::open's windows
// (the ticker, the advice panel, the newspaper: pkg/ui/newswindow.go), the
// Messages window (pkg/ui/messagewindow.go) and the news section of the
// config.
//
// OpenLoco reference: src/OpenLoco/src/MessageManager.cpp, OpenLoco.cpp
// sub_431695, Ui/Windows/News/Common.cpp, News.cpp, Ticker.cpp,
// Ui/Windows/MessageWindow.cpp, Config.cpp
//
// PLATFORM: goloco has no sound effects yet — a new message's sound, the
// ticker's and the row click's are playNewsSound, which plays nothing.
// PARTIAL: of the keyboard shortcuts (Shortcuts.cpp) Tab opens the Messages
// window (showAnnouncementsList's default); toggleLastAnnouncement (Space
// upstream, goloco's pause key: NewsWindow::openLastMessage / close) is not
// ported.

// newsSettingNames are ConfigConvert.hpp's kNewsTypeEntries; newsConfigKeys
// the news section's keys by MessageCriticality (Config.cpp).
var (
	newsSettingNames = [...]string{"none", "ticker", "newsWindow"}
	newsConfigKeys   = [sim.MessageCriticalityCount]string{"majorCompany", "majorCompetitor", "minorCompany", "minorCompetitor", "general", "advice"}
)

// loadNewsConfig is Config.cpp's read of the news section and of
// audio.playNewsSounds: every kind of news a window, the sounds on, unless
// the file says otherwise.
func loadNewsConfig() (settings [sim.MessageCriticalityCount]sim.NewsType, playSounds bool) {
	for i, key := range newsConfigKeys {
		settings[i] = sim.NewsWindow
		for v, name := range newsSettingNames {
			if configValue("news."+key) == name {
				settings[i] = sim.NewsType(v)
			}
		}
	}
	return settings, configValue("audio.playNewsSounds") != "false"
}

// writeNewsConfig is Config::write for the same keys.
func (g *Game) writeNewsConfig() {
	kvs := [][2]string{{"audio.playNewsSounds", map[bool]string{false: "false", true: "true"}[g.playNewsSounds]}}
	for i, key := range newsConfigKeys {
		kvs = append(kvs, [2]string{"news." + key, newsSettingNames[g.newsSettings[i]]})
	}
	if err := setConfigValues(kvs); err != nil {
		log.Printf("[News] saving the news settings: %v", err)
	}
}

// playNewsSound is Audio::playSound for the news (PLATFORM: no sound
// effects — nothing plays).
func (g *Game) playNewsSound(id uint16) {}

// newsHost is the game as MessageManager's update sees it (sim.NewsHost).
type newsHost struct{ g *Game }

func (h newsHost) NewsWindowOpen() bool { return h.g.newsWin() != nil }
func (h newsHost) OpenNews(id uint16)   { h.g.openNews(id) }
func (h newsHost) CloseNews()           { h.g.closeNews() }
func (h newsHost) SpeedNormal() bool    { return h.g.speedMult <= 1 }
func (h newsHost) NewsSetting(criticality uint8) sim.NewsType {
	return h.g.newsSetting(criticality)
}

func (g *Game) newsSetting(criticality uint8) sim.NewsType {
	if int(criticality) >= len(g.newsSettings) {
		return sim.NewsWindow
	}
	return g.newsSettings[criticality]
}

// newsWin is WindowManager::find(WindowType::news): the news window or the
// ticker (nil: neither).
func (g *Game) newsWin() ui.Win { return g.wins.Find(ui.WindowTypeNews, ui.AnyNumber) }

// closeNews is WindowManager::close(WindowType::news).
func (g *Game) closeNews() { g.wins.CloseType(ui.WindowTypeNews, ui.AnyNumber) }

func isNews(w ui.Win) bool { return w.Base().WinType == ui.WindowTypeNews }

// resetNews is a new scene (SceneManager's scene age back at 0; the windows
// are closed by the scene's start).
func (g *Game) resetNews() { g.sceneAge = 0 }

// updateNews is the news' share of a tick() of n game ticks (0: no tick this
// frame): the vehicle preview turns by n; then, once per interface update
// (_numFrameUpdates: n, 1 while paused — sub_431695), MessageManager's
// update and the news windows' onUpdate. Not on the title screen.
func (g *Game) updateNews(n int) {
	if n <= 0 || g.inTitleScreen {
		return
	}
	g.previewFrame = (g.previewFrame + n) & 0xFFFF
	if g.isPaused {
		n = 1
	}
	g.sceneAge = min(0xFFFF, g.sceneAge+n)
	m := &g.w.Map
	m.TitleMode = false
	m.CloseNewsWindow = g.closeNews // clearActiveMessage's close, wherever the sim clears it
	for range n {
		m.UpdateNews(newsHost{g})
		switch w := g.newsWin().(type) {
		case *ui.NewsWindow:
			w.Update(g.sw, g.sh)
		case *ui.NewsTicker:
			w.Update(g.sw, g.sh, g.wins.FindAtAlt(g.mouseX, g.mouseY) == ui.Win(w), g.isPaused)
			if m.ActiveMessageIndex == sim.MessageIDNull { // Ticker::onUpdate: it goes with its message
				g.wins.Close(w)
			}
		}
	}
}

// openNews is NewsWindow::open (0x00428F8B): the message as the active one,
// shown as its type and the news settings say (sim.OpenNews).
func (g *Game) openNews(id uint16) {
	m := &g.w.Map
	o, ok := m.OpenNews(id, g.sceneAge, g.newsSetting, g.playNewsSounds)
	if !ok || o.Kind == sim.NewsKindNone {
		return
	}
	g.syncScreen()
	if o.Kind == sim.NewsKindTicker {
		t := ui.NewNewsTicker(g.sw, g.sh)
		g.fillTicker(t)
		g.wins.Open(t)
		return
	}
	if o.Sound != sim.SoundIDNull {
		g.playNewsSound(o.Sound)
	}
	w := ui.NewNewsWindow(o.Kind == sim.NewsKindPaper, o.Old, g.speedMult <= 1, g.sw, g.sh)
	w.OnDismiss = m.ClearActiveMessage // close_button: MessageManager::clearActiveMessage
	w.OnSubject = g.openNewsSubject
	g.fillNews(w)
	g.wins.Open(w)
}

// selectMessage is MessageWindow's scrollMouseDown: the message shown as
// news, whatever was showing closed.
func (g *Game) selectMessage(i int) {
	m := &g.w.Map
	if i < 0 || i >= len(m.Messages) {
		return
	}
	m.CloseNewsWindow = g.closeNews
	g.closeNews()
	m.SelectMessage(uint16(i)) // the player's earlier choice out of the queue, this one in
	g.openNews(uint16(i))
	g.playNewsSound(0) // SoundId::clickDown
}

// locoName is a name from the sim for the interface: user and object strings
// are Loco bytes, a language pack's are UTF-8.
func locoName(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return ui.FromLocoBytes(s)
}

func messageDate(day uint32) string {
	y, mo, d := uilayout.CalcDate(day)
	return uilayout.DateDMY(d, i18n.Month(mo+1), y)
}

// fillTicker gives the ticker its message (Ticker's draw reads the active
// one each time).
func (g *Game) fillTicker(t *ui.NewsTicker) {
	if e := g.w.Map.Message(g.w.Map.ActiveMessageIndex); e != nil {
		t.General = e.Descriptor().Flags&sim.MsgFlagIsGeneralNews != 0
		t.Text = ui.FromLocoBytes(e.Text)
	}
}

// fillNews gives the news window the active message as Common's draw and
// initViewports read it each time: its text and date, and its subjects. With
// no active message (cleared this frame) the window keeps what it had.
func (g *Game) fillNews(w *ui.NewsWindow) {
	m := &g.w.Map
	e := m.Message(m.ActiveMessageIndex)
	if e == nil {
		return
	}
	d := e.Descriptor()
	w.Year, _, _ = uilayout.CalcDate(e.Date)
	w.Date = messageDate(e.Date)
	w.Text = ui.FromLocoBytes(e.Text)
	w.HasSecond = d.Flags&sim.MsgFlagHasSecondItem != 0
	w.PreviewFrame = g.previewFrame
	for i := range w.Subjects {
		w.Subjects[i] = g.newsSubject(e, i)
	}
}

// newsSubject is getView and drawViewportString for subject i (0, 1) of a
// message: none unless the type has it and the message still names one.
func (g *Game) newsSubject(e *sim.Message, i int) ui.NewsSubject {
	m, om := &g.w.Map, g.r.ObjMgr
	d := e.Descriptor()
	id := e.Subjects[i]
	if d.Flags&(sim.MsgFlagHasFirstItem<<i) == 0 || id == 0xFFFF {
		return ui.NewsSubject{}
	}
	argType := d.ArgumentTypes[i]
	s := ui.NewsSubject{Present: true}
	view := func() {
		if !g.w.HasNewsView(argType, id) {
			return
		}
		s.Kind, s.Selectable = ui.NewsSubjectView, true
		s.View = func(vw, vh int) *gfx8.RenderTarget { return g.w.NewsView(i, argType, id, vw, vh) }
		s.Labels = func(dst *ebiten.Image) { g.w.NewsViewLabels(argType, id, dst) }
	}
	switch argType {
	case sim.MsgArgIndustry:
		for k := range m.Industries {
			if uint16(m.Industries[k].ID) == id && m.Industries[k].Name != 0xFFFF {
				name, _ := m.IndustryName(k, om)
				s.Caption = locoName(name)
			}
		}
		view()
	case sim.MsgArgStation:
		s.Caption = locoName(m.StationName(int(id)))
		view()
	case sim.MsgArgTown:
		if t := m.TownByID(id); t != nil {
			s.Caption = locoName(m.TownName(t.Name, int(t.ID)))
		}
		view()
	case sim.MsgArgVehicle:
		tr, err := m.TrainOf(id)
		if err != nil {
			return s // "we can't view old news about removed vehicles": no caption, no view
		}
		h := &m.Entities[tr.Head]
		// company_vehicle "{POP16}{POP16}{STRINGID}": the vehicle's name; a
		// competitor's competitor_vehicle "{STRINGID} - {STRINGID}".
		s.Caption = locoName(m.VehicleName(h.Name, int(h.OrdinalNumber)))
		if !m.IsPlayerCompany(int(h.Owner)) {
			c := strings.Replace(i18n.Label(i18n.CompetitorVehicle), "{STRINGID}", locoName(m.CompanyName(int(h.Owner))), 1)
			s.Caption = strings.Replace(c, "{STRINGID}", s.Caption, 1)
		}
		view()
	case sim.MsgArgCompany:
		s.Kind, s.Selectable = ui.NewsSubjectFace, true
		s.Caption = locoName(m.CompanyName(int(id)))
		if face, colour := g.companyFace(int(id)); face > 0 {
			s.Face, s.FaceColour = face+1, colour // images[ownerEmotion] + 1
		}
		s.Jailed = int(id) < len(m.Companies) && m.Companies[id].JailStatus != 0
	case sim.MsgArgVehicleTab:
		s.Kind, s.Selectable = ui.NewsSubjectVehicle, true
		if om != nil && int(id) < len(om.Vehicles) && om.Vehicles[id] != nil {
			v := om.Vehicles[id]
			s.Vehicle, s.VehicleYOffset, s.Caption = v, om.VehicleYDisplayOffset(v), locoName(v.DisplayName)
			// getCompanyVehicleColourScheme: the player's main colours, or its
			// own for the vehicle's colour type.
			if c := int(m.ControllingCompany); c < len(m.CompanyColourSchemes) {
				cs := &m.CompanyColourSchemes[c]
				s.Primary, s.Secondary = cs.Main[0], cs.Main[1]
				if int(v.ColourType) < len(cs.Vehicle) && cs.CustomSet&(1<<v.ColourType) != 0 {
					s.Primary, s.Secondary = cs.Vehicle[v.ColourType][0], cs.Vehicle[v.ColourType][1]
				}
			}
		}
	}
	return s
}

// openNewsSubject is Common::onMouseUp's viewport buttons: the subject's
// window — the industry's, station's, town's, vehicle's or company's, the
// Build Vehicle window for a new vehicle.
func (g *Game) openNewsSubject(i int) {
	m := &g.w.Map
	e := m.Message(m.ActiveMessageIndex)
	if e == nil || i < 0 || i > 1 {
		return
	}
	d := e.Descriptor()
	if d.Flags&(sim.MsgFlagHasFirstItem<<i) == 0 {
		return
	}
	id := e.Subjects[i]
	switch d.ArgumentTypes[i] {
	case sim.MsgArgIndustry:
		g.openIndustryWindow(int(id))
	case sim.MsgArgStation:
		g.openStationWindow(int(id))
	case sim.MsgArgTown:
		for k := range m.Towns { // goloco's Town windows go by the town's place in the list
			if uint16(m.Towns[k].ID) == id {
				g.openTownWindow(k)
			}
		}
	case sim.MsgArgVehicle:
		if _, err := m.TrainOf(id); err == nil {
			g.openVehicleWindow(id, 0)
		}
	case sim.MsgArgCompany:
		g.openCompanyWindow(int(id))
	case sim.MsgArgVehicleTab: // BuildVehicle::openByVehicleObjectId (PARTIAL: the vehicle's row not selected)
		if om := g.r.ObjMgr; om != nil && int(id) < len(om.Vehicles) && om.Vehicles[id] != nil {
			g.openBuildVehicle(int(om.Vehicles[id].Type))
		}
	}
}

// messagesWin is the Messages window (nil: not open).
func (g *Game) messagesWin() *ui.MessagesWindow {
	return winOf[*ui.MessagesWindow](g, ui.WindowTypeMessages, ui.AnyNumber)
}

// openMessagesWindow is MessageWindow::open (0x0042A3FF; the time panel's
// date): the window on its Messages tab, scrolled to the newest — a new one
// at the screen's right under the toolbar, the windows under it moved down.
func (g *Game) openMessagesWindow() {
	w := g.messagesWin()
	if w != nil {
		g.wins.BringToFront(w)
	} else {
		company, _ := g.w.ControllingCompany()
		_, wc := g.windowColours(true)
		var tc uint8
		if company >= 0 && company < len(g.w.CompanyColours) {
			tc = g.w.CompanyColours[company] // the owner's colour: the window is the player's
		}
		g.syncScreen()
		w = ui.NewMessagesWindow(g.sw, tc, wc)
		w.Rows = func() []ui.MessageRow {
			msgs := g.w.Map.Messages
			rows := make([]ui.MessageRow, len(msgs))
			for i := range msgs {
				rows[i] = ui.MessageRow{Date: messageDate(msgs[i].Date), Text: ui.FromLocoBytes(msgs[i].Text)}
			}
			return rows
		}
		w.OnRow = g.selectMessage
		w.OpenDropdown = func(d *ui.DropdownMenu) { g.dropdown = d }
		w.OnSetting = func(criticality, value int) { // onDropdown
			if criticality >= 0 && criticality < len(g.newsSettings) && value >= 0 && value < len(newsSettingNames) {
				g.newsSettings[criticality] = sim.NewsType(value)
				g.writeNewsConfig()
			}
		}
		w.OnPlaySounds = func() {
			g.playNewsSounds = !g.playNewsSounds
			g.writeNewsConfig()
		}
		g.prepareWindow(w)
		g.showList(w)
	}
	w.ShowMessages()
}
