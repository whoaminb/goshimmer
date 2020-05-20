package utxodb

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/mr-tron/base58"
)

const (
	genesisPrivateKeyStr = "1pK9KraR4YTSHh3bq7hrigFSyq4HgWufhRyME84DPbwWpcoF1zwq6J1zaeyYUb8ut6ia9uQ9B9ughrpj2aZ7CMU"
	genesisPublicKeyStr  = "4f1W4o6PBKXXFsMHRcndYabpmaXPTdNNmtnV2NkfHnAa"
	supply               = int64(10 * 1000 * 1000 * 1000)

	ownerPrivKey1 = "2sHwx4pBsAxK5SV8Nurphsgu6XpRE2ZpBxJQ87ggdoU34NAQ5RPM6u9K8oL4UdbYgv1pRiHNHCsWipbY1TvEuXz6"
	ownerPubKey1  = "C3XgXzR6dVCsnXYqxJC5WgMEQJUDqeChUMKY9Tuvfoux"
	ownerAddress1 = "b1uERj4r1wuuVeTx4n7t8JnsuSF5YnZPf8CQdCzv6EHs"

	ownerPrivKey2 = "6788yntE9kZeg9H7VLyG9F7pjTMpmQm3MbtSGVgcYKQ9AmxpwJf6oZsiuiDb6qMwpEJeofT28M1gu2c7iqKPCgag"
	ownerPubKey2  = "AsYBP4Yd41e3noTx9rzZKekwnZ1wkPokgb7i7RKKdUKN"
	ownerAddress2 = "VSVggSWMLdT5xcXpyd1yQAJKMKzEwqP2Ff65KHFbCzjD"

	ownerPrivKey3 = "2CESGXeY11kmRec7et5ALBeDDKzyG7YEQXbVu4Z1XYiVeeZbYJSrrhtLuiggk4EfGWeFc2G33kWeeHTuhoGk9dXU"
	ownerPubKey3  = "8GrpoaDAV3om7Ufa8frapCfDr6pprN3zsBNzdh8uFxKg"
	ownerAddress3 = "S5ugKUKvntBi66g44qGTa3jcSgDh6QxYHgAhJ5s5m61x"
)

var (
	genesisTxId transaction.ID

	knownAddresses  []address.Address
	knownSigSchemes map[address.Address]signaturescheme.SignatureScheme
)

func init() {
	knownAddresses = make([]address.Address, 4)
	knownSigSchemes = make(map[address.Address]signaturescheme.SignatureScheme)

	genesis := createSigScheme(genesisPrivateKeyStr, genesisPublicKeyStr, "")
	knownAddresses[0] = genesis.Address()
	knownSigSchemes[genesis.Address()] = genesis

	sigs := createSigScheme(ownerPrivKey1, ownerPubKey1, ownerAddress1)
	knownSigSchemes[sigs.Address()] = sigs
	knownAddresses[1] = sigs.Address()

	sigs = createSigScheme(ownerPrivKey2, ownerPubKey2, ownerAddress2)
	knownSigSchemes[sigs.Address()] = sigs
	knownAddresses[2] = sigs.Address()

	sigs = createSigScheme(ownerPrivKey3, ownerPubKey3, ownerAddress3)
	knownSigSchemes[sigs.Address()] = sigs
	knownAddresses[3] = sigs.Address()

	// create genesis transaction

	var niltxid transaction.ID

	genesisInput := transaction.NewOutputID(GetGenesisAddress(), niltxid)
	inputs := transaction.NewInputs(genesisInput)
	outputs := transaction.NewOutputs(map[address.Address][]*balance.Balance{
		GetGenesisAddress(): {balance.New(balance.ColorIOTA, supply)},
	})
	genesisTx := transaction.New(inputs, outputs)
	genesisTx.Sign(GetGenesisSigScheme())

	genesisTxId = genesisTx.ID()

	transactions[genesisTxId] = genesisTx
	utxo[transaction.NewOutputID(GetGenesisAddress(), genesisTxId)] = true
	utxoByAddress[GetGenesisAddress()] = []transaction.ID{genesisTxId}

	fmt.Printf("UTXODB initialized. Genesis: address = %s, tx id = %s supply = %d\n",
		GetGenesisAddress().String(), GetGenesisTransaction().ID().String(), GetSupply())
}

func GetSupply() int64 {
	return supply
}

func GetGenesisSigScheme() signaturescheme.SignatureScheme {
	return knownSigSchemes[GetGenesisAddress()]
}

func GetGenesisTransaction() *transaction.Transaction {
	ret, ok := GetTransaction(genesisTxId)
	if !ok {
		panic("genesis tx not found")
	}
	return ret
}

func GetGenesisAddress() address.Address {
	return GetAddress(0)
}

func GetAddress(i int) address.Address {
	return knownAddresses[i]
}

func GetSigScheme(addr address.Address) signaturescheme.SignatureScheme {
	return knownSigSchemes[addr]
}

func createSigScheme(privKeyStr, pubKeyStr, addressStr string) signaturescheme.SignatureScheme {
	var privKey ed25519.PrivateKey
	var pubKey ed25519.PublicKey
	var err error

	priv, err := base58.Decode(privKeyStr)
	if err != nil || len(priv) != len(privKey) {
		panic(err)
	}
	pub, err := base58.Decode(pubKeyStr)
	if err != nil || len(pub) != len(pubKey) {
		panic(err)
	}
	copy(privKey[:], priv)
	copy(pubKey[:], pub)
	keyPair := ed25519.KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
	ret := signaturescheme.ED25519(keyPair)

	if addressStr == "" {
		return ret
	}
	addr, err := address.FromBase58(addressStr)
	if err != nil {
		panic(err)
	}
	if addr != ret.Address() {
		panic("addr != ret.Address()")
	}
	return ret
}
