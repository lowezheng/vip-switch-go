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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"vip-switch-go/internal/config"
)

type Node struct {
	raftInstance *raft.Raft
	config       *config.Config
	fsm          *FSM
	logger       *slog.Logger
	shutdown     bool
	shutdownLock sync.RWMutex
}

func NewNode(cfg *config.Config, fsm *FSM, logger *slog.Logger) (*Node, error) {
	if err := os.MkdirAll(cfg.Node.DataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	store, err := raftboltdb.NewBoltStore(filepath.Join(cfg.Node.DataDir, "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create BoltDB store: %w", err)
	}

	snapshots, err := raft.NewFileSnapshotStore(cfg.Node.DataDir, 1, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %w", err)
	}

	transport, err := NewTCPTransport(cfg.Node.RaftAddr, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(cfg.Node.ID)
	raftCfg.SnapshotInterval = 30 * time.Second
	raftCfg.SnapshotThreshold = 2
	raftCfg.HeartbeatTimeout = 1 * time.Second
	raftCfg.ElectionTimeout = 1 * time.Second
	raftCfg.LeaderLeaseTimeout = 500 * time.Millisecond
	raftCfg.Logger = NewRaftLogger(logger)

	raftInstance, err := raft.NewRaft(
		raftCfg,
		fsm,
		store,
		store,
		snapshots,
		transport,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Raft instance: %w", err)
	}

	node := &Node{
		raftInstance: raftInstance,
		config:       cfg,
		fsm:          fsm,
		logger:       logger,
	}

	return node, nil
}

func (n *Node) Start() error {
	n.logger.Info("Starting Raft node",
		"node_id", n.config.Node.ID,
		"raft_addr", n.config.Node.RaftAddr,
	)

	futureConfig := n.raftInstance.GetConfiguration()
	if err := futureConfig.Error(); err != nil {
		n.logger.Warn("Failed to get configuration", "error", err)
		return err
	}

	n.logger.Info("Bootstrapping cluster")
	bootstrapConfig := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(n.config.Node.ID),
				Address: raft.ServerAddress(n.config.Node.RaftAddr),
			},
		},
	}
	if err := n.raftInstance.BootstrapCluster(bootstrapConfig).Error(); err != nil {
		if err != raft.ErrCantBootstrap {
			return fmt.Errorf("failed to bootstrap cluster: %w", err)
		}
	}

	peers := n.config.GetClusterPeers()
	if len(peers) > 0 {
		n.logger.Info("Joining cluster peers", "peers", peers)
		for _, peerAddr := range peers {
			var peerID string
			for _, clusterNode := range n.config.Cluster.Nodes {
				if clusterNode.Addr == peerAddr && clusterNode.ID != n.config.Node.ID {
					peerID = clusterNode.ID
					break
				}
			}

			if peerID == "" {
				n.logger.Warn("Could not find node ID for peer address", "addr", peerAddr)
				continue
			}

			future := n.raftInstance.AddVoter(
				raft.ServerID(peerID),
				raft.ServerAddress(peerAddr),
				0,
				0,
			)
			if err := future.Error(); err != nil {
				n.logger.Debug("AddVoter failed", "peer", peerAddr, "error", err)
			} else {
				n.logger.Info("Successfully joined cluster", "peer", peerAddr)
			}
		}
	}

	return nil
}

func (n *Node) LeaderCh() <-chan bool {
	return n.raftInstance.LeaderCh()
}

func (n *Node) State() raft.RaftState {
	return n.raftInstance.State()
}

func (n *Node) Leader() string {
	return string(n.raftInstance.Leader())
}

func (n *Node) Stats() map[string]string {
	return n.raftInstance.Stats()
}

func (n *Node) Shutdown() error {
	n.shutdownLock.Lock()
	defer n.shutdownLock.Unlock()

	if n.shutdown {
		return nil
	}

	n.shutdown = true
	n.logger.Info("Shutting down Raft node")

	future := n.raftInstance.Shutdown()
	if err := future.Error(); err != nil {
		n.logger.Error("Error during Raft shutdown", "error", err)
		return err
	}

	n.logger.Info("Raft node shutdown complete")
	return nil
}

func (n *Node) IsLeader() bool {
	return n.raftInstance.State() == raft.Leader
}

func (n *Node) Apply(cmd []byte, timeout time.Duration) raft.ApplyFuture {
	return n.raftInstance.Apply(cmd, timeout)
}
