package main

import (
	"database/sql"
	"strconv"
)

type CrossTask struct {
	Hash         string
	From_address string
	To_address   string
	Amount       string
	Chainid      int
	Status       int
}

func (crossTask *CrossTask) Save(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO `crossTask` (Hash,From_address,To_address,Amount,Chainid,Status) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE Status=?;", strconv.Itoa(crossTask.Chainid)+crossTask.Hash, crossTask.From_address, crossTask.To_address, crossTask.Amount, crossTask.Chainid, crossTask.Status, crossTask.Status)
	if err != nil {
		return err
	}
	return nil
}

func (crossTask *CrossTask) IsExist(db *sql.DB) (bool, error) {
	rows, err := db.Query("SELECT Hash FROM `crossTask` where Hash=?;", strconv.Itoa(crossTask.Chainid)+crossTask.Hash)
	if err != nil {
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	} else {
		return false, nil
	}
}

func GetCrossTaskByHash(db *sql.DB,hash string) (*CrossTask,error) {
	rows, err := db.Query("SELECT Hash,From_address,To_address,Amount,Chainid,Status FROM `crossTask` where `Hash`=?;",hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var crossTask CrossTask
		rows.Scan(&crossTask.Hash, &crossTask.From_address, &crossTask.To_address, &crossTask.Amount, &crossTask.Chainid, &crossTask.Status)
		crossTask.Hash = crossTask.Hash[1:]
		return &crossTask, nil
	} else {
		return nil, nil
	}
}

// 获取所有任务
func GetAllNotDealTask(db *sql.DB) ([]CrossTask, error) {
	rows, err := db.Query("SELECT Hash,From_address,To_address,Amount,Chainid,Status FROM `crossTask` where `Status`=0;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	crossTaskList := make([]CrossTask, 0)
	for rows.Next() {
		var crossTask CrossTask
		rows.Scan(&crossTask.Hash, &crossTask.From_address, &crossTask.To_address, &crossTask.Amount, &crossTask.Chainid, &crossTask.Status)
		crossTask.Hash = crossTask.Hash[1:]
		crossTaskList = append(crossTaskList, crossTask)
	}
	return crossTaskList, nil
}

func SwitchChainId(id int ) int{
	if id == 1 {
		return 2
	} else {
		return 1
	}
}

func (crossTask *CrossTask) IsMatch(db *sql.DB) (bool, error) {
	thatid:=SwitchChainId(crossTask.Chainid)
	crossTask2,err := GetCrossTaskByHash(db,strconv.Itoa(thatid)+crossTask.Hash)
	if err != nil {
		return false,err
	}
	if crossTask2 == nil {
		return false,err
	}
	return true,err
}
