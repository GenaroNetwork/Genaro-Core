package utils

import (
	"errors"
	"fmt"
	"github.com/GenaroNetwork/Genaro-Core/common"
	"github.com/GenaroNetwork/Genaro-Core/common/hexutil"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func checkError(ret []byte) error {
	errStr := gjson.GetBytes(ret, "error").String()
	if !strings.EqualFold("", errStr) {
		return errors.New(errStr)
	} else {
		return nil
	}
}

func HttpPost(url string, contentType string, body string) ([]byte, error) {
	bodyio := strings.NewReader(body)
	resp, err := http.Post(url, contentType, bodyio)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	repbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return repbody, nil
}

func GetCuBlockNum(url string) (uint64, error) {
	ret, err := HttpPost(url, "application/json", `{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber","params":[]}`)
	if err != nil {
		return 0, err
	}
	err = checkError(ret)
	if err != nil {
		return 0, err
	}
	blockNumStr := gjson.GetBytes(ret, "result").String()
	blockNum, err := hexutil.DecodeUint64(blockNumStr)
	if err != nil {
		return 0, err
	}
	return blockNum, nil
}

func GetBlockByNumber(url string, blockNum uint64) ([]byte, error) {
	blockNumHex := hexutil.EncodeUint64(blockNum)
	ret, err := HttpPost(url, "application/json", `{"jsonrpc":"2.0","id":1,"method":"eth_getBlockByNumber","params":["`+blockNumHex+`",true]}`)
	if err != nil {
		return nil, err
	}
	err = checkError(ret)
	if err != nil {
		return nil, err
	}
	return ret, err
}

func GetBlockHash(url string, blockNum uint64) (string, error) {
	ret, err := GetBlockByNumber(url, blockNum)
	if err != nil {
		return "", err
	}
	blockHash := gjson.GetBytes(ret, "result.hash").String()
	return blockHash, nil
}

func SendSynState(url string, blockHash string, SynStateAccount string) (string, error) {
	ret, err := HttpPost(url, "application/json", `{"jsonrpc": "2.0","method": "eth_sendTransaction","params": [{"from": "`+SynStateAccount+`","to": "`+common.SpecialSyncAddress.String()+`","gasPrice": "0x430e23400","value": "0x0","extraData": "{\"msg\": \"`+blockHash+`\",\"type\": \"0xd\"}"}],"id": 1}`)
	if err != nil {
		return "", err
	}
	err = checkError(ret)
	if err != nil {
		return "", err
	}
	return gjson.ParseBytes(ret).String(), nil
}

func GetLastSynBlockInfo(url string) ([]byte, error) {
	ret, err := HttpPost(url, "application/json", `{"jsonrpc":"2.0","method":"eth_getLastSynBlock","params":["latest"],"id":1}`)
	err = checkError(ret)
	if err != nil {
		return nil, err
	}
	return ret, err
}

func GetLastSynBlockHash(url string) (string, error) {
	ret, err := GetLastSynBlockInfo(url)
	if err != nil {
		return "", err
	}
	hash := gjson.GetBytes(ret, "result.BlockHash").String()
	return hash, nil
}

func CheckTransaction(url string, txHash string) (bool, error) {
	if strings.EqualFold("", txHash) {
		return true, nil
	}
	ret, err := HttpPost(url, "application/json", `{"jsonrpc":"2.0","id":10,"method":"eth_getTransactionByHash","params":["`+txHash+`"]}`)
	err = checkError(ret)
	if err != nil {
		return false, err
	}
	result := gjson.GetBytes(ret, "result").String()
	if strings.EqualFold(result, "") {
		return false, nil
	}
	return true, nil
}

func CheckRecipt(url string, txHash string) (bool, error) {
	if strings.EqualFold("", txHash) {
		return true, nil
	}
	ret, err := HttpPost(url, "application/json", `{"jsonrpc":"2.0","id":10,"method":"eth_getTransactionReceipt","params":["`+txHash+`"]}`)
	err = checkError(ret)
	if err != nil {
		return false, err
	}
	result := gjson.GetBytes(ret, "result").String()
	if strings.EqualFold(result, "") {
		return true, nil
	}
	//fmt.Println(result)
	status := hexutil.MustDecodeUint64(gjson.GetBytes(ret, "result.status").String())
	if status == 0 {
		return false, nil
	}
	return true, nil
}

func AccountUnlock(url string, SynStateAccount string, SynStateAccountPasswd string) error {
	ret, err := HttpPost(url, "application/json", `{"jsonrpc": "2.0","method": "personal_unlockAccount","params": ["`+SynStateAccount+`","`+SynStateAccountPasswd+`",null],"id": 1}`)
	if err != nil {
		return err
	}
	err = checkError(ret)
	if err != nil {
		return err
	}
	return nil
}

func GetCrossTxList(url string, start int64, end int64) ([]byte, error) {
	ret, err := HttpPost(url, "application/json", `{"jsonrpc":"2.0","method":"eth_getCrossChainByBlockNumberRange","params":["`+DecHexStr(start)+`","`+DecHexStr(end)+`"],"id":1}`)
	if err != nil {
		return ret, err
	}
	err = checkError(ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// 0x...
func DecHexStr(n int64) string {
	return "0x" + DecHex(n)
}

func DecHex(n int64) string {
	if n < 0 {
		log.Println("Decimal to hexadecimal error: the argument must be greater than zero.")
		return ""
	}
	if n == 0 {
		return "0"
	}
	hex := map[int64]int64{10: 65, 11: 66, 12: 67, 13: 68, 14: 69, 15: 70}
	s := ""
	for q := n; q > 0; q = q / 16 {
		m := q % 16
		if m > 9 && m < 16 {
			m = hex[m]
			s = fmt.Sprintf("%v%v", string(m), s)
			continue
		}
		s = fmt.Sprintf("%v%v", m, s)
	}
	return s
}

func GetCrossTxStatusByAcount(url string, account string) ([]byte, error) {
	ret, err := HttpPost(url, "application/json", `{"jsonrpc":"2.0","method":"eth_getCrossChain","params":["`+account+`","latest"],"id":1}`)
	if err != nil {
		return ret, err
	}
	err = checkError(ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func GetCrossTxStatus(url string, account string, hash string) (bool, error) {
	ret, err := GetCrossTxStatusByAcount(url, account)
	if err != nil {
		return false, err
	}
	resultBuf := gjson.GetBytes(ret, "result")
	result := resultBuf.Get(hash)
	if result.Exists() {
		status := result.Get("type").Bool()
		return status, nil
	}

	return false, nil
}

func CrossSettlementTx(url string, speAccount string, account string, hash string) error {
	ret, err := HttpPost(url, "application/json", `{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"`+speAccount+`","to":"0x6000000000000000000000000000000000000000","gas":"0x345000","gasPrice":"0x91a01","value":"0x0","data":"","extraData":"{ \"cross_chain\": { \"hash\": \"`+hash+`\", \"from_address\": \"`+account+`\" },\"type\":\"0x30\"}"}],"id":1}`)
	if err != nil {
		return err
	}
	err = checkError(ret)
	if err != nil {
		return err
	}
	return nil
}
