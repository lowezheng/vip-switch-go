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

package raft

import (
	"io"
	"log/slog"
	"sync"

	"github.com/hashicorp/raft"
)

// FSM implements the Raft finite state machine
type FSM struct {
	logger *slog.Logger
	mu     sync.RWMutex
	data   map[string]string
}

// NewFSM creates a new FSM instance
func NewFSM(logger *slog.Logger) *FSM {
	return &FSM{
		logger: logger,
		data:   make(map[string]string),
	}
}

// Apply applies a Raft log entry to the FSM
func (f *FSM) Apply(log *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// For VIP-Switch, we don't need complex state machine operations
	// The actual VIP management is done through hooks triggered by leadership changes
	f.logger.Debug("FSM Apply", "index", log.Index, "type", log.Type)

	// Return index as response
	return log.Index
}

// Snapshot creates a snapshot of the FSM state
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	f.logger.Debug("FSM Snapshot")

	// Create a snapshot of our data
	snapshotData := make(map[string]string)
	for k, v := range f.data {
		snapshotData[k] = v
	}

	return &fsmSnapshot{
		data:   snapshotData,
		logger: f.logger,
	}, nil
}

// Restore restores the FSM from a snapshot
func (f *FSM) Restore(rc io.ReadCloser) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.logger.Debug("FSM Restore")

	// In our simple FSM, we just clear the data
	// In a more complex implementation, we would decode the snapshot data
	f.data = make(map[string]string)

	return nil
}

// fsmSnapshot represents a snapshot of the FSM
type fsmSnapshot struct {
	data   map[string]string
	logger *slog.Logger
}

// Persist writes the snapshot to the given sink
func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	f.logger.Debug("fsmSnapshot Persist")

	// For simplicity, we don't persist any data
	// In a real implementation, you would serialize f.data and write it to sink
	return nil
}

// Release releases any resources held by the snapshot
func (f *fsmSnapshot) Release() {
	f.logger.Debug("fsmSnapshot Release")
}

// GetState returns the current state of the FSM
func (f *FSM) GetState() map[string]string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	state := make(map[string]string)
	for k, v := range f.data {
		state[k] = v
	}
	return state
}

// SetState sets a value in the FSM state
func (f *FSM) SetState(key, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.data[key] = value
}

// GetStateValue returns a value from the FSM state
func (f *FSM) GetStateValue(key string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	value, exists := f.data[key]
	return value, exists
}
