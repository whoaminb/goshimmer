package ledgerstate

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/packages/storageprefix"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/graphviz"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/conflict"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"

	"golang.org/x/crypto/blake2b"

	"github.com/emicklei/dot"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/hive.go/objectstorage"
)

type LedgerState struct {
	storageId              []byte
	transferOutputs        *objectstorage.ObjectStorage
	transferOutputBookings *objectstorage.ObjectStorage
	realities              *objectstorage.ObjectStorage
	conflictSets           *objectstorage.ObjectStorage
}

func NewLedgerState(storageId []byte) *LedgerState {
	result := &LedgerState{
		storageId:              storageId,
		transferOutputs:        objectstorage.New(append(storageId, storageprefix.LedgerStateTransferOutput...), transfer.OutputFactory, objectstorage.CacheTime(1*time.Second)),
		transferOutputBookings: objectstorage.New(append(storageId, storageprefix.LedgerStateTransferOutputBooking...), transfer.OutputBookingFactory, objectstorage.CacheTime(1*time.Second)),
		realities:              objectstorage.New(append(storageId, storageprefix.LedgerStateReality...), realityFactory, objectstorage.CacheTime(1*time.Second)),
		conflictSets:           objectstorage.New(append(storageId, storageprefix.LedgerStateConflictSet...), conflict.Factory, objectstorage.CacheTime(1*time.Second)),
	}

	mainReality := newReality(reality.MAIN_ID)
	mainReality.ledgerState = result
	mainReality.SetPreferred()
	result.realities.Store(mainReality).Release()

	return result
}

func (ledgerState *LedgerState) AddTransferOutput(transferHash transfer.Hash, addressHash address.Address, balances ...*coloredcoins.ColoredBalance) *LedgerState {
	ledgerState.GetReality(reality.MAIN_ID).Consume(func(object objectstorage.StorableObject) {
		mainReality := object.(*Reality)

		mainReality.bookTransferOutput(transfer.NewTransferOutput(ledgerState.transferOutputBookings, reality.EmptyId, transferHash, addressHash, balances...))
	})

	return ledgerState
}

func (ledgerState *LedgerState) GetTransferOutput(transferOutputReference *transfer.OutputReference) *objectstorage.CachedObject {
	cachedTransferOutput := ledgerState.transferOutputs.Load(transferOutputReference.GetStorageKey())
	if cachedTransferOutput.Exists() {
		if transferOutput := cachedTransferOutput.Get().(*transfer.Output); transferOutput != nil {
			transferOutput.OutputBookings = ledgerState.transferOutputBookings
		}
	}

	return cachedTransferOutput
}

func (ledgerState *LedgerState) ForEachConflictSet(callback func(object *objectstorage.CachedObject) bool) {
	if err := ledgerState.conflictSets.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
		return callback(cachedObject)
	}); err != nil {
		panic(err)
	}
}

func (ledgerState *LedgerState) ForEachReality(callback func(object *objectstorage.CachedObject) bool) {
	if err := ledgerState.realities.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
		cachedObject.Get().(*Reality).ledgerState = ledgerState

		return callback(cachedObject)
	}); err != nil {
		panic(err)
	}
}

func (ledgerState *LedgerState) ForEachTransferOutput(callback func(object *objectstorage.CachedObject) bool, filters ...interface{}) {
	prefixes, searchBookings := ledgerState.generateFilterPrefixes(filters)
	if searchBookings {
		for _, prefix := range prefixes {
			if err := ledgerState.transferOutputBookings.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
				booking := cachedObject.Get().(*transfer.OutputBooking)
				cachedObject.Release()

				return callback(ledgerState.GetTransferOutput(transfer.NewOutputReference(booking.GetTransferHash(), booking.GetAddressHash())))
			}, prefix); err != nil {
				panic(err)
			}
		}
	} else {
		if len(prefixes) >= 1 {
			for _, prefix := range prefixes {
				if err := ledgerState.transferOutputs.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
					cachedObject.Get().(*transfer.Output).OutputBookings = ledgerState.transferOutputBookings

					return callback(cachedObject)
				}, prefix); err != nil {
					panic(err)
				}
			}
		} else {
			if err := ledgerState.transferOutputs.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
				cachedObject.Get().(*transfer.Output).OutputBookings = ledgerState.transferOutputBookings

				return callback(cachedObject)
			}); err != nil {
				panic(err)
			}
		}
	}
}

func (ledgerState *LedgerState) CreateReality(id reality.Id) {
	newReality := newReality(id, reality.MAIN_ID)
	newReality.ledgerState = ledgerState

	ledgerState.realities.Load(reality.MAIN_ID[:]).Consume(func(object objectstorage.StorableObject) {
		object.(*Reality).RegisterSubReality(id)
	})

	ledgerState.realities.Store(newReality).Release()
}

