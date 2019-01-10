package ethapi

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

func httpPost(parameter []byte) []byte {
	if nil == parameter {
		return nil
	}
	client := &http.Client{}
	req_parameter := bytes.NewBuffer(parameter)
	request, _ := http.NewRequest("POST", "https://kyc.gxchain.org/auth/token", req_parameter)
	request.Header.Set("Content-type", "application/json")
	response, _ := client.Do(request)
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return body
	}
	return nil
}


type GetTokenResult struct {
	Data  GetUrl  `json:"data"`
	RetMsg string `json:"retMsg"`
	RetCode int `json:"retCode"`
}

type GetUrl struct {
	KycUrl  string `json:"kycUrl"`
	Token	string `json:"token"`
}

type PublicKeyArgs struct {
	PublicKey      string  `json:"publicKey"`

}

func GetToken( addr string) string {
	fmt.Println(addr)
	parameter := PublicKeyArgs{
		PublicKey:addr,
	}
	input, _ := json.Marshal(parameter)
	result := httpPost(input)

	var getTokenResult GetTokenResult

	err := json.Unmarshal(result, &getTokenResult)
	if nil != err {
		fmt.Println(err)
		return "get token error"
	}

	return getTokenResult.Data.KycUrl
}