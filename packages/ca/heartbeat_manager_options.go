package ca

type HeartbeatManagerOptions struct {
	maxNeighborChains int
}

func (options HeartbeatManagerOptions) Override(optionalOptions ...HeartbeatManagerOption) *HeartbeatManagerOptions {
	result := &options
	for _, option := range optionalOptions {
		option(result)
	}

	return result
}

type HeartbeatManagerOption func(*HeartbeatManagerOptions)

func MaxNeighborChains(maxNeighborChains int) HeartbeatManagerOption {
	return func(args *HeartbeatManagerOptions) {
		args.maxNeighborChains = maxNeighborChains
	}
}

var DEFAULT_OPTIONS = &HeartbeatManagerOptions{
	maxNeighborChains: 8,
}