func (ledgerState *LedgerState) GetReality(id reality.Id) *objectstorage.CachedObject {
	cachedObject := ledgerState.realities.Load(id[:])
	if cachedObject.Exists() {
		if reality := cachedObject.Get().(*Reality); reality != nil {
			reality.ledgerState = ledgerState
		}
	}

	return cachedObject
}

func (ledgerState *LedgerState) BookTransfer(transfer *transfer.Transfer) (err error) {
	inputs := ledgerState.getTransferInputs(transfer)

	ledgerState.getTargetReality(inputs).Consume(func(object objectstorage.StorableObject) {
		targetReality := object.(*Reality)

		if err = targetReality.bookTransfer(transfer.GetHash(), inputs, transfer.GetOutputs()); err != nil {
			return
		}

		targetReality.Persist()
	})

	return
}

func (ledgerState *LedgerState) GenerateRealityVisualization(pngFilename string) error {
	graph := dot.NewGraph(dot.Directed)
	graph.Attr("ranksep", "1.0 equally")

	realityNodes := make(map[reality.Id]dot.Node)

	drawConflictSet := func(conflict *conflict.Conflict) {
		conflictSetNode := graph.Node(strings.Trim(conflict.GetId().String(), "\x00"))
		conflictSetNode.Attr("label", "")
		conflictSetNode.Attr("shape", "Mdiamond")
		conflictSetNode.Attr("style", "filled")
		conflictSetNode.Attr("color", "#B85450")
		conflictSetNode.Attr("fillcolor", "#F8CECC")

		for realityId := range conflict.Members {
			conflictSetNode.Edge(realityNodes[realityId]).Attr("arrowhead", "none").Attr("arrowtail", "none").Attr("color", "#B85450")
		}
	}

	var drawReality func(reality *Reality) dot.Node
	drawReality = func(reality *Reality) dot.Node {
		realityNode, exists := realityNodes[reality.id]
		if !exists {
			if reality.IsAggregated() {
				realityNode = graph.Node("AGGREGATED REALITY\n\n" + strings.Trim(reality.id.String(), "\x00") + " (" + strconv.Itoa(int(reality.GetTransferOutputCount())) + " / " + strconv.Itoa(len(reality.subRealityIds)) + ")")
				realityNode.Attr("style", "filled")
				realityNode.Attr("shape", "rect")
				realityNode.Attr("color", "#9673A6")
				realityNode.Attr("fillcolor", "#E1D5E7")
			} else {
				realityNode = graph.Node("REALITY\n\n" + strings.Trim(reality.id.String(), "\x00") + " (" + strconv.Itoa(int(reality.GetTransferOutputCount())) + " / " + strconv.Itoa(len(reality.subRealityIds)) + ")")
				realityNode.Attr("style", "filled")
				realityNode.Attr("shape", "rect")
				realityNode.Attr("color", "#6C8EBF")
				realityNode.Attr("fillcolor", "#DAE8FC")
			}

			if reality.IsLiked() {
				realityNode.Attr("penwidth", "3.0")
			}

			realityNodes[reality.id] = realityNode
		}

		if !exists {
			parentRealities := reality.GetParentRealities()

			for _, cachedReality := range parentRealities {
				cachedReality.Consume(func(object objectstorage.StorableObject) {
					realityNode.Edge(drawReality(object.(*Reality))).Attr("arrowhead", "none").Attr("arrowtail", "none")
				})
			}
		}

		return realityNode
	}

	ledgerState.ForEachReality(func(object *objectstorage.CachedObject) bool {
		object.Consume(func(object objectstorage.StorableObject) {
			drawReality(object.(*Reality))
		})

		return true
	})
	ledgerState.ForEachConflictSet(func(object *objectstorage.CachedObject) bool {
		object.Consume(func(object objectstorage.StorableObject) {
			drawConflictSet(object.(*conflict.Conflict))
		})

		return true
	})

	return graphviz.RenderPNG(graph, pngFilename)
}

