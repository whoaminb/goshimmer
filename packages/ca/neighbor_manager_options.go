package ca

var DEFAULT_NEIGHBOR_MANAGER_OPTIONS = &NeighborManagerOptions{
	maxNeighborChains: 8,
}

type NeighborManagerOptions struct {
	maxNeighborChains int
}

func (options NeighborManagerOptions) Override(optionalOptions ...NeighborManagerOption) *NeighborManagerOptions {
	result := &options
	for _, option := range optionalOptions {
		option(result)
	}

	return result
}

type NeighborManagerOption func(*NeighborManagerOptions)
