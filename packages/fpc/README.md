 # FPC

This package implements the Fast Probabilistic Consensus protocol.

## Dependencies


* `Queryables func() -> []queryable` It should provide an interface for querying the nodes
* `(queryable).Query func(context, []Hash) -> []Opinion` Add a context for canceling after timeout expires

## Interfaces

* `(*FPC) VoteOnTxs([]TxOpinion)` : adds given txs to the FPC waiting txs list and set the initial opinion to the opinion history
* `(*FPC) Tick(uint64, float64)` : updates FPC state with the new random and starts a new round
* `(*FPC) GetInterimOpinion([]Hash) -> []Opinion` : returns the current opinion of the given txs

* Finalized txs are notified via the channel `FinalizedTxs` of the particular FPC instance. TODO: maybe we can find a better way to do that.

## Test

From the `goshimmer` folder run:

```
go test -count=1 -v -cover ./packages/fpc
```

TODO: unit test