func (ledgerState *LedgerState) AggregateRealities(realityIds ...reality.Id) *objectstorage.CachedObject {
	switch len(realityIds) {
	case 0:
		loadedReality := ledgerState.realities.Load(reality.MAIN_ID[:])
		loadedReality.Get().(*Reality).ledgerState = ledgerState

		return loadedReality
	case 1:
		loadedReality := ledgerState.realities.Load(realityIds[0][:])
		loadedReality.Get().(*Reality).ledgerState = ledgerState

		return loadedReality
	default:
		aggregatedRealities := make(map[reality.Id]*objectstorage.CachedObject)

	AGGREGATE_REALITIES:
		for _, realityId := range realityIds {
			// check if we have processed this reality already
			if _, exists := aggregatedRealities[realityId]; exists {
				continue
			}

			// load reality or abort if it fails
			cachedReality := ledgerState.realities.Load(realityId[:])
			if !cachedReality.Exists() {
				panic(errors.New("referenced reality does not exist: " + realityId.String()))
			}

			// type cast the reality
			reality := cachedReality.Get().(*Reality)

			// check if the reality is already included in the aggregated realities
			for aggregatedRealityId, aggregatedReality := range aggregatedRealities {
				// if an already aggregated reality "descends" from the current reality, then we have found the more
				// "specialized" reality already and keep it
				if aggregatedReality.Get().(*Reality).DescendsFrom(realityId) {
					continue AGGREGATE_REALITIES
				}

				// if the current reality
				if reality.DescendsFrom(aggregatedRealityId) {
					delete(aggregatedRealities, aggregatedRealityId)
					aggregatedReality.Release()

					aggregatedRealities[reality.id] = cachedReality

					continue AGGREGATE_REALITIES
				}
			}

			// store the reality as a new aggregate candidate
			aggregatedRealities[realityId] = cachedReality
		}

		if len(aggregatedRealities) == 1 {
			for _, independentReality := range aggregatedRealities {
				return independentReality
			}
		}

		parentConflictRealities := make(map[reality.Id]*objectstorage.CachedObject)
		aggregatedRealityParentIds := make([]reality.Id, len(aggregatedRealities))

		aggregatedRealityIsPreferred := true

		counter := 0
		for aggregatedRealityId, cachedAggregatedReality := range aggregatedRealities {
			aggregatedRealityParentIds[counter] = aggregatedRealityId
			counter++

			aggregatedReality := cachedAggregatedReality.Get().(*Reality)

			aggregatedRealityIsPreferred = aggregatedRealityIsPreferred && aggregatedReality.IsPreferred()

			if !aggregatedReality.IsAggregated() {
				parentConflictRealities[aggregatedRealityId] = cachedAggregatedReality
			} else {
				aggregatedReality.collectParentConflictRealities(parentConflictRealities)

				cachedAggregatedReality.Release()
			}
		}

		aggregatedRealityId := ledgerState.generateAggregatedRealityId(ledgerState.sortRealityIds(parentConflictRealities))

		newAggregatedRealityCreated := false
		newCachedAggregatedReality := ledgerState.realities.ComputeIfAbsent(aggregatedRealityId[:], func(key []byte) (object objectstorage.StorableObject) {
			aggregatedReality := newReality(aggregatedRealityId, aggregatedRealityParentIds...)
			aggregatedReality.ledgerState = ledgerState
			aggregatedReality.SetPreferred(aggregatedRealityIsPreferred)

			for _, parentRealityId := range aggregatedRealityParentIds {
				ledgerState.GetReality(parentRealityId).Consume(func(object objectstorage.StorableObject) {
					object.(*Reality).RegisterSubReality(aggregatedRealityId)
				})
			}

			aggregatedReality.SetModified()

			newAggregatedRealityCreated = true

			return aggregatedReality
		})

		if !newAggregatedRealityCreated {
			aggregatedReality := newCachedAggregatedReality.Get().(*Reality)

			for _, realityId := range aggregatedRealityParentIds {
				if aggregatedReality.AddParentReality(realityId) {
					ledgerState.GetReality(realityId).Consume(func(object objectstorage.StorableObject) {
						object.(*Reality).RegisterSubReality(aggregatedRealityId)
					})
				}
			}
		}

		return newCachedAggregatedReality
	}
}

func (ledgerState *LedgerState) Prune() *LedgerState {
	time.Sleep(2 * time.Second)

	if err := ledgerState.transferOutputs.Prune(); err != nil {
		panic(err)
	}

	if err := ledgerState.transferOutputBookings.Prune(); err != nil {
		panic(err)
	}

	if err := ledgerState.realities.Prune(); err != nil {
		panic(err)
	}

	if err := ledgerState.conflictSets.Prune(); err != nil {
		panic(err)
	}

	mainReality := newReality(reality.MAIN_ID)
	mainReality.ledgerState = ledgerState
	mainReality.SetPreferred()
	ledgerState.realities.Store(mainReality).Release()

	return ledgerState
}

