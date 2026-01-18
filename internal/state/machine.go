// Copyright 2026 lowezheng
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package state

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"vip-switch-go/internal/hook"
	"vip-switch-go/internal/raft"
)

// State represents the node state
type State int

const (
	StateReady   State = iota
	StateSlave   State = iota
	StateMaster  State = iota
	StateDestroy State = iota
)

func (s State) String() string {
	switch s {
	case StateReady:
		return "Ready"
	case StateSlave:
		return "Slave"
	case StateMaster:
		return "Master"
	case StateDestroy:
		return "Destroy"
	default:
		return "Unknown"
	}
}

// Machine manages state transitions
type Machine struct {
	currentState    State
	previousState   State
	nodeID          string
	hookSystem      *hook.System
	raftNode        *raft.Node
	logger          *slog.Logger
	mu              sync.RWMutex
	shutdown        chan struct{}
	lastStateChange time.Time
	debounceDelay   time.Duration
}

// NewMachine creates a new state machine
func NewMachine(hookSystem *hook.System, nodeID string, logger *slog.Logger) *Machine {
	return &Machine{
		currentState:    StateReady,
		previousState:   StateReady,
		nodeID:          nodeID,
		hookSystem:      hookSystem,
		logger:          logger,
		shutdown:        make(chan struct{}),
		debounceDelay:   2 * time.Second,
		lastStateChange: time.Now(),
	}
}

// SetRaftNode sets the Raft node reference
func (m *Machine) SetRaftNode(node *raft.Node) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.raftNode = node
}

// GetCurrentState returns the current state
func (m *Machine) GetCurrentState() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

// Start begins monitoring Raft leadership changes
func (m *Machine) Start(ctx context.Context) error {
	m.logger.Info("Starting state machine", "state", m.currentState.String())

	go m.monitorLeadership(ctx)

	return nil
}

// monitorLeadership monitors Raft leadership changes and triggers state transitions
func (m *Machine) monitorLeadership(ctx context.Context) {
	if m.raftNode == nil {
		m.logger.Warn("Raft node not set, cannot monitor leadership")
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Stopping state machine")
			return
		case <-m.shutdown:
			m.logger.Info("Stopping state machine (shutdown)")
			return
		case isLeader := <-m.raftNode.LeaderCh():
			m.handleLeadershipChange(isLeader, ctx)
		case <-ticker.C:
			m.checkRaftState(ctx)
		}
	}
}

// checkRaftState periodically checks Raft state
func (m *Machine) checkRaftState(ctx context.Context) {
	leader := m.raftNode.Leader()

	m.mu.Lock()
	defer m.mu.Unlock()

	var newState State
	if leader == "" {
		newState = StateSlave
	} else if m.raftNode.IsLeader() {
		newState = StateMaster
	} else {
		newState = StateSlave
	}

	if newState != m.currentState {
		m.transition(newState, ctx)
	}
}

// handleLeadershipChange handles leadership change events
func (m *Machine) handleLeadershipChange(isLeader bool, ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var newState State
	if isLeader {
		newState = StateMaster
	} else {
		newState = StateSlave
	}

	if newState != m.currentState {
		m.transition(newState, ctx)
	}
}

// transition performs state transition with debounce
func (m *Machine) transition(newState State, ctx context.Context) {
	timeSinceLastChange := time.Since(m.lastStateChange)

	if timeSinceLastChange < m.debounceDelay && m.currentState != StateReady {
		m.logger.Debug("Debouncing state transition",
			"from", m.currentState.String(),
			"to", newState.String(),
			"elapsed", timeSinceLastChange,
		)
		return
	}

	m.logger.Info("State transition",
		"from", m.currentState.String(),
		"to", newState.String(),
	)

	m.previousState = m.currentState
	m.currentState = newState
	m.lastStateChange = time.Now()

	if err := m.executeHookForState(newState, ctx); err != nil {
		m.logger.Error("Hook execution failed during state transition",
			"state", newState.String(),
			"error", err,
		)
	}
}

// executeHookForState executes the appropriate hook for a state
func (m *Machine) executeHookForState(state State, ctx context.Context) error {
	var eventType string

	switch state {
	case StateReady:
		eventType = "ToReady"
	case StateMaster:
		eventType = "ToMaster"
	case StateSlave:
		eventType = "ToSlave"
	case StateDestroy:
		eventType = "ToDestroy"
	default:
		return nil
	}

	return m.hookSystem.ExecuteHook(ctx, eventType)
}

// Shutdown gracefully shuts down the state machine
func (m *Machine) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.shutdown:
		return nil
	default:
	}

	m.logger.Info("Shutting down state machine", "current_state", m.currentState.String())

	close(m.shutdown)

	if err := m.executeHookForState(StateDestroy, ctx); err != nil {
		m.logger.Error("Destroy hook failed", "error", err)
	}

	return nil
}
