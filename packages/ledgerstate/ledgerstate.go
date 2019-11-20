package ledgerstate

import (
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"time"

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

func NewLedgerState(storageId string) *LedgerState {
	result := &LedgerState{
		storageId:              []byte(storageId),
		transferOutputs:        objectstorage.New(storageId+"TRANSFER_OUTPUTS", transferOutputFactory, objectstorage.CacheTime(1*time.Second)),
		transferOutputBookings: objectstorage.New(storageId+"TRANSFER_OUTPUT_BOOKING", transferOutputBookingFactory, objectstorage.CacheTime(1*time.Second)),
		realities:              objectstorage.New(storageId+"REALITIES", realityFactory, objectstorage.CacheTime(1*time.Second)),
		conflictSets:           objectstorage.New(storageId+"CONFLICT_SETS", conflictSetFactory, objectstorage.CacheTime(1*time.Second)),
	}

	mainReality := newReality(MAIN_REALITY_ID)
	mainReality.ledgerState = result
	result.realities.Store(mainReality).Release()

	return result
}

func (ledgerState *LedgerState) AddTransferOutput(transferHash TransferHash, addressHash AddressHash, balances ...*ColoredBalance) *LedgerState {
	ledgerState.storeTransferOutput(NewTransferOutput(ledgerState, MAIN_REALITY_ID, transferHash, addressHash, balances...)).Release()
	ledgerState.storeTransferOutputBooking(newTransferOutputBooking(MAIN_REALITY_ID, addressHash, false, transferHash)).Release()

	return ledgerState
}

func (ledgerState *LedgerState) GetTransferOutput(transferOutputReference *TransferOutputReference) *objectstorage.CachedObject {
	if cachedTransferOutput, err := ledgerState.transferOutputs.Load(transferOutputReference.GetStorageKey()); err != nil {
		panic(err)
	} else {
		if cachedTransferOutput.Exists() {
			if transferOutput := cachedTransferOutput.Get().(*TransferOutput); transferOutput != nil {
				transferOutput.ledgerState = ledgerState
			}
		}

		return cachedTransferOutput
	}
}

func (ledgerState *LedgerState) ForEachConflictSet(callback func(object *objectstorage.CachedObject) bool) {
	if err := ledgerState.conflictSets.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
		cachedObject.Get().(*ConflictSet).ledgerState = ledgerState

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
				booking := cachedObject.Get().(*TransferOutputBooking)
				cachedObject.Release()

				return callback(ledgerState.GetTransferOutput(NewTransferOutputReference(booking.GetTransferHash(), booking.GetAddressHash())))
			}, prefix); err != nil {
				panic(err)
			}
		}
	} else {
		if len(prefixes) >= 1 {
			for _, prefix := range prefixes {
				if err := ledgerState.transferOutputs.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
					cachedObject.Get().(*TransferOutput).ledgerState = ledgerState

					return callback(cachedObject)
				}, prefix); err != nil {
					panic(err)
				}
			}
		} else {
			if err := ledgerState.transferOutputs.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
				cachedObject.Get().(*TransferOutput).ledgerState = ledgerState

				return callback(cachedObject)
			}); err != nil {
				panic(err)
			}
		}
	}
}

func (ledgerState *LedgerState) CreateReality(id RealityId) {
	newReality := newReality(id, MAIN_REALITY_ID)
	newReality.ledgerState = ledgerState

	ledgerState.realities.Store(newReality)
}

func (ledgerState *LedgerState) GetReality(id RealityId) *objectstorage.CachedObject {
	if cachedObject, err := ledgerState.realities.Load(id[:]); err != nil {
		panic(err)
	} else {
		if cachedObject.Exists() {
			if reality := cachedObject.Get().(*Reality); reality != nil {
				reality.ledgerState = ledgerState
			}
		}

		return cachedObject
	}
}

func (ledgerState *LedgerState) BookTransfer(transfer *Transfer) error {
	inputs := ledgerState.getTransferInputs(transfer)

	targetReality := ledgerState.getTargetReality(inputs)
	defer targetReality.Release()

	if bookingErr := targetReality.Get().(*Reality).bookTransfer(transfer.GetHash(), inputs, transfer.GetOutputs()); bookingErr != nil {
		return bookingErr
	}

	if !targetReality.IsStored() {
		targetReality.Store()
	}

	return nil
}

