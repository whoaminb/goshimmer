package social_consensus

import (
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type Node struct {
	id         string
	reputation uint64
	elderMask  ElderMask
}

func NewNode(id string) *Node {
	return &Node{
		id: id,
	}
}

func (node *Node) GetElderMask() ElderMask {
	return node.elderMask
}

func (node *Node) SetElderMask(mask ElderMask) {
	node.elderMask = mask
}

func (node *Node) IsElder() bool {
	return node.elderMask != 0
}

func (node *Node) GetId() string {
	return node.id
}

func (node *Node) GetReputation() uint64 {
	return node.reputation
}

func (node *Node) SetReputation(reputation uint64) {
	node.reputation = reputation
}

func (node *Node) String() string {
	return stringify.Struct("Node",
		stringify.StructField("id", node.id),
		stringify.StructField("reputation", node.reputation),
	)
}
