package types

import (
	"github.com/GenaroNetwork/Genaro-Core/common"
	"strings"
	"testing"
)

func TestBindingTable(t *testing.T) {
	var bindingTable = BindingTable{
		MainAccounts: make(map[common.Address][]common.Address),
		SubAccounts:  make(map[common.Address]common.Address),
	}
	bindingTable.UpdateBinding(common.HexToAddress("0x1000000000000000000000000000000000000000"), common.HexToAddress("0x1100000000000000000000000000000000000000"))
	bindingTable.UpdateBinding(common.HexToAddress("0x1000000000000000000000000000000000000000"), common.HexToAddress("0x1200000000000000000000000000000000000000"))
	bindingTable.UpdateBinding(common.HexToAddress("0x1000000000000000000000000000000000000000"), common.HexToAddress("0x1200000000000000000000000000000000000000"))
	bindingTable.UpdateBinding(common.HexToAddress("0x1000000000000000000000000000000000000000"), common.HexToAddress("0x1300000000000000000000000000000000000000"))
	t.Log(bindingTable)
	if bindingTable.GetSubAccountSizeInMainAccount(common.HexToAddress("0x1000000000000000000000000000000000000000")) != 3 {
		t.Error("UpdateBinding error")
	}
	bindingTable.DelSubAccount(common.HexToAddress("0x1300000000000000000000000000000000000000"))
	t.Log(bindingTable)
	if bindingTable.GetSubAccountSizeInMainAccount(common.HexToAddress("0x1000000000000000000000000000000000000000")) != 2 {
		t.Error("UpdateBinding error")
	}
	bindingTable.UpdateBinding(common.HexToAddress("0x1000000000000000000000000000000000000000"), common.HexToAddress("0x1300000000000000000000000000000000000000"))
	if !bindingTable.IsAccountInBinding(common.HexToAddress("0x1000000000000000000000000000000000000000")) {
		t.Error("Binding error")
	}
	if bindingTable.IsAccountInBinding(common.HexToAddress("0x1000000000000000000000000000000000111111")) {
		t.Error("Binding error")
	}
	if !bindingTable.IsAccountInBinding(common.HexToAddress("0x1300000000000000000000000000000000000000")) {
		t.Error("Binding error")
	}
	bindingTable.UpdateBinding(common.HexToAddress("0x2000000000000000000000000000000000000000"), common.HexToAddress("0x1300000000000000000000000000000000000000"))
	if bindingTable.GetSubAccountSizeInMainAccount(common.HexToAddress("0x1000000000000000000000000000000000000000")) != 2 {
		t.Error("UpdateBinding error")
	}
	if bindingTable.GetSubAccountSizeInMainAccount(common.HexToAddress("0x2000000000000000000000000000000000000000")) != 1 {
		t.Error("UpdateBinding error")
	}
	t.Log(bindingTable)
	bindingTable.DelSubAccount(common.HexToAddress("0x1300000000000000000000000000000000000000"))
	if bindingTable.IsMainAccountExist(common.HexToAddress("0x2000000000000000000000000000000000000000")) {
		t.Error("DelSubAccount error")
	}
	t.Log(bindingTable)
	subAccounts := bindingTable.DelMainAccount(common.HexToAddress("0x1000000000000000000000000000000000000000"))
	if bindingTable.IsMainAccountExist(common.HexToAddress("0x1000000000000000000000000000000000000000")) {
		t.Error("DelSubAccount error")
	}
	t.Log(bindingTable)
	if len(subAccounts) != 2 {
		t.Error("DelMainAccount error")
	}
	t.Log(subAccounts)
}

func (notes *PromissoryNotes) Print(t *testing.T) {
	for _, note := range *notes {
		t.Log("RestoreBlock:", note.RestoreBlock)
		t.Log("Num:", note.Num)
	}
}

func TestPromissoryNote(t *testing.T) {
	note1 := PromissoryNote{
		RestoreBlock: 10,
		Num:          20,
	}
	note2 := PromissoryNote{
		RestoreBlock: 20,
		Num:          30,
	}
	note3 := PromissoryNote{
		RestoreBlock: 30,
		Num:          40,
	}
	note4 := PromissoryNote{
		RestoreBlock: 10,
		Num:          50,
	}
	note5 := PromissoryNote{
		RestoreBlock: 20,
		Num:          60,
	}

	notes := new(PromissoryNotes)
	notes.Add(note1)
	notes.Add(note2)
	notes.Add(note3)
	notes.Add(note4)
	notes.Add(note5)

	notes.Print(t)
	t.Log(notes.GetNum(20))
	notes.Del(note2)
	t.Log(notes.GetNum(20))
	notes.Del(note5)
	t.Log(notes.GetNum(20))
	t.Log(notes.GetNum(30))
	notes.Print(t)

	t.Log(notes.DelBefor(10))
	notes.Print(t)

	notes.Add(note1)
	notes.Add(note2)
	notes.Print(t)
	t.Log(notes.GetAllNum())
	t.Log(notes.GetBefor(20))
	t.Log(notes.DelBefor(20))
	notes.Print(t)

}

func TestAccountName(t *testing.T) {
	var name AccountName
	err := name.SetString("myname")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(name.String())
	hash := name.ToHash()
	t.Log(hash.String())
	var name2 AccountName
	name2.SetHash(hash)
	t.Log(name2.String())
	if !strings.EqualFold(name.String(), name2.String()) {
		t.Fatal("AccountName trans error")
	}
	err = name2.SetString("myname.TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT")
	if err != nil {
		t.Log(err)
	}
	name.SetString("asd.efwf..")
	if !name.IsValid() {
		t.Log("test1 name is invalid")
	}
	name.SetString("asd.efwf.23414.ttt")
	if !name.IsValid() {
		t.Log("test2 name is invalid")
	}
	name.SetString("qwd12EE@dw")
	if !name.IsValid() {
		t.Log("test3 name is invalid")
	}
	name.SetString("qwd1 2EE@dw")
	if !name.IsValid() {
		t.Log("test4 name is invalid")
	}
}
