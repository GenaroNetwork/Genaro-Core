package autoStack

import (
	"math/big"
	"context"
	"io/ioutil"

	"github.com/GenaroNetwork/Genaro-Core/accounts/keystore"
	"github.com/GenaroNetwork/Genaro-Core/cmd/utils"
	"github.com/GenaroNetwork/Genaro-Core/node"
	"github.com/GenaroNetwork/Genaro-Core/params"
	"github.com/GenaroNetwork/Genaro-Core/common"
	"github.com/GenaroNetwork/Genaro-Core/core/types"
	"encoding/json"
	"strings"
	"fmt"
)

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = "eth"
	cfg.Version = params.VersionWithCommit("")
	cfg.HTTPModules = append(cfg.HTTPModules, "eth", "shh")
	cfg.WSModules = append(cfg.WSModules, "eth", "shh")
	cfg.IPCPath = "geth.ipc"
	return cfg
}

func makeConfigNode(ctx context.Context) *node.Node {
	nodeConfig := defaultNodeConfig()
	stack, err := node.New(&nodeConfig)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	return stack
}

func SignTxString(address,keyDir,Password string,Nonce uint64,extraData ExtraData) string {
	keyJson, err := ioutil.ReadFile(keyDir)
	if err != nil {
		return ""
	}

	ctx := context.Background()
	stack := makeConfigNode(ctx)

	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	acct, err := ks.Import(keyJson, Password, Password)
	if err != nil {
		utils.Fatalf("%v", err)
	}

	//var Nonce uint64 = 0
	var To common.Address = common.HexToAddress(address)
	chain := big.NewInt(300)
	Value := new(big.Int).SetUint64(0)
	data, _ := json.Marshal(extraData)
	tx := types.NewTransaction(Nonce, To, Value, Gas, GasPrice, []byte(string(data)))
	signTx, err := ks.SignTxWithPassphrase(acct, Password, tx, chain)
	if err != nil {
		return ""
	}else {
		arr := strings.Split(signTx.String(),":")
		Hex := arr[len(arr)-1]
		return fmt.Sprintf("0x%s", strings.TrimSpace(Hex))
	}
}