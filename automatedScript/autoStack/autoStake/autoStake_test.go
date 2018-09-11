package autoStack

import (
	"testing"
	"fmt"
)

func TestGetBalance(t *testing.T) {
	result := GetMainAccountRank()
	fmt.Println(result)
}

func TestGetHeft(t *testing.T) {
	result := GetHeft("0x7fdead78e814124039abb62c5017ced1a031b53b")
	fmt.Println(result)
}

func TestGetStake(t *testing.T) {
	result := GetStake("0x7fdead78e814124039abb62c5017ced1a031b53b")
	fmt.Println(result)
}

func TestCalculationWeight(t *testing.T) {
	weightResult,_,_ := CalculationWeight()
	for i:=0; i< len(weightResult); i++{
		fmt.Println(weightResult[i])
	}
}



func TestCheckOfficialCommitteeWeight(t *testing.T) {
	CheckOfficialCommitteeWeight()
}