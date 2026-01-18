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
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/raft"
)

func TestNewFSM(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	if fsm == nil {
		t.Fatal("NewFSM() returned nil")
	}

	if fsm.logger != logger {
		t.Errorf("NewFSM().logger = %v, want %v", fsm.logger, logger)
	}

	if fsm.data == nil {
		t.Error("NewFSM().data is nil")
	}
}

func TestFSM_Apply(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	log := &raft.Log{
		Index: 1,
		Type:  raft.LogCommand,
		Data:  []byte("test data"),
	}

	result := fsm.Apply(log)

	if result == nil {
		t.Error("Apply() returned nil")
	}

	if result.(uint64) != 1 {
		t.Errorf("Apply() result = %v, want 1", result)
	}

	state := fsm.GetState()
	if state == nil {
		t.Error("Apply() did not update state")
	}
}

func TestFSM_Apply_Concurrent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index uint64) {
			defer wg.Done()
			log := &raft.Log{
				Index: index,
				Type:  raft.LogCommand,
				Data:  []byte("test data"),
			}
			fsm.Apply(log)
		}(uint64(i))
	}
	wg.Wait()

	state := fsm.GetState()
	if state == nil {
		t.Error("Apply() concurrent calls did not maintain state")
	}
}

func TestFSM_Snapshot(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	fsm.SetState("key1", "value1")
	fsm.SetState("key2", "value2")

	snapshot, err := fsm.Snapshot()

	if err != nil {
		t.Errorf("Snapshot() unexpected error: %v", err)
	}

	if snapshot == nil {
		t.Error("Snapshot() returned nil")
	}

	snapFSM, ok := snapshot.(*fsmSnapshot)
	if !ok {
		t.Error("Snapshot() did not return *fsmSnapshot")
	}

	if snapFSM.data == nil {
		t.Error("Snapshot() data is nil")
	}

	if len(snapFSM.data) != 2 {
		t.Errorf("Snapshot() data length = %d, want 2", len(snapFSM.data))
	}
}

func TestFSM_Restore(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	fsm.SetState("key1", "value1")
	fsm.SetState("key2", "value2")

	reader := io.NopCloser(nil)
	err := fsm.Restore(reader)

	if err != nil {
		t.Errorf("Restore() unexpected error: %v", err)
	}

	state := fsm.GetState()
	if state == nil {
		t.Error("Restore() did not clear state")
	}

	if len(state) != 0 {
		t.Errorf("Restore() state length = %d, want 0", len(state))
	}
}

func TestFsmSnapshot_Persist(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	snapshot := &fsmSnapshot{
		data:   map[string]string{"key": "value"},
		logger: logger,
	}

	sink := &testSnapshotSink{}
	err := snapshot.Persist(sink)

	if err != nil {
		t.Errorf("Persist() unexpected error: %v", err)
	}
}

func TestFsmSnapshot_Release(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	snapshot := &fsmSnapshot{
		data:   map[string]string{"key": "value"},
		logger: logger,
	}

	snapshot.Release()
}

func TestFSM_GetState(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	fsm.SetState("key1", "value1")
	fsm.SetState("key2", "value2")

	state := fsm.GetState()

	if state == nil {
		t.Fatal("GetState() returned nil")
	}

	if len(state) != 2 {
		t.Errorf("GetState() length = %d, want 2", len(state))
	}

	if state["key1"] != "value1" {
		t.Errorf("GetState()[key1] = %v, want value1", state["key1"])
	}

	if state["key2"] != "value2" {
		t.Errorf("GetState()[key2] = %v, want value2", state["key2"])
	}

	state["external"] = "modification"
	originalState := fsm.GetState()
	if _, ok := originalState["external"]; ok {
		t.Error("GetState() returned reference to internal state")
	}
}

func TestFSM_GetState_Concurrent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	fsm.SetState("key1", "value1")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fsm.GetState()
		}()
	}
	wg.Wait()
}

func TestFSM_SetState(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	fsm.SetState("key1", "value1")
	fsm.SetState("key2", "value2")
	fsm.SetState("key1", "updated")

	state := fsm.GetState()

	if len(state) != 2 {
		t.Errorf("SetState() state length = %d, want 2", len(state))
	}

	if state["key1"] != "updated" {
		t.Errorf("SetState()[key1] = %v, want updated", state["key1"])
	}

	if state["key2"] != "value2" {
		t.Errorf("SetState()[key2] = %v, want value2", state["key2"])
	}
}

func TestFSM_SetState_Concurrent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			fsm.SetState(string(rune('a'+index%26)), "value")
		}(i)
	}
	wg.Wait()

	state := fsm.GetState()
	if state == nil {
		t.Error("SetState() concurrent calls did not maintain state")
	}
}

func TestFSM_GetStateValue(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	fsm.SetState("key1", "value1")
	fsm.SetState("key2", "value2")

	value, exists := fsm.GetStateValue("key1")

	if !exists {
		t.Error("GetStateValue() exists = false, want true")
	}

	if value != "value1" {
		t.Errorf("GetStateValue() value = %v, want value1", value)
	}

	value, exists = fsm.GetStateValue("nonexistent")
	if exists {
		t.Error("GetStateValue() exists = true for nonexistent key, want false")
	}

	if value != "" {
		t.Errorf("GetStateValue() value for nonexistent key = %v, want empty string", value)
	}
}

func TestFSM_GetStateValue_Concurrent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fsm := NewFSM(logger)

	fsm.SetState("key1", "value1")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fsm.GetStateValue("key1")
		}()
	}
	wg.Wait()
}

type testSnapshotSink struct{}

func (s *testSnapshotSink) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (s *testSnapshotSink) Close() error {
	return nil
}

func (s *testSnapshotSink) ID() string {
	return "test-snapshot"
}

func (s *testSnapshotSink) Cancel() error {
	return nil
}