func (ledgerState *LedgerState) GenerateRealityVisualization(pngFilename string) error {
	di := dot.NewGraph(dot.Directed)

	realityNodes := make(map[RealityId]dot.Node)

	var drawReality func(reality *Reality) dot.Node
	drawReality = func(reality *Reality) dot.Node {
		realityNode, exists := realityNodes[reality.GetId()]
		if !exists {
			if len(reality.parentRealities) > 1 {
				realityNode = di.Node("AGGREGATED REALITY: " + strings.Trim(reality.GetId().String(), "\x00"))
			} else {
				realityNode = di.Node("REALITY: " + strings.Trim(reality.GetId().String(), "\x00"))
			}
			realityNode.Attr("style", "rounded")
			realityNode.Attr("shape", "rect")

			realityNodes[reality.GetId()] = realityNode
		}

		if !exists {
			parentRealities := reality.GetParentRealities()

			for _, cachedReality := range parentRealities {
				cachedReality.Consume(func(object objectstorage.StorableObject) {
					realityNode.Edge(drawReality(object.(*Reality)))
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

	d1 := []byte(di.String())
	if err := ioutil.WriteFile("_tmp.dot", d1, 0755); err != nil {
		return err
	}

	_, err := exec.Command("dot", "-Tpng", "_tmp.dot", "-o", pngFilename).Output()
	if err != nil {
		return err
	}

	if err := os.Remove("_tmp.dot"); err != nil {
		return err
	}

	return nil
}

func (ledgerState *LedgerState) GenerateVisualization() error {
	di := dot.NewGraph(dot.Directed)

	realityGraphs := make(map[RealityId]*dot.Graph)

	getRealityGraph := func(realityId RealityId) *dot.Graph {
		realityGraph, exists := realityGraphs[realityId]
		if !exists {
			realityGraph = di.Subgraph("REALITY: "+strings.Trim(realityId.String(), "\x00"), dot.ClusterOption{})
			realityGraph.Attr("style", "filled")
			realityGraph.Attr("color", "lightgrey")
			realityGraph.Attr("fillcolor", "lightgrey")

			realityGraphs[realityId] = realityGraph
		}

		return realityGraph
	}

	var drawTransferOutput func(transferOutput *TransferOutput) dot.Node
	drawTransferOutput = func(transferOutput *TransferOutput) dot.Node {
		transferOutputNode := getRealityGraph(transferOutput.GetRealityId()).Node(transferOutput.GetTransferHash().String() + "\n" + transferOutput.GetAddressHash().String())
		transferOutputNode.Attr("style", "filled")
		transferOutputNode.Attr("color", "white")
		transferOutputNode.Attr("fillcolor", "white")

		for transferHash, addresses := range transferOutput.GetConsumers() {
			for _, addressHash := range addresses {
				ledgerState.GetTransferOutput(NewTransferOutputReference(transferHash, addressHash)).Consume(func(object objectstorage.StorableObject) {
					transferOutputNode.Edge(drawTransferOutput(object.(*TransferOutput)))
				})
			}
		}

		return transferOutputNode
	}

	ledgerState.ForEachTransferOutput(func(object *objectstorage.CachedObject) bool {
		object.Consume(func(object objectstorage.StorableObject) {
			drawTransferOutput(object.(*TransferOutput))
		})

		return true
	}, MAIN_REALITY_ID)

	d1 := []byte(di.String())
	if err := ioutil.WriteFile("_tmp.dot", d1, 0755); err != nil {
		return err
	}

	_, err := exec.Command("dot", "-Tpng", "_tmp.dot", "-o", "viz.png").Output()
	if err != nil {
		return err
	}

	if err := os.Remove("_tmp.dot"); err != nil {
		return err
	}

	return nil
}

func (ledgerState *LedgerState) MergeRealities(realityIds ...RealityId) *objectstorage.CachedObject {
	switch len(realityIds) {
	case 0:
		if loadedReality, loadedRealityErr := ledgerState.realities.Load(MAIN_REALITY_ID[:]); loadedRealityErr != nil {
			panic(loadedRealityErr)
		} else {
			loadedReality.Get().(*Reality).ledgerState = ledgerState

			return loadedReality
		}
	case 1:
		if loadedReality, loadedRealityErr := ledgerState.realities.Load(realityIds[0][:]); loadedRealityErr != nil {
			panic(loadedRealityErr)
		} else {
			loadedReality.Get().(*Reality).ledgerState = ledgerState

			return loadedReality
		}
	default:
		aggregatedRealities := make(map[RealityId]*objectstorage.CachedObject)

	AGGREGATE_REALITIES:
		for _, realityId := range realityIds {
			// check if we have processed this reality already
			if _, exists := aggregatedRealities[realityId]; exists {
				continue
			}

			// load reality or abort if it fails
			cachedReality, loadingErr := ledgerState.realities.Load(realityId[:])
			if loadingErr != nil {
				panic(loadingErr)
			} else if !cachedReality.Exists() {
				panic(errors.New("referenced reality does not exist: " + realityId.String()))
			}

			// type cast the reality
			reality := cachedReality.Get().(*Reality)

			// check if the reality is already included in the aggregated realities
			for aggregatedRealityId, aggregatedReality := range aggregatedRealities {
				// if an already aggregated reality "descends" from the current reality, then we have found the more
				// "specialized" reality already and keep it
				if aggregatedReality.Get().(*Reality).DescendsFromReality(realityId) {
					continue AGGREGATE_REALITIES
				}

				// if the current reality
				if reality.DescendsFromReality(aggregatedRealityId) {
					delete(aggregatedRealities, aggregatedRealityId)
					aggregatedReality.Release()

					aggregatedRealities[reality.GetId()] = cachedReality

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

		sortedRealityIds := ledgerState.sortRealityIds(aggregatedRealities)

		aggregatedReality := newReality(ledgerState.generateAggregatedRealityId(sortedRealityIds), sortedRealityIds...)
		aggregatedReality.ledgerState = ledgerState

		return ledgerState.realities.Prepare(aggregatedReality)
	}
}

func (ledgerState *LedgerState) Prune() *LedgerState {
	if err := ledgerState.transferOutputs.Prune(); err != nil {
		panic(err)
	}

	if err := ledgerState.transferOutputBookings.Prune(); err != nil {
		panic(err)
	}

	if err := ledgerState.realities.Prune(); err != nil {
		panic(err)
	}

	mainReality := newReality(MAIN_REALITY_ID)
	mainReality.ledgerState = ledgerState
	ledgerState.realities.Store(mainReality).Release()

	return ledgerState
}

func (ledgerState *LedgerState) generateFilterPrefixes(filters []interface{}) ([][]byte, bool) {
	filteredRealities := make([]RealityId, 0)
	filteredAddresses := make([]AddressHash, 0)
	filteredTransfers := make([]TransferHash, 0)
	filterSpent := false
	filterUnspent := false

	for _, filter := range filters {
		switch typeCastedValue := filter.(type) {
		case RealityId:
			filteredRealities = append(filteredRealities, typeCastedValue)
		case AddressHash:
			filteredAddresses = append(filteredAddresses, typeCastedValue)
		case TransferHash:
			filteredTransfers = append(filteredTransfers, typeCastedValue)
		case SpentIndicator:
			switch typeCastedValue {
			case SPENT:
				filterSpent = true
			case UNSPENT:
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
							spentPrefix = append(spentPrefix, byte(SPENT))
						} else {
							spentPrefix = append(spentPrefix, byte(UNSPENT))
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

func (ledgerState *LedgerState) storeTransferOutput(transferOutput *TransferOutput) *objectstorage.CachedObject {
	return ledgerState.transferOutputs.Store(transferOutput)
}

func (ledgerState *LedgerState) storeTransferOutputBooking(transferOutputBooking *TransferOutputBooking) *objectstorage.CachedObject {
	return ledgerState.transferOutputBookings.Store(transferOutputBooking)
}

func (ledgerState *LedgerState) sortRealityIds(aggregatedRealities map[RealityId]*objectstorage.CachedObject) []RealityId {
	counter := 0
	sortedRealityIds := make([]RealityId, len(aggregatedRealities))
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

func (ledgerState *LedgerState) generateAggregatedRealityId(sortedRealityIds []RealityId) [32]byte {
	aggregatedRealityId := make([]byte, 0)
	for _, realityId := range sortedRealityIds {
		aggregatedRealityId = append(aggregatedRealityId, realityId[:]...)
	}

	return blake2b.Sum256(aggregatedRealityId)
}

func (ledgerState *LedgerState) getTargetReality(inputs []*objectstorage.CachedObject) *objectstorage.CachedObject {
	realityIds := make([]RealityId, len(inputs))
	for i, input := range inputs {
		realityIds[i] = input.Get().(*TransferOutput).GetRealityId()
	}

	return ledgerState.MergeRealities(realityIds...)
}

func (ledgerState *LedgerState) getTransferInputs(transfer *Transfer) []*objectstorage.CachedObject {
	inputs := transfer.GetInputs()
	result := make([]*objectstorage.CachedObject, len(inputs))

	for i, transferOutputReference := range inputs {
		result[i] = ledgerState.GetTransferOutput(transferOutputReference)
	}

	return result
}

func transferOutputFactory(key []byte) objectstorage.StorableObject {
	result := &TransferOutput{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}

func transferOutputBookingFactory(key []byte) objectstorage.StorableObject {
	result := &TransferOutputBooking{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}

func realityFactory(key []byte) objectstorage.StorableObject {
	result := &Reality{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}

func conflictSetFactory(key []byte) objectstorage.StorableObject {
	result := &ConflictSet{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}
