package network

import (
	"github.com/iotaledger/goshimmer/plugins/autopeering"
	"github.com/iotaledger/hive.go/autopeering/peer/service"
)

func GetKnownPeers() []string {
	peerAddresses := []string{}
	for _, peer := range autopeering.Discovery.GetVerifiedPeers() {
		fpcService := peer.Services().Get(service.FPCKey)
		if fpcService != nil {
			peerAddresses = append(peerAddresses, fpcService.String())
		}
	}
	return peerAddresses
}
