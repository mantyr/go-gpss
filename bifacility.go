// Copyright 2019 Sergey Soldatov. All rights reserved.
// This software may be modified and distributed under the terms
// of the Apache license. See the LICENSE file for details.

package gpss

// A Bifacility as facility, but without advance in it and present in two parts,
// first for takes ownership of a Facility, second for release ownership of a Facility

import (
	"fmt"
)

// The first part of a Bifacility, it takes ownership of a Facility
type InFacility struct {
	BaseObj
	// Holded transast ID
	HoldedTransactID int
	// For backuping Facility/Bifacility name if we includes Bifacility in Bifacility
	bakupFacilityName string
	// For counting the transacts that go through Bifacility
	cnt_transact float64
	// For counting the advance of transact
	sum_advance float64
	// For saving time of input transact in Bifacility
	timeOfInput int
}

// The second part of a Bifacility, for release ownership of a Facility
type OutFacility struct {
	BaseObj
	// Pointer to inFacility structure
	inFacility *InFacility
}

// Creates new Bifacility (InFacility + OutFacility).
// name - name of object
func NewBifacility(name string) (*InFacility, *OutFacility) {
	inObj := &InFacility{}
	inObj.BaseObj.Init(name)
	outObj := &OutFacility{}
	outObj.name = name + "_OUT"
	outObj.tb = inObj.tb
	outObj.inFacility = inObj
	return inObj, outObj
}

func (obj *InFacility) HandleTransact(transact ITransaction) {
	transact.PrintInfo()
	for _, v := range obj.GetDst() {
		if v.AppendTransact(transact) {
			return
		}
	}
}

func (obj *InFacility) AppendTransact(transact ITransaction) bool {
	if obj.tb.GetLen() != 0 {
		// Facility is busy
		return false
	}
	obj.GetLogger().GetTrace().Println("Append transact ", transact.GetId(), " to Facility")
	transact.SetHolderName(obj.name)
	if transact.GetParameterByName("Facility") != nil {
		obj.bakupFacilityName = transact.GetParameterByName("Facility").(string)
	}
	transact.SetParameters([]Parameter{{Name: "Facility", Value: obj.name}})
	obj.HoldedTransactID = transact.GetId()
	obj.tb.Push(transact)
	obj.cnt_transact++
	obj.timeOfInput = obj.GetPipeline().GetModelTime()
	obj.HandleTransact(transact)
	return true
}

func (obj *InFacility) PrintReport() {
	obj.BaseObj.PrintReport()
	avr := obj.sum_advance / obj.cnt_transact
	fmt.Printf("Average advance %.2f \tAverage utilization %.2f%%\tNumber entries %.2f \t", avr,
		100*avr*obj.cnt_transact/float64(obj.GetPipeline().GetSimTime()), obj.cnt_transact)
	if obj.HoldedTransactID > 0 {
		fmt.Print("Transact ", obj.HoldedTransactID, " in facility")
	} else {
		fmt.Print("Facility is empty")
	}
	fmt.Printf("\n\n")
}

func (obj *InFacility) IsEmpty() bool {
	if obj.tb.GetLen() != 0 {
		// Facility is busy
		return false
	}
	return true
}

func (obj *OutFacility) HandleTransact(transact ITransaction) {
	transact.PrintInfo()
	if obj.inFacility.bakupFacilityName != "" {
		transact.SetParameters([]Parameter{{Name: "Facility",
			Value: obj.inFacility.bakupFacilityName}})
	} else {
		transact.SetParameters([]Parameter{{Name: "Facility", Value: nil}})
	}

	for _, v := range obj.GetDst() {
		if v.AppendTransact(transact) {
			advance := obj.GetPipeline().GetModelTime() - obj.inFacility.timeOfInput
			obj.inFacility.sum_advance += float64(advance)
			obj.tb.Remove(transact)
			obj.inFacility.HoldedTransactID = -1
			return
		}
	}
	transact.SetParameters([]Parameter{{Name: "Facility", Value: obj.name}})
}

func (obj *OutFacility) AppendTransact(transact ITransaction) bool {
	if obj.inFacility.HoldedTransactID != transact.GetId() {
		return false
	}
	obj.GetLogger().GetTrace().Println("Append transact ", transact.GetId(), " to Facility")
	obj.HandleTransact(transact)
	if obj.tb.GetLen() == 0 {
		return true
	}
	return false
}

func (obj *OutFacility) PrintReport() {
	return
}
