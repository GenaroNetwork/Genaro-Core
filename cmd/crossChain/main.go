package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/GenaroNetwork/Genaro-Core/cmd/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"time"
)

var rpcurl1 string
var rpcurl2 string
var MYSQLUSER string
var MYSQLPASSWD string
var MYSQLIP string
var MYSQLPORT int
var MYSQLDB string
var BEGIN int
var start1 uint64
var start2 uint64
var speAccount string
var speAccountPasswd string

func initarg() {
	flag.StringVar(&rpcurl1, "rpc1", "http://127.0.0.1:8545", "rpcurl1")
	flag.StringVar(&rpcurl2, "rpc2", "http://127.0.0.1:8545", "rpcurl2")
	flag.StringVar(&MYSQLUSER, "muser", "user", "mysql user")
	flag.StringVar(&MYSQLPASSWD, "mpasswd", "passwd", "mysql password")
	flag.StringVar(&MYSQLIP, "mip", "127.0.0.1", "mysql ip")
	flag.IntVar(&MYSQLPORT, "mport", 3306, "mysql port")
	flag.StringVar(&MYSQLDB, "mdb", "db", "mysql database")
	flag.IntVar(&BEGIN, "begin", 0, "scan begin block")
	flag.StringVar(&speAccount, "addr", "addr", "special account")
	flag.StringVar(&speAccountPasswd, "pwd", "pwd", "special account password")
	flag.Parse()
}

func main() {
	initarg()
	mysqlStr := MYSQLUSER + ":" + MYSQLPASSWD + "@tcp(" + MYSQLIP + ":" + strconv.Itoa(MYSQLPORT) + ")/" + MYSQLDB
	db, err := sql.Open("mysql", mysqlStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	go deal(db)
	//go Scan(db)

	select {}
}

func deal(db *sql.DB) {
	for {
		dealCrossTask(db)
		time.Sleep(5 * time.Second)
	}

}

func dealCrossTask(db *sql.DB) {
	crossTaskList, err := GetAllNotDealTask(db)
	if err != nil {
		log.Println(err)
		return
	}
	for _, crossTask := range crossTaskList {
		fmt.Println(crossTask)
		match, err := crossTask.IsMatch(db)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(match)
		if match {
			ok := updateStatus(db, crossTask)
			if !ok {
				err = utils.AccountUnlock(getRpcurl(crossTask), speAccount, speAccountPasswd)
				if err != nil {
					log.Println(err)
				}
				err = utils.CrossSettlementTx(getRpcurl(crossTask), speAccount, crossTask.From_address, crossTask.Hash)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func getRpcurl(crossTask CrossTask) string {
	rpcurl := rpcurl1
	if crossTask.Chainid == 2 {
		rpcurl = rpcurl2
	}
	return rpcurl
}

func updateStatus(db *sql.DB, crossTask CrossTask) bool {
	rpcurl := getRpcurl(crossTask)
	ok, err := utils.GetCrossTxStatus(rpcurl, crossTask.From_address, crossTask.Hash)
	if err != nil {
		log.Println(err)
		return false
	}
	if ok {
		crossTask.Status = 1
		crossTask.Save(db)
		return ok
	}
	return false
}

func Scan(db *sql.DB) {
	for {
		ScanChain1(db)
		ScanChain2(db)
		time.Sleep(1 * time.Minute)
	}

}

func ScanChain1(db *sql.DB) {
	start1 = uint64(BEGIN)
	blockNum, err := utils.GetCuBlockNum(rpcurl1)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		if start1 >= blockNum-12 {
			log.Println("Chain1 scan to end")
			return
		} else if start1+100 < blockNum-12 {
			ScanTx1(db, int64(start1), int64(start1+100))
			start1 = start1 + 100
		} else {
			ScanTx1(db, int64(start1), int64(blockNum-12))
			start1 = blockNum - 12
		}
	}
	return
}

func ScanChain2(db *sql.DB) {
	start1 = uint64(BEGIN)
	blockNum, err := utils.GetCuBlockNum(rpcurl2)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		if start2 >= blockNum-12 {
			log.Println("Chain2 scan to end")
			return
		} else if start2+100 < blockNum-12 {
			ScanTx2(db, int64(start2), int64(start2+100))
			start2 = start2 + 100
		} else {
			ScanTx2(db, int64(start2), int64(blockNum-12))
			start2 = blockNum - 12
		}
	}
	return
}

func ScanTx1(db *sql.DB, start int64, end int64) {
	ret, err := utils.GetCrossTxList(rpcurl1, start, end)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(ret))
	resultBuf := gjson.GetBytes(ret, "result")
	results := resultBuf.Array()
	for _, result := range results {
		var crossTask CrossTask
		fmt.Println(result.String())
		crossTask.Hash = gjson.Get(result.String(), "hash").String()
		crossTask.From_address = gjson.Get(result.String(), "from_address").String()
		crossTask.To_address = gjson.Get(result.String(), "address").String()
		crossTask.Amount = gjson.Get(result.String(), "amount").String()
		crossTask.Chainid = 1
		crossTask.Status = 0
		exist, err := crossTask.IsExist(db)
		if err != nil {
			log.Println(err)
			return
		}
		if !exist {
			err := crossTask.Save(db)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func ScanTx2(db *sql.DB, start int64, end int64) {
	ret, err := utils.GetCrossTxList(rpcurl2, start, end)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(ret))
	resultBuf := gjson.GetBytes(ret, "result")
	results := resultBuf.Array()
	for _, result := range results {
		var crossTask CrossTask
		fmt.Println(result.String())
		crossTask.Hash = gjson.Get(result.String(), "hash").String()
		crossTask.From_address = gjson.Get(result.String(), "from_address").String()
		crossTask.To_address = gjson.Get(result.String(), "address").String()
		crossTask.Amount = gjson.Get(result.String(), "amount").String()
		crossTask.Chainid = 2
		crossTask.Status = 0
		exist, err := crossTask.IsExist(db)
		if err != nil {
			log.Println(err)
			return
		}
		if !exist {
			err := crossTask.Save(db)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
