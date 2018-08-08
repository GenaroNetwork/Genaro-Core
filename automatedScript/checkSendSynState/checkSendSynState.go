package main
import (
	"encoding/json"
	"io/ioutil"
	_"math/big"
	"fmt"
	"github.com/GenaroNetwork/Genaro-Core/common/hexutil"
	"net/http"
	"bytes"
	"os/exec"
)

type parameter struct {
	Jsonrpc string `json:"jsonrpc"`
	Result *hexutil.Big `json:"result"`
}

func main() {
	JsonParse := NewJsonStruct()
	configParameter := parameter{}
	JsonParse.Load("./checkSendSynStateInfo", &configParameter)
	getBlockNumber := GetBlockNumber()
	parameterJson,err := json.Marshal(getBlockNumber)
	fmt.Println(getBlockNumber.Result)
	fmt.Println(configParameter.Result)
	if nil != err {
		fmt.Println("json.Marshal(parameter) error")
		return
	}
	if ioutil.WriteFile("./checkSendSynStateInfo",parameterJson, 0644) != nil {
		fmt.Println("parameterJson write failed")
		return
	}

	if getBlockNumber.Result.ToInt().Cmp(configParameter.Result.ToInt()) > 0 {
		return
	}
	cmd := exec.Command("/bin/bash", "-c", "./checkSendSynState.sh")
	cmd.Run()
}

func GetBlockNumber() parameter {
	blockNumberResult := parameter{}
	result := httpPost([]byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":83}`))
	if nil == result {
		return parameter{}
	}
	json.Unmarshal(result, &blockNumberResult)
	return blockNumberResult
}



func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

type JsonStruct struct {
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	fmt.Println("###########")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(data, v)
}

var ServeUrl string = "http://127.0.0.1:8550"

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