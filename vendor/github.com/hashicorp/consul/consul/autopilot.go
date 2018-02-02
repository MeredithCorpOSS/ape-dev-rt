package consul

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/consul/consul/agent"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

// AutopilotPolicy is the interface for the Autopilot mechanism
type AutopilotPolicy interface {
	// PromoteNonVoters defines the handling of non-voting servers
	PromoteNonVoters(*structs.AutopilotConfig) error
}

func (s *Server) startAutopilot() {
	s.autopilotShutdownCh = make(chan struct{})
	s.autopilotWaitGroup = sync.WaitGroup{}
	s.autopilotWaitGroup.Add(1)

	go s.autopilotLoop()
}

func (s *Server) stopAutopilot() {
	close(s.autopilotShutdownCh)
	s.autopilotWaitGroup.Wait()
}

// autopilotLoop periodically looks for nonvoting servers to promote and dead servers to remove.
func (s *Server) autopilotLoop() {
	defer s.autopilotWaitGroup.Done()

	// Monitor server health until shutdown
	ticker := time.NewTicker(s.config.AutopilotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.autopilotShutdownCh:
			return
		case <-ticker.C:
			state := s.fsm.State()
			_, autopilotConf, err := state.AutopilotConfig()
			if err != nil {
				s.logger.Printf("[ERR] consul: error retrieving autopilot config: %s", err)
				break
			}

			if err := s.autopilotPolicy.PromoteNonVoters(autopilotConf); err != nil {
				s.logger.Printf("[ERR] consul: error checking for non-voters to promote: %s", err)
			}

			if err := s.pruneDeadServers(); err != nil {
				s.logger.Printf("[ERR] consul: error checking for dead servers to remove: %s", err)
			}
		case <-s.autopilotRemoveDeadCh:
			if err := s.pruneDeadServers(); err != nil {
				s.logger.Printf("[ERR] consul: error checking for dead servers to remove: %s", err)
			}
		}
	}
}

// pruneDeadServers removes up to numPeers/2 failed servers
func (s *Server) pruneDeadServers() error {
	state := s.fsm.State()
	_, autopilotConf, err := state.AutopilotConfig()
	if err != nil {
		return err
	}

	// Find any failed servers
	var failed []string
	if autopilotConf.CleanupDeadServers {
		for _, member := range s.serfLAN.Members() {
			valid, _ := agent.IsConsulServer(member)
			if valid && member.Status == serf.StatusFailed {
				failed = append(failed, member.Name)
			}
		}
	}

	// Nothing to remove, return early
	if len(failed) == 0 {
		return nil
	}

	peers, err := s.numPeers()
	if err != nil {
		return err
	}

	// Only do removals if a minority of servers will be affected
	if len(failed) < peers/2 {
		for _, server := range failed {
			s.logger.Printf("[INFO] consul: Attempting removal of failed server: %v", server)
			go s.serfLAN.RemoveFailedNode(server)
		}
	} else {
		s.logger.Printf("[ERR] consul: Failed to remove dead servers: too many dead servers: %d/%d", len(failed), peers)
	}

	return nil
}

// BasicAutopilot defines a policy for promoting non-voting servers in a way
// that maintains an odd-numbered voter count.
type BasicAutopilot struct {
	server *Server
}

// PromoteNonVoters promotes eligible non-voting servers to voters.
func (b *BasicAutopilot) PromoteNonVoters(autopilotConf *structs.AutopilotConfig) error {
	minRaftProtocol, err := ServerMinRaftProtocol(b.server.LANMembers())
	if err != nil {
		return fmt.Errorf("error getting server raft protocol versions: %s", err)
	}

	// If we don't meet the minimum version for non-voter features, bail early
	if minRaftProtocol < 3 {
		return nil
	}

	future := b.server.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return fmt.Errorf("failed to get raft configuration: %v", err)
	}

	var promotions []raft.Server
	raftServers := future.Configuration().Servers
	voterCount := 0
	for _, server := range raftServers {
		// If this server has been stable and passing for long enough, promote it to a voter
		if server.Suffrage == raft.Nonvoter {
			health := b.server.getServerHealth(string(server.ID))
			if health.IsStable(time.Now(), autopilotConf) {
				promotions = append(promotions, server)
			}
		} else {
			voterCount++
		}
	}

	// Exit early if there's nothing to promote
	if len(promotions) == 0 {
		return nil
	}

	// If there's currently an even number of servers, we can promote the first server in the list
	// to get to an odd-sized quorum
	newServers := false
	if voterCount%2 == 0 {
		addFuture := b.server.raft.AddVoter(promotions[0].ID, promotions[0].Address, 0, 0)
		if err := addFuture.Error(); err != nil {
			return fmt.Errorf("failed to add raft peer: %v", err)
		}
		promotions = promotions[1:]
		newServers = true
	}

	// Promote remaining servers in twos to maintain an odd quorum size
	for i := 0; i < len(promotions)-1; i += 2 {
		addFirst := b.server.raft.AddVoter(promotions[i].ID, promotions[i].Address, 0, 0)
		if err := addFirst.Error(); err != nil {
			return fmt.Errorf("failed to add raft peer: %v", err)
		}
		addSecond := b.server.raft.AddVoter(promotions[i+1].ID, promotions[i+1].Address, 0, 0)
		if err := addSecond.Error(); err != nil {
			return fmt.Errorf("failed to add raft peer: %v", err)
		}
		newServers = true
	}

	// If we added a new server, trigger a check to remove dead servers
	if newServers {
		select {
		case b.server.autopilotRemoveDeadCh <- struct{}{}:
		default:
		}
	}

	return nil
}

