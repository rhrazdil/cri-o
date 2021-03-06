package server

import (
	"github.com/cri-o/cri-o/internal/config/node"
	"github.com/cri-o/cri-o/server/cri/types"
	"github.com/gogo/protobuf/proto"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// UpdateContainerResources updates ContainerConfig of the container.
func (s *Server) UpdateContainerResources(ctx context.Context, req *types.UpdateContainerResourcesRequest) error {
	c, err := s.GetContainerFromShortID(req.ContainerID)
	if err != nil {
		return err
	}

	if err := c.IsAlive(); err != nil {
		return errors.Errorf("container is not created or running: %v", err)
	}

	if req.Linux != nil {
		resources := toOCIResources(req.Linux)
		if err := s.Runtime().UpdateContainer(c, resources); err != nil {
			return err
		}

		// update memory store with updated resources
		s.UpdateContainerLinuxResources(c, resources)
	}

	return nil
}

// toOCIResources converts CRI resource constraints to OCI.
func toOCIResources(r *types.LinuxContainerResources) *rspec.LinuxResources {
	var swap int64
	memory := r.MemoryLimitInBytes
	if node.CgroupHasMemorySwap() {
		swap = memory
	}
	return &rspec.LinuxResources{
		CPU: &rspec.LinuxCPU{
			Shares: proto.Uint64(uint64(r.CPUShares)),
			Quota:  proto.Int64(r.CPUQuota),
			Period: proto.Uint64(uint64(r.CPUPeriod)),
			Cpus:   r.CPUsetCPUs,
			Mems:   r.CPUsetMems,
		},
		Memory: &rspec.LinuxMemory{
			Limit: proto.Int64(memory),
			Swap:  proto.Int64(swap),
		},
		// TODO(runcom): OOMScoreAdj is missing
	}
}
