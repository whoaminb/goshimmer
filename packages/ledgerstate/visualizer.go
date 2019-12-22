package ledgerstate

import (
	"strings"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transfer"
	"github.com/iotaledger/goshimmer/packages/graphviz"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transferoutput"

	"github.com/emicklei/dot"
	"github.com/iotaledger/hive.go/objectstorage"
)

type transferOutputId [transfer.HashLength + address.Length]byte

type Visualizer struct {
	ledgerState         *LedgerState
	graph               *dot.Graph
	realitySubGraphs    map[reality.Id]*dot.Graph
	transferOutputNodes map[transferOutputId]dot.Node
}

func NewVisualizer(ledgerState *LedgerState) *Visualizer {
	return &Visualizer{
		ledgerState: ledgerState,
	}
}

func (visualizer *Visualizer) RenderTransferOutputs(pngFileName string) error {
	visualizer.reset()

	visualizer.graph.Attr("ranksep", "1.0 equally")
	visualizer.graph.Attr("compound", "true")

	visualizer.ledgerState.ForEachTransferOutput(func(object *objectstorage.CachedObject) bool {
		object.Consume(func(object objectstorage.StorableObject) {
			visualizer.drawTransferOutput(object.(*TransferOutput))
		})

		return true
	}, MAIN_REALITY_ID)

	return graphviz.RenderPNG(visualizer.graph, pngFileName)
}

func (visualizer *Visualizer) reset() *Visualizer {
	visualizer.graph = dot.NewGraph(dot.Directed)
	visualizer.realitySubGraphs = make(map[reality.Id]*dot.Graph)
	visualizer.transferOutputNodes = make(map[transferOutputId]dot.Node)

	return visualizer
}

func (visualizer *Visualizer) drawTransferOutput(transferOutput *TransferOutput) dot.Node {
	transferOutputIdentifier := visualizer.generateTransferOutputId(transferOutput)
	transferOutputNode, transferOutputDrawn := visualizer.transferOutputNodes[transferOutputIdentifier]

	if !transferOutputDrawn {
		transferOutputNode = visualizer.getRealitySubGraph(transferOutput.GetRealityId()).Node("OUTPUT: " + strings.Trim(transferOutput.GetTransferHash().String(), "\x00") + " / " + strings.Trim(transferOutput.GetAddressHash().String(), "\x00"))

		visualizer.styleTransferOutputNode(transferOutputNode)

		for transferHash, addresses := range transferOutput.GetConsumers() {
			for _, addressHash := range addresses {
				visualizer.ledgerState.GetTransferOutput(transferoutput.NewTransferOutputReference(transferHash, addressHash)).Consume(func(object objectstorage.StorableObject) {
					transferOutputNode.Edge(visualizer.drawTransferOutput(object.(*TransferOutput)))
				})
			}
		}

		visualizer.transferOutputNodes[transferOutputIdentifier] = transferOutputNode
	}

	return transferOutputNode
}

func (visualizer *Visualizer) generateTransferOutputId(transferOutput *TransferOutput) (result transferOutputId) {
	transferHash := transferOutput.GetTransferHash()
	addressHash := transferOutput.GetAddressHash()

	copy(result[:], transferHash[:])
	copy(result[transfer.HashLength:], addressHash[:])

	return
}

func (Visualizer *Visualizer) styleTransferOutputNode(transferOutputNode dot.Node) {
	transferOutputNode.Attr("fontname", "helvetica")
	transferOutputNode.Attr("fontsize", "11")
	transferOutputNode.Attr("style", "filled")
	transferOutputNode.Attr("shape", "box")
	transferOutputNode.Attr("color", "#6C8EBF")
	transferOutputNode.Attr("fillcolor", "white")
}

func (visualizer *Visualizer) getRealitySubGraph(realityId reality.Id) *dot.Graph {
	realityGraph, exists := visualizer.realitySubGraphs[realityId]
	if !exists {
		visualizer.ledgerState.GetReality(realityId).Consume(func(object objectstorage.StorableObject) {
			reality := object.(*Reality)

			parentRealities := reality.GetParentRealityIds()
			switch true {
			case len(parentRealities) > 1:
				realityGraph = visualizer.getRealitySubGraph(MAIN_REALITY_ID).Subgraph("AGGREGATED REALITY [ "+visualizer.generateRealityName(realityId)+" ]", dot.ClusterOption{})

				visualizer.styleRealitySubGraph(realityGraph, realityTypeAggregated)

				//dummyNode := realityGraph.Node(ledgerState.generateRealityName(parentRealities.ToList()...) + "_dummy")
				//dummyNode.Attr("shape", "point")
				//dummyNode.Attr("style", "invis")
				//dummyNode.Attr("peripheries", "0")
				//dummyNode.Attr("height", "0")
				//dummyNode.Attr("width", "0")
			case len(parentRealities) == 1:
				for parentRealityId := range parentRealities {
					realityGraph = visualizer.getRealitySubGraph(parentRealityId).Subgraph("REALITY [ "+visualizer.generateRealityName(realityId)+" ]", dot.ClusterOption{})

					visualizer.styleRealitySubGraph(realityGraph, realityTypeDefault)
				}
			default:
				realityGraph = visualizer.graph.Subgraph(visualizer.generateRealityName(realityId), dot.ClusterOption{})

				visualizer.styleRealitySubGraph(realityGraph, realityTypeMain)
			}
		})

		visualizer.realitySubGraphs[realityId] = realityGraph
	}

	return realityGraph
}

func (visualizer *Visualizer) styleRealitySubGraph(realitySubGraph *dot.Graph, realityType realityType) {
	realitySubGraph.Attr("fontname", "helvetica")
	realitySubGraph.Attr("fontsize", "11")
	realitySubGraph.Attr("style", "filled")
	realitySubGraph.Attr("nodesep", "0")

	switch realityType {
	case realityTypeAggregated:
		realitySubGraph.Attr("color", "#9673A6")
		realitySubGraph.Attr("fillcolor", "#E1D5E7")
	case realityTypeMain:
		realitySubGraph.Attr("color", "#D6B656")
		realitySubGraph.Attr("fillcolor", "#FFF2CC")
	case realityTypeDefault:
		realitySubGraph.Attr("color", "#6C8EBF")
		realitySubGraph.Attr("fillcolor", "#DAE8FC")
	}
}

func (visualizer *Visualizer) generateRealityName(realityId reality.Id) (result string) {
	visualizer.ledgerState.GetReality(realityId).Consume(func(object objectstorage.StorableObject) {
		reality := object.(*Reality)

		if reality.IsAggregated() {
			parentConflictRealities := reality.GetParentConflictRealities()
			realityIdCount := len(parentConflictRealities)
			counter := 1
			for realityId, parentConflictReality := range parentConflictRealities {
				result += visualizer.generateRealityName(realityId)

				if counter != realityIdCount {
					result += " + "
				}

				counter++

				parentConflictReality.Release()
			}
		} else {
			result = strings.Trim(realityId.String(), "\x00")
		}
	})

	return
}

type realityType int

const (
	realityTypeAggregated = iota
	realityTypeMain
	realityTypeDefault
)