func (ledgerState *LedgerState) generateFilterPrefixes(filters []interface{}) ([][]byte, bool) {
	filteredRealities := make([]reality.Id, 0)
	filteredAddresses := make([]address.Address, 0)
	filteredTransfers := make([]transfer.Hash, 0)
	filterSpent := false
	filterUnspent := false

	for _, filter := range filters {
		switch typeCastedValue := filter.(type) {
		case reality.Id:
			filteredRealities = append(filteredRealities, typeCastedValue)
		case address.Address:
			filteredAddresses = append(filteredAddresses, typeCastedValue)
		case transfer.Hash:
			filteredTransfers = append(filteredTransfers, typeCastedValue)
		case transfer.SpentIndicator:
			switch typeCastedValue {
			case transfer.SPENT:
				filterSpent = true
			case transfer.UNSPENT:
				filterUnspent = true
			default:
				panic("unknown SpentIndicator")
			}
		default:
			panic("unknown filter type: " + reflect.ValueOf(filter).Kind().String())
		}
	}

	prefixes := make([][]byte, 0)
	if len(filteredRealities) >= 1 {
		for _, realityId := range filteredRealities {
			realityPrefix := append([]byte{}, realityId[:]...)

			if len(filteredAddresses) >= 1 {
				for _, addressHash := range filteredAddresses {
					addressPrefix := append([]byte{}, realityPrefix...)
					addressPrefix = append(addressPrefix, addressHash[:]...)

					if filterSpent != filterUnspent {
						spentPrefix := append([]byte{}, addressPrefix...)
						if filterSpent {
							spentPrefix = append(spentPrefix, byte(transfer.SPENT))
						} else {
							spentPrefix = append(spentPrefix, byte(transfer.UNSPENT))
						}

						// TODO: FILTER TRANSFER HASH
						prefixes = append(prefixes, spentPrefix)
					} else {
						prefixes = append(prefixes, addressPrefix)
					}
				}
			} else {
				prefixes = append(prefixes, realityPrefix)
			}
		}

		return prefixes, true
	} else if len(filteredTransfers) >= 1 {
		for _, transferHash := range filteredTransfers {
			transferPrefix := append([]byte{}, transferHash[:]...)

			if len(filteredAddresses) >= 1 {
				for _, addressHash := range filteredAddresses {
					addressPrefix := append([]byte{}, transferPrefix...)
					addressPrefix = append(addressPrefix, addressHash[:]...)

					prefixes = append(prefixes, addressPrefix)
				}
			} else {
				prefixes = append(prefixes, transferPrefix)
			}
		}
	}

	return prefixes, false
}

func (ledgerState *LedgerState) storeTransferOutput(transferOutput *transfer.Output) *objectstorage.CachedObject {
	return ledgerState.transferOutputs.Store(transferOutput)
}

func (ledgerState *LedgerState) sortRealityIds(aggregatedRealities map[reality.Id]*objectstorage.CachedObject) []reality.Id {
	counter := 0
	sortedRealityIds := make([]reality.Id, len(aggregatedRealities))
	for realityId, aggregatedReality := range aggregatedRealities {
		sortedRealityIds[counter] = realityId

		counter++

		aggregatedReality.Release()
	}

	sort.Slice(sortedRealityIds, func(i, j int) bool {
		for k := 0; k < len(sortedRealityIds[k]); k++ {
			if sortedRealityIds[i][k] < sortedRealityIds[j][k] {
				return true
			} else if sortedRealityIds[i][k] > sortedRealityIds[j][k] {
				return false
			}
		}

		return false
	})

	return sortedRealityIds
}

func (ledgerState *LedgerState) generateAggregatedRealityId(sortedRealityIds []reality.Id) [32]byte {
	aggregatedRealityId := make([]byte, 0)
	for _, realityId := range sortedRealityIds {
		aggregatedRealityId = append(aggregatedRealityId, realityId[:]...)
	}

	return blake2b.Sum256(aggregatedRealityId)
}

func (ledgerState *LedgerState) getTargetReality(inputs []*objectstorage.CachedObject) *objectstorage.CachedObject {
	realityIds := make([]reality.Id, len(inputs))
	for i, input := range inputs {
		realityIds[i] = input.Get().(*transfer.Output).GetRealityId()
	}

	return ledgerState.AggregateRealities(realityIds...)
}

func (ledgerState *LedgerState) getTransferInputs(transfer *transfer.Transfer) []*objectstorage.CachedObject {
	inputs := transfer.GetInputs()
	result := make([]*objectstorage.CachedObject, len(inputs))

	for i, transferOutputReference := range inputs {
		result[i] = ledgerState.GetTransferOutput(transferOutputReference)
	}

	return result
}

func realityFactory(key []byte) objectstorage.StorableObject {
	result := &Reality{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}
