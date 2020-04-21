package clientapi

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"sort"
)

type NewOriginParams struct {
	AssemblyId      *HashValue
	ConfigId        *HashValue
	AssemblyAccount *HashValue
	OwnerAccount    *HashValue
}

// transfer is not signed

func NewScOriginTransaction(par NewOriginParams) (sc.Transaction, error) {
	ret := sc.NewTransaction()
	_, err := MoveFundsFromToAddress(ret, par.OwnerAccount, par.AssemblyAccount, []uint64{1})
	if err != nil {
		return nil, err
	}
	state := sc.NewStateBlock(par.AssemblyId, par.ConfigId, nil)
	configVars := state.Config().Vars()
	configVars.SetString(sc.MAP_KEY_ASSEMBLY_ACCOUNT, par.AssemblyAccount.String())
	configVars.SetString(sc.MAP_KEY_OWNER_ACCOUNT, par.OwnerAccount.String())
	ret.SetState(state)
	return ret, nil
}

type NewRequestParams struct {
	AssemblyId       *HashValue
	AssemblyAccount  *HashValue
	RequesterAccount *HashValue
	Reward           uint64
	Deposit          uint64
	Vars             generic.ValueMap
}

func NewRequestTransaction(par NewRequestParams) (sc.Transaction, error) {
	tx := sc.NewTransaction()
	_, err := AddNewRequestBlock(tx, par)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func AddNewRequestBlock(tx sc.Transaction, par NewRequestParams) (*sc.RequestRef, error) {
	amounts := []uint64{1}
	if par.Reward > 0 {
		amounts = append(amounts, par.Reward)
	}
	if par.Deposit > 0 {
		amounts = append(amounts, par.Deposit)
	}
	outIndices, err := MoveFundsFromToAddress(tx, par.RequesterAccount, par.AssemblyAccount, amounts)
	if err != nil {
		return nil, err
	}
	reqBlk := sc.NewRequestBlock(par.AssemblyId, false)
	if par.Vars != nil {
		// must be before setting indices
		reqBlk.WithVars(par.Vars)
	}
	reqBlk.WithRequestChainOutputIndex(outIndices[0])
	if par.Reward > 0 {
		reqBlk.WithRewardOutputIndex(outIndices[1])
	}
	if par.Deposit > 0 {
		reqBlk.WithDepositOutputIndex(outIndices[2])
	}
	reqIdx := tx.AddRequest(reqBlk)
	return sc.NewRequestRefFromTx(tx, reqIdx)
}

func NewResultTransaction(reqRef *sc.RequestRef, config sc.Config) (sc.Transaction, error) {
	reqBlock := reqRef.RequestBlock()
	// check if request block points to valid chain output
	// which can be used as request->result chain
	requestChainOutput := reqBlock.MainOutputs(reqRef.Tx()).RequestChainOutput
	if requestChainOutput == nil {
		return nil, fmt.Errorf("can't find request chain output in the request transaction")
	}
	if requestChainOutput.Value != 1 {
		return nil, fmt.Errorf("request chain output must be 1i")
	}
	if !value.OutputCanBeChained(&requestChainOutput.OutputRef, config.SContractAccount()) {
		return nil, fmt.Errorf("invalid request chain output")
	}
	tx := sc.NewTransaction()
	// add request chain link
	// transfer 1i from RequestChainAddress to itself
	tx.Transfer().AddInput(value.NewInputFromOutputRef(&requestChainOutput.OutputRef))
	tx.Transfer().AddOutput(value.NewOutput(config.SContractAccount(), 1))
	return tx, nil
}

func NextStateUpdateTransaction(prevStateTx sc.Transaction, reqRef *sc.RequestRef) (sc.Transaction, error) {
	prevState, ok := prevStateTx.State()
	if !ok {
		return nil, fmt.Errorf("NextStateUpdateTransaction: state block not found")
	}
	tx, err := NewResultTransaction(reqRef, prevState.Config())
	if err != nil {
		return nil, err
	}
	tx.Transfer().AddInput(value.NewInput(prevStateTx.Transfer().Id(), prevStateTx.MustState().StateChainOutputIndex()))
	chainOutIdx := tx.Transfer().AddOutput(value.NewOutput(prevState.Config().SContractAccount(), 1))

	nextState := sc.NewStateBlock(prevState.SContractId(), prevState.Config().Id(), reqRef)
	nextState.
		WithStateIndex(prevState.StateIndex() + 1).
		WithVars(prevState.Vars()).
		WithStateChainOutputIndex(chainOutIdx)
	nextState.Config().With(prevState.Config())
	tx.SetState(nextState)
	return tx, nil
}

func ErrorTransaction(reqRef *sc.RequestRef, config sc.Config, resultErr error) (sc.Transaction, error) {
	tx := sc.NewTransaction()
	// with empty transfer
	errState := sc.NewStateBlock(reqRef.RequestBlock().SContractId(), config.Id(), reqRef).
		WithError(resultErr)
	tx.SetState(errState)
	return tx, nil
}

func SendAllOutputsToAddress(tx sc.Transaction, outputs []*generic.OutputRef, addr *HashValue) error {
	sum := uint64(0)
	for _, outp := range outputs {
		tx.Transfer().AddInput(value.NewInputFromOutputRef(outp))
		oav, err := value.GetOutputAddrValue(outp)
		if err != nil {
			return err
		}
		sum += oav.Value
	}
	tx.Transfer().AddOutput(value.NewOutput(addr, sum))
	return nil
}

func SendOutputsToOutputs(tx sc.Transaction, inOutputs []*generic.OutputRef, outOutputs []value.Output, reminderAddr *HashValue) error {
	sumInp := uint64(0)
	for _, outp := range inOutputs {
		oav, err := value.GetOutputAddrValue(outp)
		if err != nil {
			return err
		}
		sumInp += oav.Value
	}
	sumOutp := uint64(0)
	for _, outp := range outOutputs {
		sumOutp += outp.Value()
	}
	if sumOutp > sumInp {
		return fmt.Errorf("not enough funds")
	}
	for _, outp := range inOutputs {
		tx.Transfer().AddInput(value.NewInputFromOutputRef(outp))
	}
	for _, outp := range outOutputs {
		tx.Transfer().AddOutput(outp)
	}
	if sumOutp < sumInp {
		tx.Transfer().AddOutput(value.NewOutput(reminderAddr, sumInp-sumOutp))
	}
	return nil
}

type outputsByValue []*generic.OutputRefWithAddrValue

func (s outputsByValue) Len() int {
	return len(s)
}

func (s outputsByValue) Less(i, j int) bool {
	return s[i].Value < s[j].Value
}

func (s outputsByValue) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func MoveFundsFromToAddress(tx sc.Transaction, addrFrom, addrTo *HashValue, amounts []uint64) ([]uint16, error) {
	uos := value.GetUnspentOutputs(addrFrom)
	sort.Sort(outputsByValue(uos))
	sumToSend := uint64(0)
	for _, s := range amounts {
		sumToSend += s
	}
	if sumToSend == 0 {
		return nil, fmt.Errorf("wrong params")
	}
	minimumNeededOutputs := make([]*generic.OutputRefWithAddrValue, 0)
	sum := uint64(0)
	for _, uo := range uos {
		if sum >= sumToSend {
			break
		}
		minimumNeededOutputs = append(minimumNeededOutputs, uo)
		sum += uo.Value
	}
	if sum < sumToSend {
		return nil, fmt.Errorf("not enough funds")
	}
	for _, outp := range minimumNeededOutputs {
		tx.Transfer().AddInput(value.NewInputFromOutputRef(&outp.OutputRef))
	}
	ret := make([]uint16, 0, len(minimumNeededOutputs))
	for _, amountToSend := range amounts {
		outIdx := tx.Transfer().AddOutput(value.NewOutput(addrTo, amountToSend))
		ret = append(ret, outIdx)
	}
	reminder := sum - sumToSend
	if reminder != 0 {
		tx.Transfer().AddOutput(value.NewOutput(addrFrom, reminder))
	}
	if len(amounts) != len(ret) {
		panic("len(amounts) != len(ret)")
	}
	return ret, nil
}
