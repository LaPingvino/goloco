package sim

// Events for the company AIs (AI-INTERFACE.md G12): the news. Every message
// MessageManager::post files (PostMessage) that is addressed to the
// company, or to everyone (company 0xFF: an industry closing, a new vehicle,
// a promotion), becomes an AIEvent for that company's AI — its vehicle
// crashed, slipped, can't wait for a full load or can't land; cargo accepted
// or no longer at its station; its bankruptcy warnings. A player learns the
// same things the same way. Events wait for the company's next think and
// are handed over before it (OnEvent, then Think); they are saved with the
// AI.
//
// Breakdowns are events of their own: OpenLoco posts no news for one (a
// player sees the train stop and smoke), and its AI never looks at them
// (CompanyAi.cpp only uses breakdownPending as a "place me" marker). A car
// breaking down (updateBreakdown, breakdown.go) and its breakdown ending
// (VehicleHead::updateDaily, date.go) each reach the owner's AI, the way a
// player sees them on the map.

// AIEvent is one piece of news as the AI hears it: the message type, its
// three subjects (as Message.Subjects: a vehicle head, station, town,
// industry or company id by the type's ArgumentTypes; 0xFFFF none) and the
// day it was posted.
type AIEvent struct {
	Type     MessageType
	Subjects [3]uint16
	Day      uint32
}

// goloco's own event types, past every MessageType (no message is posted for
// them). Subjects: the vehicle head's id and the car's front bogie id (each
// car breaks down by itself; the train stops while any powered car is
// broken down), 0xFFFF.
const (
	EventVehicleBrokenDown MessageType = 0xF0 + iota // a car broke down
	EventVehicleRepaired                             // its breakdown is over
)

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
	for id := range m.companyAIs {
		if msg.CompanyID == uint8(id) || msg.CompanyID == companyIDNull {
			m.aiEvent(id, AIEvent{Type: msg.Type, Subjects: msg.Subjects, Day: msg.Date})
		}
	}
}

// aiVehicleEvent gives one of goloco's own vehicle events to the owner's AI.
func (m *Map) aiVehicleEvent(owner uint8, t MessageType, head, car uint16) {
	m.aiEvent(int(owner), AIEvent{Type: t, Subjects: [3]uint16{head, car, 0xFFFF}, Day: m.CurrentDay})
}

// aiEvent queues e for company id's AI, if it listens.
func (m *Map) aiEvent(id int, e AIEvent) {
	if id < 0 || id >= len(m.companyAIs) {
		return
	}
	st := m.companyAIs[id]
	if st == nil {
		return
	}
	if _, ok := st.ai.(AIEventHandler); !ok {
		return
	}
	if len(st.events) >= aiEventsMax {
		st.events = st.events[1:]
	}
	st.events = append(st.events, e)
}
