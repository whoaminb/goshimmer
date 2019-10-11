package ledgerstate

type LedgerState struct {
	mainReality  *Reality
	subRealities map[string]*Reality
}
