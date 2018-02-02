package consul

import (
	"fmt"
	"net"

	"github.com/hashicorp/consul/consul/agent"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

// Operator endpoint is used to perform low-level operator tasks for Consul.
type Operator struct {
	srv *Server
}

// RaftGetConfiguration is used to retrieve the current Raft configuration.
func (op *Operator) RaftGetConfiguration(args *structs.DCSpecificRequest, reply *structs.RaftConfigurationResponse) error {
	if done, err := op.srv.forward("Operator.RaftGetConfiguration", args, args, reply); done {
		return err
	}

	// This action requires operator read access.
	acl, err := op.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && !acl.OperatorRead() {
		return permissionDeniedErr
	}

	// We can't fetch the leader and the configuration atomically with
	// the current Raft API.
	future := op.srv.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return err
	}

	// Index the Consul information about the servers.
	serverMap := make(map[raft.ServerAddress]serf.Member)
	for _, member := range op.srv.serfLAN.Members() {
		valid, parts := agent.IsConsulServer(member)
		if !valid {
			continue
		}

		addr := (&net.TCPAddr{IP: member.Addr, Port: parts.Port}).String()
		serverMap[raft.ServerAddress(addr)] = member
	}

	// Fill out the reply.
	leader := op.srv.raft.Leader()
	reply.Index = future.Index()
	for _, server := range future.Configuration().Servers {
		node := "(unknown)"
		if member, ok := serverMap[server.Address]; ok {
			node = member.Name
		}

		entry := &structs.RaftServer{
			ID:      server.ID,
			Node:    node,
			Address: server.Address,
			Leader:  server.Address == leader,
			Voter:   server.Suffrage == raft.Voter,
		}
		reply.Servers = append(reply.Servers, entry)
	}
	return nil
}

// RaftRemovePeerByAddress is used to kick a stale peer (one that it in the Raft
// quorum but no longer known to Serf or the catalog) by address in the form of
// "IP:port". The reply argument is not used, but it required to fulfill the RPC
// interface.
func (op *Operator) RaftRemovePeerByAddress(args *structs.RaftPeerByAddressRequest, reply *struct{}) error {
	if done, err := op.srv.forward("Operator.RaftRemovePeerByAddress", args, args, reply); done {
		return err
	}

	// This is a super dangerous operation that requires operator write
	// access.
	acl, err := op.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && !acl.OperatorWrite() {
		return permissionDeniedErr
	}

	// Since this is an operation designed for humans to use, we will return
	// an error if the supplied address isn't among the peers since it's
	// likely they screwed up.
	{
		future := op.srv.raft.GetConfiguration()
		if err := future.Error(); err != nil {
			return err
		}
		for _, s := range future.Configuration().Servers {
			if s.Address == args.Address {
				goto REMOVE
			}
		}
		return fmt.Errorf("address %q was not found in the Raft configuration",
			args.Address)
	}

REMOVE:
	// The Raft library itself will prevent various forms of foot-shooting,
	// like making a configuration with no voters. Some consideration was
	// given here to adding more checks, but it was decided to make this as
	// low-level and direct as possible. We've got ACL coverage to lock this
	// down, and if you are an operator, it's assumed you know what you are
	// doing if you are calling this. If you remove a peer that's known to
	// Serf, for example, it will come back when the leader does a reconcile
	// pass.
	future := op.srv.raft.RemovePeer(args.Address)
	if err := future.Error(); err != nil {
		op.srv.logger.Printf("[WARN] consul.operator: Failed to remove Raft peer %q: %v",
			args.Address, err)
		return err
	}

	op.srv.logger.Printf("[WARN] consul.operator: Removed Raft peer %q", args.Address)
	return nil
}

// AutopilotGetConfiguration is used to retrieve the current Autopilot configuration.
func (op *Operator) AutopilotGetConfiguration(args *structs.DCSpecificRequest, reply *structs.AutopilotConfig) error {
	if done, err := op.srv.forward("Operator.AutopilotGetConfiguration", args, args, reply); done {
		return err
	}

	// This action requires operator read access.
	acl, err := op.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && !acl.OperatorRead() {
		return permissionDeniedErr
	}

	state := op.srv.fsm.State()
	_, config, err := state.AutopilotConfig()
	if err != nil {
		return err
	}

	*reply = *config

	return nil
}

// AutopilotSetConfiguration is used to set the current Autopilot configuration.
func (op *Operator) AutopilotSetConfiguration(args *structs.AutopilotSetConfigRequest, reply *bool) error {
	if done, err := op.srv.forward("Operator.AutopilotSetConfiguration", args, args, reply); done {
		return err
	}

	// This action requires operator write access.
	acl, err := op.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && !acl.OperatorWrite() {
		return permissionDeniedErr
	}

	// Apply the update
	resp, err := op.srv.raftApply(structs.AutopilotRequestType, args)
	if err != nil {
		op.srv.logger.Printf("[ERR] consul.operator: Apply failed: %v", err)
		return err
	}
	if respErr, ok := resp.(error); ok {
		return respErr
	}

	// Check if the return type is a bool.
	if respBool, ok := resp.(bool); ok {
		*reply = respBool
	}
	return nil
}

// ServerHealth is used to get the current health of the servers.
func (op *Operator) ServerHealth(args *structs.DCSpecificRequest, reply *structs.OperatorHealthReply) error {
	// This must be sent to the leader, so we fix the args since we are
	// re-using a structure where we don't support all the options.
	args.RequireConsistent = true
	args.AllowStale = false
	if done, err := op.srv.forward("Operator.ServerHealth", args, args, reply); done {
		return err
	}

	// This action requires operator read access.
	acl, err := op.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && !acl.OperatorRead() {
		return permissionDeniedErr
	}

	// Exit early if the min Raft version is too low
	minRaftProtocol, err := ServerMinRaftProtocol(op.srv.LANMembers())
	if err != nil {
		return fmt.Errorf("error getting server raft protocol versions: %s", err)
	}
	if minRaftProtocol < 3 {
		return fmt.Errorf("all servers must have raft_protocol set to 3 or higher to use this endpoint")
	}

	var status structs.OperatorHealthReply
	future := op.srv.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return err
	}

	healthyCount := 0
	servers := future.Configuration().Servers
	for _, s := range servers {
		health := op.srv.getServerHealth(string(s.ID))
		if health != nil {
			if health.Healthy {
				healthyCount++
			}
			status.Servers = append(status.Servers, *health)
		}
	}
	status.Healthy = healthyCount == len(servers)

	// If we have extra healthy servers, set FailureTolerance
	if healthyCount > len(servers)/2+1 {
		status.FailureTolerance = healthyCount - (len(servers)/2 + 1)
	}

	*reply = status

	return nil
}
