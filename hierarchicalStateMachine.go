package hierarchicalStateMachine

import "fmt"

const MaxStates = 10 // MaxStates is used to create fixed-size arrays to avoid heap allocation

type StateName string
type Predicate func() bool
type Action func()

type State struct {
	Name        StateName
	Entry       []Action
	Exit        []Action
	Handle      []Action
	ParentState *State
}

type Transition struct {
	CurrentState *State
	Event        Predicate
	Guards       []Predicate
	Actions      []Action
	NextState    *State
}

type HierarchicalStateMachine struct {
	CurrentState *State
	states       []State
	transitions  []Transition
}

func NewHierarchicalStateMachine(initialState *State, states []State, transitions []Transition) (*HierarchicalStateMachine, error) {
	if len(states) > MaxStates {
		return nil, fmt.Errorf("too many states declared: %d. max allowed is %d", len(states), MaxStates)
	}
	sm := &HierarchicalStateMachine{
		CurrentState: initialState,
		states:       states,
		transitions:  transitions,
	}

	// Execute all entry actions in current state hierarchy
	executeActionsInHierarchy(sm.CurrentState, func(s *State) []Action { return s.Entry })

	return sm, nil
}

// HandleStateMachine processes state transitions and executes actions accordingly
func HandleStateMachine(sm *HierarchicalStateMachine) {
	// Execute all handlers in current state hierarchy
	executeActionsInHierarchy(sm.CurrentState, func(s *State) []Action { return s.Handle })

	for _, transition := range sm.transitions {
		if sm.CurrentState == transition.CurrentState {
			if !transition.Event() {
				continue
			}

			guardsPassed := true
			for _, guard := range transition.Guards {
				if !guard() {
					guardsPassed = false
					break
				}
			}
			if !guardsPassed {
				continue
			}

			executeTransitionActions(transition)
			sm.CurrentState = transition.NextState
			break
		}
	}
}

func executeActions(actions []Action) {
	for _, action := range actions {
		action()
	}
}

// Parent actions are executed first
func executeActionsInHierarchy(state *State, actions func(s *State) []Action) {
	if state == nil {
		return
	}
	executeActionsInHierarchy(state.ParentState, actions)
	executeActions(actions(state))
}

func executeTransitionActions(transition Transition) {
	commonAncestor := findCommonAncestor(transition.CurrentState, transition.NextState)
	exitToCommonAncestor(transition.CurrentState, commonAncestor)
	executeActions(transition.Actions)
	enterFromCommonAncestor(transition.NextState, commonAncestor)
}

// Returns the deepest common ancestor of the two states
func findCommonAncestor(state1, state2 *State) *State {
	var visited [MaxStates]*State
	visitedCount := 0

	for state1 != nil {
		visited[visitedCount] = state1
		visitedCount++
		state1 = state1.ParentState
	}

	for state2 != nil {
		for i := 0; i < visitedCount; i++ {
			if state2 == visited[i] {
				return state2
			}
		}
		state2 = state2.ParentState
	}

	return nil
}

// Executes exit actions up to the common ancestor
func exitToCommonAncestor(state *State, commonAncestor *State) {
	for state != commonAncestor {
		executeActions(state.Exit)
		state = state.ParentState
	}
}

// Executes entry actions from the common ancestor
func enterFromCommonAncestor(state *State, commonAncestor *State) {

	var stack [MaxStates]*State
	stackCount := 0

	for state != commonAncestor {
		stack[stackCount] = state
		stackCount++
		state = state.ParentState
	}

	for i := stackCount - 1; i >= 0; i-- {
		executeActions(stack[i].Entry)
	}
}
