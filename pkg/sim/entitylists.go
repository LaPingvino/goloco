package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// EntityManager's lists (Entities/EntityManager.cpp), item 5P.a: every entity
// slot sits in one linked list — free (null), free for money effects
// (nullMoney), vehicles, misc … — and a new entity takes the head of the free
// list, so the slot (and id) each new vehicle gets follows upstream. Kept from
// the save, or rebuilt by reset for a new game.

// EntityListType (EntityManager.h).
const (
	EntityListNull        = 0
	EntityListNullMoney   = 1
	EntityListVehicleHead = 2
	EntityListMisc        = 4
	EntityListVehicle     = 6
)

const (
	maxEntities       = 20000 // Limits::kMaxEntities
	maxNormalEntities = 19800 // Limits::maxNormalEntities (the rest: money effects)
	entityIDNull      = EntityIDNull
)

// EntityLists is the manager's state: the lists' heads and counts and every
// slot's link.
type EntityLists struct {
	Heads, Counts [7]uint16
	Links         []scenario.EntityLink
}

// resetEntityLists is EntityManager::reset: every slot free, 0–19799 in the
// null list and the rest in nullMoney, each in id order.
func resetEntityLists() *EntityLists {
	l := &EntityLists{Links: make([]scenario.EntityLink, maxEntities)}
	for i := range l.Heads {
		l.Heads[i] = entityIDNull
	}
	build := func(from, to int, list uint8) {
		l.Heads[list] = uint16(from)
		for id := from; id < to; id++ {
			ln := scenario.EntityLink{BaseType: scenario.EntityBaseTypeFree, ListOffset: list * 2, Next: entityIDNull, Prev: entityIDNull}
			if id > from {
				ln.Prev = uint16(id - 1)
				l.Links[id-1].Next = uint16(id)
			}
			l.Links[id] = ln
		}
		l.Counts[list] = uint16(to - from)
	}
	build(0, maxNormalEntities, EntityListNull)
	build(maxNormalEntities, maxEntities, EntityListNullMoney)
	return l
}

// moveEntityToList is EntityManager::moveEntityToList (0x0047019F): out of
// its list, in at the head of list.
func (l *EntityLists) moveEntityToList(id uint16, list uint8) {
	e := &l.Links[id]
	if e.ListOffset == list*2 {
		return
	}
	old := e.ListOffset / 2
	if e.Prev == entityIDNull {
		l.Heads[old] = e.Next
	} else {
		l.Links[e.Prev].Next = e.Next
	}
	if e.Next != entityIDNull {
		l.Links[e.Next].Prev = e.Prev
	}
	e.Prev = entityIDNull
	e.ListOffset = list * 2
	e.Next = l.Heads[list]
	l.Heads[list] = id
	if e.Next != entityIDNull {
		l.Links[e.Next].Prev = id
	}
	l.Counts[old]--
	l.Counts[list]++
}

// createEntityVehicle is EntityManager::createEntityVehicle: the free list's
// head, moved to the vehicle list — its id; false when none is free.
func (l *EntityLists) createEntityVehicle() (uint16, bool) {
	if l.Counts[EntityListNull] == 0 {
		return 0, false
	}
	id := l.Heads[EntityListNull]
	l.moveEntityToList(id, EntityListVehicle)
	l.Links[id].BaseType = scenario.EntityBaseTypeVehicle
	return id, true
}

// freeEntity is EntityManager::freeEntity: the slot back at the head of its
// free list (nullMoney for the money effects' slots).
func (l *EntityLists) freeEntity(id uint16) {
	list := uint8(EntityListNull)
	if id >= maxNormalEntities {
		list = EntityListNullMoney
	}
	l.moveEntityToList(id, list)
	l.Links[id].BaseType = scenario.EntityBaseTypeFree
}

// numFreeEntities is getListCount(null) — checkNumFreeEntities' count.
func (l *EntityLists) numFreeEntities() int { return int(l.Counts[EntityListNull]) }

// EntityIDNull is EntityId::null: no entity (a Build Vehicle window with no
// target vehicle, …).
const EntityIDNull = 0xFFFF
