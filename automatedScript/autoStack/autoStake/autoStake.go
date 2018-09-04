package autoStack

import (
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"github.com/GenaroNetwork/Genaro-Core/common/hexutil"
	"math/big"
	"fmt"

)

type GetMainAccountRankParameter struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      int      `json:"id"`
}


type MainAccountRankResult struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  []string
}


func GetMainAccountRank() []string {
	parameter := GetMainAccountRankParameter{
		Jsonrpc: "2.0",
		Method:  "eth_getMainAccountRank",
		Id:      1,
	}
	parameter.Params = append(parameter.Params,"latest")
	input, _ := json.Marshal(parameter)
	result := httpPost(input)
	if nil == result {
		return nil
	}

	var getMainAccountRankResult MainAccountRankResult
	if nil == result {
		return nil
	}
	err := json.Unmarshal(result, &getMainAccountRankResult)
	if nil != err {
		return nil
	}
	return getMainAccountRankResult.Result
}


type GetHeftParameter struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      int      `json:"id"`
}


type GetHeftResult struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  *hexutil.Big	`json:"result"`
}

func GetHeft(address string) *big.Int {
	if "" == address {
		return nil
	}
	parameter := GetHeftParameter{
		Jsonrpc: "2.0",
		Method:  "eth_getHeft",
		Id:      1,
	}
	parameter.Params = append(parameter.Params,address)
	parameter.Params = append(parameter.Params,"latest")
	input, _ := json.Marshal(parameter)
	result := httpPost(input)
	if nil == result {
		return nil
	}
	var getHeftResult GetHeftResult
	if nil == result {
		return nil
	}
	err := json.Unmarshal(result, &getHeftResult)
	if nil != err {
		return nil
	}
	return getHeftResult.Result.ToInt()
}



type GetStakeParameter struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      int      `json:"id"`
}


type GetStakeResult struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  *hexutil.Big	`json:"result"`
}

func GetStake(address string) *big.Int {
	if "" == address {
		return nil
	}
	parameter := GetHeftParameter{
		Jsonrpc: "2.0",
		Method:  "eth_getStake",
		Id:      1,
	}
	parameter.Params = append(parameter.Params,address)
	parameter.Params = append(parameter.Params,"latest")
	input, _ := json.Marshal(parameter)
	result := httpPost(input)
	if nil == result {
		return nil
	}

	var getStakeResult GetStakeResult
	if nil == result {
		return nil
	}
	err := json.Unmarshal(result, &getStakeResult)
	if nil != err {
		return nil
	}
	return getStakeResult.Result.ToInt()
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

type WeightResult struct {
	Heft *big.Int
	Stake  *big.Int
	Weight  float64
	Address string
	StakeWeight float64
	HeftWeight float64
	AddGNX int64
}

func CalculationWeight()([]WeightResult,*big.Int,*big.Int){
	getMainAccountRankResult := GetMainAccountRank()
	var weightResult []WeightResult
	var weight WeightResult
	if 0== len(getMainAccountRankResult){
		return nil, nil, nil
	}
	heftTotal := big.NewInt(0)
	stakeTotal := big.NewInt(0)
	for _,v := range getMainAccountRankResult{
		weight.Heft = GetHeft(v)
		weight.Stake = GetStake(v)
		weight.Address = v
		weightResult = append(weightResult, weight)
		heftTotal = heftTotal.Add(heftTotal,weight.Heft)
		stakeTotal = stakeTotal.Add(stakeTotal,weight.Stake)
	}
	for i:=0; i< len(weightResult); i++ {
		stake := weightResult[i].Stake
		heft := weightResult[i].Heft
		stakeWeight := float64(stake.Uint64())/float64(stakeTotal.Uint64()) //stake.Div(stake,stakeTotal)
		heftWeight := float64(heft.Uint64())/float64(stakeTotal.Uint64()) //heft.Div(heft,heftTotal)
		weightResult[i].StakeWeight = stakeWeight
		weightResult[i].HeftWeight = heftWeight
		weightResult[i].Weight = stakeWeight+heftWeight
	}
	return weightResult,heftTotal,stakeTotal
}

func CheckOfficialCommitteeWeight()  {
	weightResultTmp,_,stakeTotal := CalculationWeight()
	var weightResult []WeightResult
	for i:=0; i< len(weightResultTmp); i++{
		if weightResultTmp[i].Heft.Cmp(big.NewInt(0)) > 0 {
			weightResult =append(weightResult,weightResultTmp[i])
		}
	}
	if Ranking > len(weightResult) {
		return
	}
	aimsWeight := float64(weightResult[Ranking].Weight)
	var official []WeightResult
	for _, v := range  OfficialCommittee {
		for i:=0; i< len(weightResult); i++ {
			if v == weightResult[i].Address && weightResult[i].Weight < aimsWeight {
				addGNX := (aimsWeight - weightResult[i].HeftWeight)*float64(stakeTotal.Uint64()) - float64(weightResult[i].Stake.Uint64())
				weightResult[i].AddGNX = int64(addGNX)
				official = append(official, weightResult[i])
			}
		}
	}
	for _,v := range official {
		nonce := GetTransactionCount(v.Address)
		if v.Stake.Cmp(big.NewInt(5000)) <  0 {
			v.Stake = big.NewInt(5000)
		}
		extraData :=ExtraData{
			Address:v.Address,
			Type:"0x1",
			Stake:v.Stake,
		}
		signTx := SignTxString("0x6000000000000000000000000000000000000000",OfficialCommitteeKey,Password,nonce.Uint64(),extraData)
		SendRawTransaction(signTx)
	}
}

type ExtraData struct {
	Address string `json:"address"`
	Type  string `json:"type"`
	Stake *big.Int `json:"stake"`
}

func GetTransactionCount(address string) *big.Int {
	if "" == address {
		return nil
	}
	parameter := GetHeftParameter{
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

	var getStakeResult GetStakeResult
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
	parameter := GetHeftParameter{
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