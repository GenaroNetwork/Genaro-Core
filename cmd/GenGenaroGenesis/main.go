package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/GenaroNetwork/Genaro-Core/common"
	"github.com/GenaroNetwork/Genaro-Core/common/hexutil"
	"github.com/GenaroNetwork/Genaro-Core/common/math"
	"github.com/GenaroNetwork/Genaro-Core/consensus/genaro"
	"github.com/GenaroNetwork/Genaro-Core/core"
	"github.com/GenaroNetwork/Genaro-Core/core/state"
	"github.com/GenaroNetwork/Genaro-Core/core/types"
	"github.com/GenaroNetwork/Genaro-Core/crypto"
	"github.com/GenaroNetwork/Genaro-Core/params"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"time"
)

var accountfile string

// 初始化相关参数
var PromissoryNoteEnable bool            // 是否分发期票
var PromissoryNotePercentage uint64      // 初始期票占比
var PromissoryNotePrice uint64           // 期票面值
var LastPromissoryNoteBlockNumber uint64 // 最后的期票返还块号
var PromissoryNotePeriod uint64          // 期票返还周期间隔
var SurplusCoin int64                    // 初始币池中的金额，单位GNX
var SynStateAccount string               // 用于同步
var HeftAccount string                   // 用于heft设置
var BindingAccount string                // 用于账号绑定
var OfficialAddress string               // 官方账号

func initarg() {
	flag.StringVar(&accountfile, "f", "account.json", "account file")
	flag.Parse()
}

func GenGenaroPriceAccount() core.GenesisAccount {
	var genaroPrice = types.GenaroPrice{
		BucketApplyGasPerGPerDay: (*hexutil.Big)(common.DefaultBucketApplyGasPerGPerDay),
		TrafficApplyGasPerG:      (*hexutil.Big)(common.DefaultTrafficApplyGasPerG),
		StakeValuePerNode:        (*hexutil.Big)(common.DefaultStakeValuePerNode),
		OneDayMortgageGes:        (*hexutil.Big)(common.DefaultOneDayMortgageGes),
		OneDaySyncLogGsaCost:     (*hexutil.Big)(common.DefaultOneDaySyncLogGsaCost),
		MaxBinding:               common.MaxBinding,
		MinStake:                 common.MinStake,
		CommitteeMinStake:        common.CommitteeMinStake,
		BackStackListMax:         common.BackStackListMax,
		CoinRewardsRatio:         common.CoinRewardsRatio,
		StorageRewardsRatio:      common.StorageRewardsRatio,
		RatioPerYear:             common.RatioPerYear,
		SynStateAccount:          SynStateAccount,
		HeftAccount:              HeftAccount,
		BindingAccount:           BindingAccount,
	}
	data, _ := json.Marshal(genaroPrice)
	GenaroPriceAccount := core.GenesisAccount{
		Balance:  big.NewInt(0),
		CodeHash: data,
	}
	return GenaroPriceAccount
}

func GenRewardsValuesAccount(surplusCoinUint int64) core.GenesisAccount {
	surplusCoin := big.NewInt(surplusCoinUint)
	surplusCoin.Mul(surplusCoin, common.BaseCompany)
	var rewardsValues = types.RewardsValues{
		CoinActualRewards:       big.NewInt(0),
		PreCoinActualRewards:    big.NewInt(0),
		StorageActualRewards:    big.NewInt(0),
		PreStorageActualRewards: big.NewInt(0),
		TotalActualRewards:      big.NewInt(0),
		SurplusCoin:             surplusCoin,
		PreSurplusCoin:          big.NewInt(0),
	}
	data, _ := json.Marshal(rewardsValues)
	RewardsValuesAccount := core.GenesisAccount{
		Balance:  big.NewInt(0),
		CodeHash: data,
	}
	return RewardsValuesAccount
}

// generate first committees list special account
func GenCandidateAccount(committees []common.Address) core.GenesisAccount {
	committeesData, _ := json.Marshal(committees)
	CandidateAccount := core.GenesisAccount{
		Balance:  big.NewInt(0),
		CodeHash: committeesData,
	}
	return CandidateAccount
}

// generate first SynState Account
func GenLastSynStateAccount() core.GenesisAccount {
	var lastRootStates = make(map[common.Hash]uint64)
	lastRootStates[common.Hash{}] = 0
	var lastSynState = types.LastSynState{
		LastRootStates:  lastRootStates,
		LastSynBlockNum: 0,
	}
	b, _ := json.Marshal(lastSynState)
	LastSynStateAccount := core.GenesisAccount{
		Balance:  big.NewInt(0),
		CodeHash: b,
	}
	return LastSynStateAccount
}

