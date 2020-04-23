package sctransaction

// body of the state update
// it will contain:
// - program hash
// - optional reference to the state uodatein the qnode ledger (db)
// - optional variable/value pairs, kept in the transaction
type StateBody struct {
}

// variable/value pairs
type RequestBody struct {
}

// creates empty request body
func NewRequestBody() *RequestBody {
	return &RequestBody{}
}
