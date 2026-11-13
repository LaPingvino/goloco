package sim

// The news' share of MessageManager (MessageManager.cpp sub_428E47,
// getActiveHighestPriority) and NewsWindow::open's choice of what shows a
// message (Ui/Windows/News/Common.cpp open): none, the ticker in the bottom
// bar, the advice panel or the newspaper. The windows themselves are the
// game's (cmd/goloco/news.go, pkg/ui/newswindow.go); this is what decides,
// kept free of them so it can be tested.
//
// OpenLoco reference: src/OpenLoco/src/MessageManager.cpp, Message.h,
// Ui/Windows/News/Common.cpp

// NewsType is Config::NewsType: how a criticality's messages are shown.
type NewsType uint8

const (
	NewsNone NewsType = iota
	NewsTicker
	NewsWindow
)

// MessageCriticality (Message.h): the index of a news setting.
const (
	MsgCritMajorCompany uint8 = iota
	MsgCritMajorCompetitor
	MsgCritMinorCompany
	MsgCritMinorCompetitor
	MsgCritGeneral
	MsgCritAdvice
	MessageCriticalityCount = 6
)

// MessageTypeFlags (Message.h).
const (
	MsgFlagIsGeneralNews = 1 << 1 // the newspaper of its time; unset, the red advice panel
	MsgFlagHasFirstItem  = 1 << 2
	MsgFlagHasSecondItem = 1 << 3
	MsgFlagSmallerFont   = 1 << 5 // never set
)

// MessageIDNull is MessageId::null.
const MessageIDNull = messageIDNull

// Descriptor is getMessageTypeDescriptor (a zero one for a type out of range).
func (e *Message) Descriptor() MessageTypeDescriptor {
	if int(e.Type) >= len(MessageTypeDescriptors) {
		return MessageTypeDescriptor{ArgumentTypes: [3]uint8{MsgArgNull, MsgArgNull, MsgArgNull}}
	}
	return MessageTypeDescriptors[e.Type]
}

// IsActive, IsUserSelected, SetActive and SetUserSelected are Message's.
func (e *Message) IsActive() bool       { return e.TimeActive != 0xFFFF }
func (e *Message) IsUserSelected() bool { return e.TimeActive&(1<<15) != 0 }
func (e *Message) SetActive(on bool) {
	if on {
		e.TimeActive = 0
	} else {
		e.TimeActive = 0xFFFF
	}
}
func (e *Message) SetUserSelected() { e.SetActive(true); e.TimeActive |= 1 << 15 }

// Message is MessageManager::get: nil for an id past the messages.
func (m *Map) Message(id uint16) *Message {
	if int(id) >= len(m.Messages) {
		return nil
	}
	return &m.Messages[id]
}

// ClearActiveMessage is clearActiveMessage (0x00428E0F).
func (m *Map) ClearActiveMessage() { m.clearActiveMessage() }

// ActiveHighestPriorityMessage is getActiveHighestPriority (0x00428F38): of
// the active messages the one of the highest priority, the first of equals —
// one the player opened before all others; MessageIDNull when none is active.
func (m *Map) ActiveHighestPriorityMessage() uint16 {
	highest, best := uint8(0), uint16(MessageIDNull)
	for i := range m.Messages {
		e := &m.Messages[i]
		if !e.IsActive() {
			continue
		}
		p := e.Descriptor().Priority
		if e.IsUserSelected() {
			p = 0xFF
		}
		if p > highest {
			highest, best = p, uint16(i)
		}
	}
	return best
}

// NewsHost is what sub_428E47 needs of the game: the news window
// (WindowManager::find / close of WindowType::news, NewsWindow::open), the
// game speed and the news settings.
type NewsHost interface {
	NewsWindowOpen() bool
	OpenNews(id uint16)
	CloseNews()
	SpeedNormal() bool
	NewsSetting(criticality uint8) NewsType
}