// generate Promissory Notes
// balance will edit
func GenPromissoryNotes(balance *big.Int, PromissoryNotePercentage uint64, PromissoryNotePrice uint64, LastPromissoryNoteBlockNumber uint64, PromissoryNotePeriod uint64) types.PromissoryNotes {
	var balanceGNX = big.NewInt(0)
	balanceGNXUint := balanceGNX.Div(balance, common.BaseCompany).Uint64()
	PromissoryNoteGNX := balanceGNXUint * PromissoryNotePercentage / 100
	if PromissoryNoteGNX > PromissoryNotePrice {
		PromissoryNoteNum := PromissoryNoteGNX / PromissoryNotePrice
		timeNum := LastPromissoryNoteBlockNumber / PromissoryNotePeriod
		notes := new(types.PromissoryNotes)
		for i := uint64(1); i <= PromissoryNoteNum; i++ {
			var note types.PromissoryNote
			note.Num = 1
			n := i % timeNum
			if n == 0 {
				note.RestoreBlock = timeNum * PromissoryNotePeriod
			} else {
				note.RestoreBlock = n * PromissoryNotePeriod
			}
			notes.Add(note)
		}

		allPromissoryNotePrice := big.NewInt(int64(notes.GetAllNum() * PromissoryNotePrice))
		allPromissoryNotePrice.Mul(allPromissoryNotePrice, common.BaseCompany)
		balance.Sub(balance, allPromissoryNotePrice)
		return *notes
	}
	return nil
}

// generate user account
func GenAccount(balanceStr string, stake, heft uint64) core.GenesisAccount {
	balance, ok := math.ParseBig256(balanceStr)
	if !ok {
		log.Fatal(errors.New("GenAccount ParseBig256 error"))
	}

	stakeLog := types.NumLog{
		BlockNum: 0,
		Num:      stake,
	}
	stakeLogs := types.NumLogs{stakeLog}

	heftLog := types.NumLog{
		BlockNum: 0,
		Num:      heft,
	}
	heftLogs := types.NumLogs{heftLog}

	var notes types.PromissoryNotes
	if PromissoryNoteEnable {
		notes = GenPromissoryNotes(balance, PromissoryNotePercentage, PromissoryNotePrice, LastPromissoryNoteBlockNumber, PromissoryNotePeriod)
	}

	genaroData := types.GenaroData{
		Stake:           stake,
		Heft:            heft,
		StakeLog:        stakeLogs,
		HeftLog:         heftLogs,
		PromissoryNotes: notes,
	}
	genaroDataByte, _ := json.Marshal(genaroData)
	account := core.GenesisAccount{
		Balance:  balance,
		CodeHash: genaroDataByte,
	}
	return account
}

func GenesisAllocToCandidateInfos(genesisAlloc core.GenesisAlloc) state.CandidateInfos {
	candidateInfos := make(state.CandidateInfos, 0)
	for addr, account := range genesisAlloc {
		var genaroData types.GenaroData
		json.Unmarshal(account.CodeHash, &genaroData)
		if genaroData.Stake > 0 {
			var candidateInfo state.CandidateInfo
			candidateInfo.Stake = genaroData.Stake
			candidateInfo.Heft = genaroData.Heft
			candidateInfo.Signer = addr
			candidateInfos = append(candidateInfos, candidateInfo)
		}
	}
	return candidateInfos
}

type account struct {
	Balance string `json:"balance"`
	Heft    uint64 `json:"heft"`
	Stake   uint64 `json:"stake"`
}
type MyAlloc map[common.Address]account

//type firstAccounts struct {
//	Alloc      FirstAlloc        `json:"alloc"      gencodec:"required"`
//}

type header struct {
	Encryption  string `json:"encryption"`
	Timestamp   int64  `json:"timestamp"`
	Key         string `json:"key"`
	Partnercode int    `json:"partnercode"`
}

func parseConfig(fileData []byte) {
	PromissoryNoteEnable = gjson.GetBytes(fileData, "config.PromissoryNoteEnable").Bool()
	fmt.Println("PromissoryNoteEnable:", PromissoryNoteEnable)
	PromissoryNotePercentage = gjson.GetBytes(fileData, "config.PromissoryNotePercentage").Uint()
	PromissoryNotePrice = gjson.GetBytes(fileData, "config.PromissoryNotePrice").Uint()
	LastPromissoryNoteBlockNumber = gjson.GetBytes(fileData, "config.LastPromissoryNoteBlockNumber").Uint()
	PromissoryNotePeriod = gjson.GetBytes(fileData, "config.PromissoryNotePeriod").Uint()
	SurplusCoin = gjson.GetBytes(fileData, "config.SurplusCoin").Int()
	fmt.Println("PromissoryNotePercentage:", PromissoryNotePercentage)
	fmt.Println("PromissoryNotePrice:", PromissoryNotePrice)
	fmt.Println("LastPromissoryNoteBlockNumber:", LastPromissoryNoteBlockNumber)
	fmt.Println("PromissoryNotePeriod:", PromissoryNotePeriod)
	fmt.Println("SurplusCoin:", SurplusCoin)

	SynStateAccount = gjson.GetBytes(fileData, "config.SynStateAccount").String()
	HeftAccount = gjson.GetBytes(fileData, "config.HeftAccount").String()
	BindingAccount = gjson.GetBytes(fileData, "config.BindingAccount").String()
	fmt.Println("SynStateAccount:", SynStateAccount)
	fmt.Println("HeftAccount:", HeftAccount)
	fmt.Println("BindingAccount:", BindingAccount)
	OfficialAddress = gjson.GetBytes(fileData, "config.OfficialAddress").String()
	fmt.Println("OfficialAddress:", OfficialAddress)
}

