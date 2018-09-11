package automaticTransfer

import "math/big"

//server address
var ServeUrl string = "http://127.0.0.1:8549"
var OfficialCommittee= []string{"0x7fdead78e814124039abb62c5017ced1a031b53b", "0xe25de09fb1afd3cacfad2e91cf5d5f2862597667", "0xebb97ad3ca6b4f609da161c0b2b0eaa4ad58f3e8"}
var OfficialCommitteeKeyDir = "/home/qian/gopath/src/github.com/GenaroNetwork/Genaro-Core/automatedScript/automaticTransfer/keystore/"
var Password  = "123456"
var Gas uint64 = 100000
var GasPrice = new(big.Int).SetUint64(180000)
var OfficialAddr = "0xd248d954f6f87756dde05fba72b0a20f2517ae1c"