package hierarchicalStateMachine

import (
	"reflect"
	"testing"
)

var executedActions []string // Track executed actions for verification

func recordAction(actionName string) Action {
	return func() {
		executedActions = append(executedActions, actionName)
	}
}

func resetExecutedActions() {
	executedActions = []string{}
}

func TestStateMachineInitialization(t *testing.T) {
	initialState := State{}
	states := []State{initialState}
	transitions := []Transition{}

	sm, err := NewHierarchicalStateMachine(&initialState, states, transitions)

	// Check for a valid state machine
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if sm.CurrentState != &initialState {
		t.Errorf("Expected current state to be %v, got %v", &initialState, sm.CurrentState)
	}

	// Check for too many states
	states = make([]State, MaxStates+1)
	_, err = NewHierarchicalStateMachine(&initialState, states, transitions)
	if err == nil {
		t.Fatalf("Expected an error due to too many states, got none")
	}
}

// TestFlatStateMachine simulates a state machine with no hiearchies and verifies its behavior
func TestFlatStateMachine(t *testing.T) {
	resetExecutedActions()

	state1 := State{
		Entry:  []Action{recordAction("State 1 Entry")},
		Exit:   []Action{recordAction("State 1 Exit")},
		Handle: []Action{recordAction("State 1 Handle")},
	}
	state2 := State{
		Entry:  []Action{recordAction("State 2 Entry")},
		Exit:   []Action{recordAction("State 2 Exit")},
		Handle: []Action{recordAction("State 2 Handle")},
	}

	states := []State{state1, state2}

	transitions := []Transition{
		{
			CurrentState: &state1,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 1 -> State 2 Transition")},
			NextState:    &state2,
		},
		{
			CurrentState: &state2,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 2 -> State 1 Transition")},
			NextState:    &state1,
		},
	}

	// Check for valid state machine
	sm, err := NewHierarchicalStateMachine(&state1, states, transitions)
	if err != nil {
		t.Fatalf("failed to initialize state machine: %v", err)
	}

	// Check current state
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &state1, sm.CurrentState)
	}

	// Check initial transition
	expectedActions := []string{"State 1 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()
	HandleStateMachine(sm) // Transition from State 1 to State 2

	// Check current state
	if sm.CurrentState != &state2 {
		t.Errorf("Expected current state to be %v, got %v", &state2, sm.CurrentState)
	}

	// Check first transition
	expectedActions = []string{"State 1 Handle", "State 1 Exit", "State 1 -> State 2 Transition", "State 2 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()

	HandleStateMachine(sm) // Transition from State 2 to State 3

	// Check current state
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &state2, sm.CurrentState)
	}

	// Check second transition
	expectedActions = []string{"State 2 Handle", "State 2 Exit", "State 2 -> State 1 Transition", "State 1 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}
}

// TestNilFlatStateMachine simulates a state machine with all nil actions
func TestNilFlatStateMachine(t *testing.T) {
	resetExecutedActions()

	state1 := State{}
	state2 := State{}

	states := []State{state1, state2}

	transitions := []Transition{
		{
			CurrentState: &state1,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      nil,
			NextState:    &state2,
		},
		{
			CurrentState: &state2,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      nil,
			NextState:    &state1,
		},
	}

	// Check for valid state machine
	sm, err := NewHierarchicalStateMachine(&state1, states, transitions)
	if err != nil {
		t.Fatalf("failed to initialize state machine: %v", err)
	}

	// Check current state
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &state1, sm.CurrentState)
	}

	// Check initial transition
	expectedActions := []string{}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()
	HandleStateMachine(sm) // Transition from State 1 to State 2

	// Check current state
	if sm.CurrentState != &state2 {
		t.Errorf("Expected current state to be %v, got %v", &state2, sm.CurrentState)
	}

	// Check first transition
	expectedActions = []string{}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()

	HandleStateMachine(sm) // Transition from State 2 to State 3

	// Check current state
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &state2, sm.CurrentState)
	}

	// Check second transition
	expectedActions = []string{}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}
}