func main() {
	initarg()
	fileData, err := ioutil.ReadFile(accountfile)
	if err != nil {
		log.Fatal(err)
	}

	parseConfig(fileData)

	myAccounts := new(MyAlloc)
	accountStr := gjson.GetBytes(fileData, "accounts").String()
	json.Unmarshal([]byte(accountStr), myAccounts)

	genaroConfig := &params.ChainConfig{
		ChainId:        big.NewInt(300),
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(0),
		EIP150Hash:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		EIP155Block:    big.NewInt(0),
		EIP158Block:    big.NewInt(0),
		ByzantiumBlock: big.NewInt(0),
		Genaro: &params.GenaroConfig{
			Epoch:               43200,               //the number of blocks in one committee term
			Period:              2,                   // Number of seconds between blocks to enforce
			BlockInterval:       1,                   //a peer create BlockInterval blocks one time
			ElectionPeriod:      1,                   //a committee list write time
			ValidPeriod:         1,                   //a written committee list waiting time to come into force
			CurrencyRates:       4,                   //最少出块人数
			CommitteeMaxSize:    21,                  //max number of committee member
			OptionTxMemorySize:  20,                  //the number of save option tx
			PromissoryNotePrice: PromissoryNotePrice, // Promissory Note Price
			OfficialAddress:     OfficialAddress,
		},
	}
	genesis := new(core.Genesis)
	genesis.Config = genaroConfig
	genesis.Difficulty = big.NewInt(1)
	genesis.GasLimit = 20000000
	genesis.GasUsed = 0
	genesis.Mixhash = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")
	genesis.ParentHash = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")
	genesis.Timestamp = uint64(time.Now().Unix())
	genesis.Nonce = 0
	genesis.Coinbase = common.HexToAddress("0x0000000000000000000000000000000000000000")
	genesis.Alloc = make(core.GenesisAlloc, 1)

	// To write init Committee
	committees := make([]common.Address, 0)
	for addr := range *myAccounts {
		if (*myAccounts)[addr].Stake > 0 {
			committees = append(committees, addr)
		}
	}
	candidateAccount := GenCandidateAccount(committees)
	LastSynStateAccount := GenLastSynStateAccount()
	rewardsValuesAccount := GenRewardsValuesAccount(SurplusCoin)
	genaroPriceAccount := GenGenaroPriceAccount()
	genesis.Alloc[common.CandidateSaveAddress] = candidateAccount
	genesis.Alloc[common.LastSynStateSaveAddress] = LastSynStateAccount
	genesis.Alloc[common.RewardsSaveAddress] = rewardsValuesAccount
	genesis.Alloc[common.GenaroPriceAddress] = genaroPriceAccount
	//accounts := make([]core.GenesisAccount,len(*myAccounts))
	for addr := range *myAccounts {
		account := GenAccount((*myAccounts)[addr].Balance, (*myAccounts)[addr].Stake, (*myAccounts)[addr].Heft)
		genesis.Alloc[addr] = account
	}

	extra := new(genaro.ExtraData)
	var candidateInfos state.CandidateInfos
	candidateInfos = GenesisAllocToCandidateInfos(genesis.Alloc)
	extra.CommitteeRank, extra.Proportion = state.RankWithLenth(candidateInfos, int(genaroConfig.Genaro.CommitteeMaxSize), uint64(common.CommitteeMinStake))
	extraByte, _ := json.Marshal(extra)
	genesis.ExtraData = extraByte

	// create json file
	byt, err := json.Marshal(genesis)
	if err != nil {
		log.Fatal(err)
	}
	dirname, err := ioutil.TempDir(os.TempDir(), "genaro_test")
	genesisPath := dirname + "Genesis.json"
	fmt.Println(genesisPath)
	file, err := os.Create(genesisPath)
	if err != nil {
		log.Fatal(err)
	}
	file.Write(byt)
	file.Close()
}

func genAddrs(n int) []common.Address {
	addrs := make([]common.Address, 0)

	for i := 0; i < n; i++ {
		prikey, _ := crypto.GenerateKey()
		addr := crypto.PubkeyToAddress(prikey.PublicKey)

		fmt.Println(addr.String())
		addrs = append(addrs, addr)
	}
	return addrs
}
