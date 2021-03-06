// Copyright 2019 Sergey Soldatov. All rights reserved.
// This software may be modified and distributed under the terms
// of the Apache license. See the LICENSE file for details.

package gpss

import (
	"fmt"
	"sync"
)

// IFacility implements Facility interface
type IFacility interface {
	IsEmpty() bool
}

// Facility entity with advance in it
type Facility struct {
	BaseObj
	// The mean time increment
	Interval int
	// The time half-range
	Modificator int
	// Holded transast ID
	HoldedTransactID int
	// For backuping Facility/Bifacility name if we includes Facility in Bifacility
	bakupFacilityName string
	// For counting the advance of transact
	sum_advance float64
	// For counting the transacts that go through Bifacility
	cnt_transact float64
}

// Creates new Facility.
// name - name of object; interval - the mean time increment;
// modificator - the time half-range
func NewFacility(name string, interval, modificator int) *Facility {
	obj := &Facility{}
	obj.BaseObj.Init(name)
	obj.Interval = interval
	obj.Modificator = modificator
	obj.HoldedTransactID = -1
	return obj
}

func (obj *Facility) GenerateAdvance() int {
	advance := obj.Interval
	if obj.Modificator > 0 {
		advance += GetRandom(-obj.Modificator, obj.Modificator)
	}
	return advance
}

func (obj *Facility) HandleTransact(transact ITransaction) {
	transact.DecTiсks()
	transact.PrintInfo()
	if transact.IsTheEnd() {
		if obj.bakupFacilityName != "" {
			transact.SetParameters([]Parameter{{Name: "Facility",
				Value: obj.bakupFacilityName}})
		} else {
			transact.SetParameters([]Parameter{{Name: "Facility", Value: nil}})
		}
		for _, v := range obj.GetDst() {
			if v.AppendTransact(transact) {
				obj.tb.Remove(transact)
				obj.HoldedTransactID = -1
				return
			}
		}
		transact.SetParameters([]Parameter{{Name: "Facility", Value: obj.name}})
	}
}
func (obj *Facility) HandleTransacts(wg *sync.WaitGroup) {
	if obj.tb.GetLen() == 0 {
		wg.Done()
		return
	}
	go func() {
		defer wg.Done()
		transacts := obj.tb.GetItems()
		for _, tr := range transacts {
			obj.HandleTransact(tr.transact)
		}
	}()
}

func (obj *Facility) AppendTransact(transact ITransaction) bool {
	if obj.tb.GetLen() != 0 {
		// Facility is busy
		return false
	}
	obj.GetLogger().GetTrace().Println("Append transact ", transact.GetId(), " to Facility")
	transact.SetHolderName(obj.name)
	advance := obj.GenerateAdvance()
	obj.sum_advance += float64(advance)
	transact.SetTiсks(advance)
	if transact.GetParameterByName("Facility") != nil {
		obj.bakupFacilityName = transact.GetParameterByName("Facility").(string)
	}
	transact.SetParameters([]Parameter{{Name: "Facility", Value: obj.name}})
	obj.HoldedTransactID = transact.GetId()
	obj.tb.Push(transact)
	obj.cnt_transact++
	return true
}

func (obj *Facility) PrintReport() {
	obj.BaseObj.PrintReport()
	avr := obj.sum_advance / obj.cnt_transact
	fmt.Printf("Average advance %.2f \tAverage utilization %.2f%%\tNumber entries %.2f \t", avr,
		100*avr*obj.cnt_transact/float64(obj.GetPipeline().GetSimTime()), obj.cnt_transact)
	if obj.HoldedTransactID > 0 {
		fmt.Print("Transact ", obj.HoldedTransactID, " in facility")
		part, _, parent_id := obj.tb.GetItem(obj.HoldedTransactID).transact.GetParts()
		if parent_id > 0 {
			fmt.Print(", parent transact ", parent_id, " part ", part)
		}
	} else {
		fmt.Print("Facility is empty")
	}
	fmt.Printf("\n\n")
}

func (obj *Facility) IsEmpty() bool {
	if obj.tb.GetLen() != 0 {
		// Facility is busy
		return false
	}
	return true
}
