package main

import "github.com/GenaroNetwork/Genaro-Core/automatedScript/automaticTransfer/automaticTransfer"

func main() {
	for _,v := range automaticTransfer.OfficialCommittee {
		automaticTransfer.AutomaticTransfer(v)
	}
}