// UpdateNews is sub_428E47 (0x00428E47), once per interface update: with no
// news window, the active message of the highest priority opens; with one,
// a message of a higher priority takes its place, else the shown message's
// time runs — 480 updates … 288 with more messages after it, less at a
// faster game speed — and then it is cleared. Quirks kept: the "messages
// waiting" count is numMessages − the active index, so never 0 (the 480 is
// never used); with no active message left of any priority the shown one's
// time stands still.
func (m *Map) UpdateNews(h NewsHost) {
	if m.TitleMode {
		return
	}
	if !h.NewsWindowOpen() {
		if id := m.ActiveHighestPriorityMessage(); id != MessageIDNull {
			h.OpenNews(id)
		}
		return
	}
	if m.ActiveMessageIndex == MessageIDNull {
		m.clearActiveMessage()
		return
	}
	id := m.ActiveHighestPriorityMessage()
	if id != MessageIDNull && id != m.ActiveMessageIndex {
		h.CloseNews()
		m.ActiveMessageIndex = MessageIDNull
		h.OpenNews(id)
		return
	}
	e := m.Message(id)
	if e == nil {
		return
	}
	e.TimeActive++
	waiting := len(m.Messages) - int(m.ActiveMessageIndex)
	var time uint16
	switch {
	case waiting == 0:
		time = 480
	case waiting <= 2:
		time = 384
	case waiting <= 4:
		time = 320
	default:
		time = 288
	}
	time2 := uint16(0xFFFF)
	if !h.SpeedNormal() && waiting > 0 {
		if waiting <= 3 {
			time2 = 256
		} else if h.NewsSetting(e.Descriptor().Criticality) == NewsWindow {
			time2 = 128
		}
	}
	if min(time, time2) > e.TimeActive&0x7FFF {
		return
	}
	m.clearActiveMessage()
}

// NewsKind is what NewsWindow::open makes of a message.
type NewsKind uint8

const (
	NewsKindNone   NewsKind = iota // the setting says none: the message no longer active
	NewsKindTicker                 // the ticker in the bottom bar
	NewsKindAdvice                 // News1: the red panel, 360 × 117
	NewsKindPaper                  // News2: the newspaper, 360 × 159
)

// SoundIDNull is Audio::SoundId::null.
const SoundIDNull = 0xFFFF

// NewsOpening is NewsWindow::open's outcome: the kind of window, whether the
// message is old news (seen before: no slide-in, no sound, never the ticker)
// and the sound to play (SoundIDNull: none).
type NewsOpening struct {
	Kind  NewsKind
	Old   bool
	Sound uint16
}

// OpenNews is NewsWindow::open (0x00428F8B) up to the window: the message
// becomes the active one; it is old when it has been active before and the
// scene is 10 updates old. A new one goes by the news setting of its
// criticality — a company's major / minor news counted as a competitor's
// when the company isn't the player's —: none makes it inactive (the active
// index stays, as upstream), ticker shows the ticker. The sound of a new
// message is the notification, or the type's own when the message is no
// company's or the player's; none with news sounds off. ok false: no such
// message.
func (m *Map) OpenNews(id uint16, sceneAge int, setting func(criticality uint8) NewsType, playSounds bool) (o NewsOpening, ok bool) {
	e := m.Message(id)
	if e == nil {
		return o, false
	}
	o.Sound = SoundIDNull
	o.Old = e.TimeActive != 0 && sceneAge >= 10
	m.ActiveMessageIndex = id
	d := e.Descriptor()
	if !o.Old {
		crit := d.Criticality
		if crit == MsgCritMajorCompany && e.CompanyID != m.ControllingCompany {
			crit = MsgCritMajorCompetitor
		}
		if crit == MsgCritMinorCompany && e.CompanyID != m.ControllingCompany {
			crit = MsgCritMinorCompetitor
		}
		switch setting(crit) {
		case NewsNone:
			e.SetActive(false)
			return o, true
		case NewsTicker:
			o.Kind = NewsKindTicker
			return o, true
		}
		if playSounds {
			o.Sound = soundNotification
			if e.CompanyID == companyIDNull || e.CompanyID == m.ControllingCompany {
				o.Sound = uint16(d.Sound)
			}
		}
	}
	o.Kind = NewsKindAdvice
	if d.Flags&MsgFlagIsGeneralNews != 0 {
		o.Kind = NewsKindPaper
	}
	return o, true
}

const soundNotification = 11 // Audio::SoundId::notification

// SelectMessage is what MessageWindow's scrollMouseDown and NewsWindow::
// openLastMessage do before they open a message of the player's choice: an
// active message the player had opened leaves the queue, the active index is
// cleared, and message id becomes the player's (active, user selected, its
// time at 1). The caller closes the news window before and opens the
// message after. false: no such message.
func (m *Map) SelectMessage(id uint16) bool {
	if e := m.Message(m.ActiveMessageIndex); e != nil && e.IsActive() && e.IsUserSelected() {
		e.SetActive(false)
	}
	m.ActiveMessageIndex = MessageIDNull
	e := m.Message(id)
	if e == nil {
		return false
	}
	e.SetUserSelected()
	e.TimeActive++
	return true
}
