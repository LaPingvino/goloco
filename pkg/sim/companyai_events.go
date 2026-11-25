package sim

// Events for the company AIs (AI-INTERFACE.md G12): the news. Every message
// MessageManager::post files (PostMessage) that is addressed to the
// company, or to everyone (company 0xFF: an industry closing, a new vehicle,
// a promotion), becomes an AIEvent for that company's AI — its vehicle
// crashed, slipped, can't wait for a full load or can't land; cargo accepted
// or no longer at its station; its bankruptcy warnings. A player learns the
// same things the same way. Events wait for the company's next think and
// are handed over before it (OnEvent, then Think); they are saved with the
// AI. Breakdowns are not among them: OpenLoco posts no news for one, and
// goloco doesn't break vehicles down yet (BreakdownFlags::brokenDown is
// never set); when it does, a breakdown is an event of its own.

// AIEvent is one piece of news as the AI hears it: the message type, its
// three subjects (as Message.Subjects: a vehicle head, station, town,
// industry or company id by the type's ArgumentTypes; 0xFFFF none) and the
// day it was posted.
type AIEvent struct {
	Type     MessageType
	Subjects [3]uint16
	Day      uint32
}

// AIEventHandler is a CompanyAI that listens to events (optional).
type AIEventHandler interface {
	// OnEvent is called at the company's think, before Think, for each event
	// since the last one, oldest first. v is as in Think: the AI may act,
	// or end its company.
	OnEvent(v *AIView, e AIEvent)
}

// aiEventsMax bounds the events kept for one think (the oldest go first).
const aiEventsMax = 64

// aiNews gives a posted message to the AIs it is addressed to.
func (m *Map) aiNews(msg Message) {
	for id, st := range m.companyAIs {
		if st == nil || (msg.CompanyID != uint8(id) && msg.CompanyID != companyIDNull) {
			continue
		}
		if _, ok := st.ai.(AIEventHandler); !ok {
			continue
		}
		if len(st.events) >= aiEventsMax {
			st.events = st.events[1:]
		}
		st.events = append(st.events, AIEvent{Type: msg.Type, Subjects: msg.Subjects, Day: msg.Date})
	}
}