// state1 and state2 are children of parent state. state3 has no parent.
// Test internal transition (state1->state2), and external transitions (in/out state3)
// Test sub transition (parentState -> state1)
func TestHierarchicalStateMachine(t *testing.T) {

	resetExecutedActions()

	parentState := State{
		Entry:  []Action{recordAction("Parent State Entry")},
		Exit:   []Action{recordAction("Parent State Exit")},
		Handle: []Action{recordAction("Parent State Handle")},
	}
	state1 := State{
		Entry:       []Action{recordAction("State 1 Entry")},
		Exit:        []Action{recordAction("State 1 Exit")},
		Handle:      []Action{recordAction("State 1 Handle")},
		ParentState: &parentState,
	}
	state2 := State{
		Entry:       []Action{recordAction("State 2 Entry")},
		Exit:        []Action{recordAction("State 2 Exit")},
		Handle:      []Action{recordAction("State 2 Handle")},
		ParentState: &parentState,
	}
	state3 := State{
		Entry:  []Action{recordAction("State 3 Entry")},
		Exit:   []Action{recordAction("State 3 Exit")},
		Handle: []Action{recordAction("State 3 Handle")},
	}

	states := []State{state1, state2, state3, parentState}

	transitions := []Transition{
		{
			CurrentState: &state1,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 1 -> State 2 Transition")},
			NextState:    &state2,
		},
		{
			CurrentState: &state2,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 2 -> State 3 Transition")},
			NextState:    &state3,
		},
		{
			CurrentState: &state3,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 3 -> Parent State Transition")},
			NextState:    &parentState,
		},
		{
			CurrentState: &parentState,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("Parent State -> State 1 Transition")},
			NextState:    &state1,
		},
	}

	// Check for valid state machine
	sm, err := NewHierarchicalStateMachine(&state1, states, transitions)
	if err != nil {
		t.Fatalf("failed to initialize state machine: %v", err)
	}

	// Check current state
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &state1, sm.CurrentState)
	}

	// Check initial transition
	expectedActions := []string{"Parent State Entry", "State 1 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()
	HandleStateMachine(sm) // Transition from State 1 to State 2

	// Check current state
	if sm.CurrentState != &state2 {
		t.Errorf("Expected current state to be %v, got %v", &state2, sm.CurrentState)
	}

	// Check first transition
	expectedActions = []string{"Parent State Handle", "State 1 Handle", "State 1 Exit", "State 1 -> State 2 Transition", "State 2 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()

	HandleStateMachine(sm) // Transition from State 2 to State 3

	// Check current state
	if sm.CurrentState != &state3 {
		t.Errorf("Expected current state to be %v, got %v", &state3, sm.CurrentState)
	}

	// Check second transition (State 2 -> State 3)
	expectedActions = []string{
		"Parent State Handle", "State 2 Handle", "State 2 Exit", "Parent State Exit",
		"State 2 -> State 3 Transition", "State 3 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()

	HandleStateMachine(sm) // Transition from State 3 to Parent State

	// Check current state
	if sm.CurrentState != &parentState {
		t.Errorf("Expected current state to be %v, got %v", &parentState, sm.CurrentState)
	}

	// Check third transition (State 3 -> Parent State)
	expectedActions = []string{"State 3 Handle", "State 3 Exit", "State 3 -> Parent State Transition", "Parent State Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()

	HandleStateMachine(sm) // Transition from Parent State to State 1

	// Check current state
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &state1, sm.CurrentState)
	}

	// Check fourth transition (Parent State -> State 1)
	expectedActions = []string{"Parent State Handle", "Parent State -> State 1 Transition", "State 1 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}
}

// parentState2 contains state2 and parentstate
// parentState contains state1
// state 3 has no parent
func TestNestedStates(t *testing.T) {

	resetExecutedActions()

	parentState2 := State{
		Entry:  []Action{recordAction("Parent State 2 Entry")},
		Exit:   []Action{recordAction("Parent State 2 Exit")},
		Handle: []Action{recordAction("Parent State 2 Handle")},
	}
	parentState := State{
		Entry:       []Action{recordAction("Parent State Entry")},
		Exit:        []Action{recordAction("Parent State Exit")},
		Handle:      []Action{recordAction("Parent State Handle")},
		ParentState: &parentState2,
	}
	state1 := State{
		Entry:       []Action{recordAction("State 1 Entry")},
		Exit:        []Action{recordAction("State 1 Exit")},
		Handle:      []Action{recordAction("State 1 Handle")},
		ParentState: &parentState,
	}
	state2 := State{
		Entry:       []Action{recordAction("State 2 Entry")},
		Exit:        []Action{recordAction("State 2 Exit")},
		Handle:      []Action{recordAction("State 2 Handle")},
		ParentState: &parentState2,
	}
	state3 := State{
		Entry:  []Action{recordAction("State 3 Entry")},
		Exit:   []Action{recordAction("State 3 Exit")},
		Handle: []Action{recordAction("State 3 Handle")},
	}

	states := []State{state1, state2, state3, parentState, parentState2}

	transitions := []Transition{
		{
			CurrentState: &state1,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 1 -> State 2 Transition")},
			NextState:    &state2,
		},
		{
			CurrentState: &state2,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 2 -> State 3 Transition")},
			NextState:    &state3,
		},
		{
			CurrentState: &state3,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 3 -> Parent State Transition")},
			NextState:    &state1,
		},
	}

	// Check for valid state machine
	sm, err := NewHierarchicalStateMachine(&state1, states, transitions)
	if err != nil {
		t.Fatalf("failed to initialize state machine: %v", err)
	}

	// Check current state
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &state1, sm.CurrentState)
	}

	// Check initial transition
	expectedActions := []string{"Parent State 2 Entry", "Parent State Entry", "State 1 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()
	HandleStateMachine(sm) // Transition from State 1 to State 2

	// Check current state
	if sm.CurrentState != &state2 {
		t.Errorf("Expected current state to be %v, got %v", &state2, sm.CurrentState)
	}

	// Check first transition
	expectedActions = []string{
		"Parent State 2 Handle", "Parent State Handle", "State 1 Handle",
		"State 1 Exit", "Parent State Exit", "State 1 -> State 2 Transition", "State 2 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()

	HandleStateMachine(sm) // Transition from State 2 to State 3

	// Check current state
	if sm.CurrentState != &state3 {
		t.Errorf("Expected current state to be %v, got %v", &state3, sm.CurrentState)
	}

	// Check second transition (State 2 -> State 3)
	expectedActions = []string{
		"Parent State 2 Handle", "State 2 Handle", "State 2 Exit", "Parent State 2 Exit",
		"State 2 -> State 3 Transition", "State 3 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()

	HandleStateMachine(sm) // Transition from State 3 to Parent State

	// Check current state
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &parentState, sm.CurrentState)
	}

	// Check third transition (State 3 -> Parent State)
	expectedActions = []string{
		"State 3 Handle", "State 3 Exit", "State 3 -> Parent State Transition",
		"Parent State 2 Entry", "Parent State Entry", "State 1 Entry"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}
}

func TestTransitionWithGuards(t *testing.T) {
	resetExecutedActions()

	state1 := State{}
	state2 := State{}

	canTransition := false // This will control if the transition should occur

	transitions := []Transition{
		{
			CurrentState: &state1,
			Event:        func() bool { return true },
			Guards:       []Predicate{func() bool { return canTransition }},
			Actions:      []Action{recordAction("State 1 -> State 2 Transition")},
			NextState:    &state2,
		},
	}

	sm, err := NewHierarchicalStateMachine(&state1, []State{state1, state2}, transitions)
	if err != nil {
		t.Fatalf("failed to initialize state machine: %v", err)
	}

	HandleStateMachine(sm) // Should not transition since canTransition is false
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &state1, sm.CurrentState)
	}

	canTransition = true
	HandleStateMachine(sm) // Should transition now
	if sm.CurrentState != &state2 {
		t.Errorf("Expected current state to be %v, got %v", &state2, sm.CurrentState)
	}
}