// serverHealthLoop monitors the health of the servers in the cluster
func (s *Server) serverHealthLoop() {
	// Monitor server health until shutdown
	ticker := time.NewTicker(s.config.ServerHealthInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdownCh:
			return
		case <-ticker.C:
			serverHealths := make(map[string]*structs.ServerHealth)

			// Don't do anything if the min Raft version is too low
			minRaftProtocol, err := ServerMinRaftProtocol(s.LANMembers())
			if err != nil {
				s.logger.Printf("[ERR] consul: error getting server raft protocol versions: %s", err)
				break
			}
			if minRaftProtocol < 3 {
				break
			}

			state := s.fsm.State()
			_, autopilotConf, err := state.AutopilotConfig()
			if err != nil {
				s.logger.Printf("[ERR] consul: error retrieving autopilot config: %s", err)
				break
			}
			// Bail early if autopilot config hasn't been initialized yet
			if autopilotConf == nil {
				break
			}

			// Build an updated map of server healths
			for _, member := range s.LANMembers() {
				if member.Status == serf.StatusLeft {
					continue
				}

				valid, parts := agent.IsConsulServer(member)
				if valid {
					health, err := s.queryServerHealth(member, parts, autopilotConf)
					if err != nil {
						s.logger.Printf("[ERR] consul: error fetching server health: %s", err)
						serverHealths[parts.ID] = &structs.ServerHealth{
							ID:      parts.ID,
							Name:    parts.Name,
							Healthy: false,
						}
					} else {
						serverHealths[parts.ID] = health
					}
				}
			}

			s.serverHealthLock.Lock()
			s.serverHealths = serverHealths
			s.serverHealthLock.Unlock()
		}
	}
}

// queryServerHealth fetches the raft stats for the given server and uses them
// to update its ServerHealth
func (s *Server) queryServerHealth(member serf.Member, server *agent.Server,
	autopilotConf *structs.AutopilotConfig) (*structs.ServerHealth, error) {
	stats, err := s.getServerStats(server)
	if err != nil {
		return nil, fmt.Errorf("error getting raft stats: %s", err)
	}

	health := &structs.ServerHealth{
		ID:          server.ID,
		Name:        server.Name,
		SerfStatus:  member.Status,
		LastContact: -1,
		LastTerm:    stats.LastTerm,
		LastIndex:   stats.LastIndex,
	}

	if stats.LastContact != "never" {
		health.LastContact, err = time.ParseDuration(stats.LastContact)
		if err != nil {
			return nil, fmt.Errorf("error parsing last_contact duration: %s", err)
		}
	}

	// Set LastContact to 0 for the leader
	if s.config.NodeName == member.Name {
		health.LastContact = 0
	}

	lastTerm, err := strconv.ParseUint(s.raft.Stats()["last_log_term"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing last_log_term: %s", err)
	}
	health.Healthy = health.IsHealthy(lastTerm, s.raft.LastIndex(), autopilotConf)

	// If this is a new server or the health changed, reset StableSince
	lastHealth := s.getServerHealth(server.ID)
	if lastHealth == nil || lastHealth.Healthy != health.Healthy {
		health.StableSince = time.Now()
	} else {
		health.StableSince = lastHealth.StableSince
	}

	return health, nil
}

func (s *Server) getServerHealth(id string) *structs.ServerHealth {
	s.serverHealthLock.RLock()
	defer s.serverHealthLock.RUnlock()
	h, ok := s.serverHealths[id]
	if !ok {
		return nil
	}
	return h
}

func (s *Server) getServerStats(server *agent.Server) (structs.ServerStats, error) {
	var args struct{}
	var reply structs.ServerStats
	err := s.connPool.RPC(s.config.Datacenter, server.Addr, server.Version, "Status.RaftStats", &args, &reply)
	return reply, err
}
