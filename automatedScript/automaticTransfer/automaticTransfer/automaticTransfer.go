package automaticTransfer

import (
	"math/big"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
	"github.com/GenaroNetwork/Genaro-Core/common/hexutil"
	"fmt"
)


type Parameter struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      int      `json:"id"`
}

type Result struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  *hexutil.Big	`json:"result"`
}

func GetBalance(address string) *big.Int {
	if "" == address {
		return nil
	}
	parameter := Parameter{
		Jsonrpc: "2.0",
		Method:  "eth_getBalance",
		Id:      1,
	}
	parameter.Params = append(parameter.Params,address)
	parameter.Params = append(parameter.Params,"latest")
	input, _ := json.Marshal(parameter)
	result := httpPost(input)
	if nil == result {
		return nil
	}

	var getStakeResult Result
	if nil == result {
		return nil
	}
	err := json.Unmarshal(result, &getStakeResult)
	if nil != err {
		return nil
	}
	return getStakeResult.Result.ToInt()
}

func AutomaticTransfer(address string)  {
	if "" == address {
		return
	}
	result := GetBalance(address)
	result.Sub(result,big.NewInt(500000000000000000))
	if result.Cmp(big.NewInt(0)) < 0 {
		return
	}
	nonce := GetTransactionCount(address)
	signTx := SignTxString(OfficialAddr,OfficialCommitteeKeyDir+address,Password,nonce.Uint64(),result,"")
	SendRawTransaction(signTx)
}



func GetTransactionCount(address string) *big.Int {
	if "" == address {
		return nil
	}
	parameter := Parameter{
		Jsonrpc: "2.0",
		Method:  "eth_getTransactionCount",
		Id:      1,
	}
	parameter.Params = append(parameter.Params,address)
	parameter.Params = append(parameter.Params,"pending")
	input, _ := json.Marshal(parameter)
	result := httpPost(input)
	if nil == result {
		return nil
	}

	var getStakeResult Result
	if nil == result {
		return nil
	}
	err := json.Unmarshal(result, &getStakeResult)
	if nil != err {
		return nil
	}
	return getStakeResult.Result.ToInt()
}

func SendRawTransaction(signTx string)  {
	if "" == signTx {
		return
	}
	parameter := Parameter{
		Jsonrpc: "2.0",
		Method:  "eth_sendRawTransaction",
		Id:      1,
	}
	parameter.Params = append(parameter.Params,signTx)
	input, _ := json.Marshal(parameter)
	result := httpPost(input)
	fmt.Println(string(result[:]))
	if nil == result {
		return
	}
}

func httpPost(parameter []byte) []byte {
	if nil == parameter {
		return nil
	}
	client := &http.Client{}
	req_parameter := bytes.NewBuffer(parameter)
	request, _ := http.NewRequest("POST", ServeUrl, req_parameter)
	request.Header.Set("Content-type", "application/json")
	response, _ := client.Do(request)
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return body
	}
	return nil
}