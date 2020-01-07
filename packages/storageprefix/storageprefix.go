package storageprefix

var (
	TangleTransaction         = []byte{0}
	TangleTransactionMetadata = []byte{6}
	TangleApprovers           = []byte{1}
	TangleMissingTransaction  = []byte{7}

	LedgerStateTransferOutput        = []byte{2}
	LedgerStateTransferOutputBooking = []byte{3}
	LedgerStateReality               = []byte{4}
	LedgerStateConflictSet           = []byte{5}
)
