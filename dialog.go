package fbbot

import (
	log "github.com/Sirupsen/logrus"
)

// Event represents event triggered by state
type Event string

const ResetEvent Event = "reset"
const NilEvent Event = ""

type State interface {
	Enter(*Bot, *Message) Event
	Process(*Bot, *Message) Event
	Leave(*Bot, *Message) Event
}

// BaseState is base struct for State
type BaseState struct {
	Name string
}

func (s BaseState) Enter(bot *Bot, msg *Message) (e Event)   { return e } // Do nothing
func (s BaseState) Process(bot *Bot, msg *Message) (e Event) { return e } // Do nothing
func (s BaseState) Leave(bot *Bot, msg *Message) (e Event)   { return e } // Do nothing

type Dialog struct {
	beginState State
	endState   State

	states   map[State]bool   // stores all states of this dialog
	stateMap map[string]State // maps an user ID to his current state
	transMap map[State]map[Event]State
}

func NewDialog() *Dialog {
	var d Dialog
	d.states = make(map[State]bool)
	d.stateMap = make(map[string]State)
	d.transMap = make(map[State]map[Event]State)

	return &d
}

func (d *Dialog) AddStates(states ...State) {
	for _, state := range states {
		d.states[state] = true
	}
}

func (d *Dialog) SetBeginState(s State) {
	d.beginState = s
}

func (d *Dialog) SetEndState(s State) {
	d.endState = s
}

func (d *Dialog) AddTransition(src State, event Event, dst State) {
	_, exist := d.transMap[src]
	if !exist {
		d.transMap[src] = make(map[Event]State)
	}
	d.transMap[src][event] = dst
}

func (d *Dialog) AddGlobalTransition(event Event, dst State) {
	for state := range d.states {
		d.AddTransition(state, event, dst)
	}
}

func (d *Dialog) Handle(bot *Bot, msg *Message) {
	if d.beginState == nil || d.endState == nil {
		log.Fatal("BeginState and EndState are not set.")
	}

	var event Event
	state := d.getState(msg.Sender.ID)
	if state == nil || state == d.endState {
		bot.STMemory.Delete(msg.Sender.ID)
		d.setState(msg.Sender.ID, d.beginState)
		state = d.getState(msg.Sender.ID)
		event = state.Enter(bot, msg)
	} else {
		event = state.Process(bot, msg)
	}
	d.transition(bot, msg, state, event)
}

func (d *Dialog) transition(bot *Bot, msg *Message, src State, event Event) {
	if event == ResetEvent {
		d.resetState(msg.Sender.ID)
		return
	}

	dst, exist := d.transMap[src][event]
	if !exist {
		return
	}
	src.Leave(bot, msg)
	d.setState(msg.Sender.ID, dst)
	event = d.getState(msg.Sender.ID).Enter(bot, msg)
	d.transition(bot, msg, dst, event)
}

func (d *Dialog) setState(user_id string, state State) {
	d.stateMap[user_id] = state
}

func (d *Dialog) getState(user_id string) State {
	return d.stateMap[user_id]
}

func (d *Dialog) resetState(user_id string) {
	delete(d.stateMap, user_id)
}