func TestMultipleActionsPerState(t *testing.T) {
	resetExecutedActions()

	state1 := State{
		Entry:  []Action{recordAction("State 1 Entry Action 1"), recordAction("State 1 Entry Action 2")},
		Exit:   []Action{recordAction("State 1 Exit Action 1"), recordAction("State 1 Exit Action 2")},
		Handle: []Action{recordAction("State 1 Handle Action 1"), recordAction("State 1 Handle Action 2")},
	}
	state2 := State{
		Entry:  []Action{recordAction("State 2 Entry Action 1"), recordAction("State 2 Entry Action 2")},
		Exit:   []Action{recordAction("State 2 Exit Action 1"), recordAction("State 2 Exit Action 2")},
		Handle: []Action{recordAction("State 2 Handle Action 1"), recordAction("State 2 Handle Action 2")},
	}

	states := []State{state1, state2}

	transitions := []Transition{
		{
			CurrentState: &state1,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 1 -> State 2 Transition")},
			NextState:    &state2,
		},
		{
			CurrentState: &state2,
			Event:        func() bool { return true },
			Guards:       nil,
			Actions:      []Action{recordAction("State 2 -> State 1 Transition")},
			NextState:    &state1,
		},
	}

	// Check for valid state machine
	sm, err := NewHierarchicalStateMachine(&state1, states, transitions)
	if err != nil {
		t.Fatalf("failed to initialize state machine: %v", err)
	}

	// Check initial state
	expectedActions := []string{"State 1 Entry Action 1", "State 1 Entry Action 2"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()
	HandleStateMachine(sm) // Transition from State 1 to State 2

	// Check current state
	if sm.CurrentState != &state2 {
		t.Errorf("Expected current state to be %v, got %v", &state2, sm.CurrentState)
	}

	// Check first transition
	expectedActions = []string{
		"State 1 Handle Action 1", "State 1 Handle Action 2",
		"State 1 Exit Action 1", "State 1 Exit Action 2",
		"State 1 -> State 2 Transition",
		"State 2 Entry Action 1", "State 2 Entry Action 2"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}

	resetExecutedActions()
	HandleStateMachine(sm) // Transition from State 2 to State 1

	// Check current state
	if sm.CurrentState != &state1 {
		t.Errorf("Expected current state to be %v, got %v", &state1, sm.CurrentState)
	}

	// Check second transition
	expectedActions = []string{
		"State 2 Handle Action 1", "State 2 Handle Action 2",
		"State 2 Exit Action 1", "State 2 Exit Action 2",
		"State 2 -> State 1 Transition",
		"State 1 Entry Action 1", "State 1 Entry Action 2"}
	if !reflect.DeepEqual(executedActions, expectedActions) {
		t.Errorf("expected actions %v, got %v", expectedActions, executedActions)
	}
}
