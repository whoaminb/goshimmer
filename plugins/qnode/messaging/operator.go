package messaging

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
)

// An abstract interface of the consensus operator
// Each smart contract is run by one operator
// For the messaging module it is the main interface with the smart contract
type SCOperator interface {
	// smart contract ID
	SContractID() *hashing.HashValue
	// size of the committee
	CommitteeSize() uint16
	// quorum, a minimum amount of peers to be able to operate the smart contract
	Quorum() uint16
	// index of the peer in the committee
	PeerIndex() uint16
	// list of all node addresses of peers.
	// Consistent with the PeerIndex() i.e. using peer index in the array gives address of that peer
	PeerLocations() []*registry.PortAddr
	// called each time new state update transaction comes to the smart contract
	// it is guaranteed that transaction referenced in the parameter contains state block
	ReceiveStateUpdate(*sc.StateUpdateMsg)
	// called each time new request comes for the smart contract
	// if transaction contains several request blocks, it is called for each with respective
	// request index in the parameter
	ReceiveRequest(*sc.RequestRef)
	// thread safe way to determine if sc operator is dismissed
	IsDismissed() bool
}
