package go-schedule-manager

import (
	"time"
)

type Slot struct {
	Id            int        `json:"id"`
	Range         Range      `json:"range"`
	Status        string     `json:"status"`
	PartnerRef    PartnerRef `json:"partnerRef"`
	Discount      float64    `json:"discount"`
	RecurrenceId  int        `json:"recurrenceId"`
	HairdresserId int        `json:"hairdresserId,omitempty"`
	CategoryIds   []int      `json:"categoryIds,omitempty"`
}

type Slots []Slot

type Unavailability struct {
	Id            int        `json:"id"`
	Range         Range      `json:"range"`
	PartnerRef    PartnerRef `json:"partnerRef"`
	RecurrenceId  int        `json:"recurrenceId"`
	HairdresserId int        `json:"hairdresserId,omitempty"`
}

type Unavailabilities []Unavailability

type Range struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Availabilities struct {
	Slot Slot
	Next *Availabilities
}

func SlotAndUnavailabiltyOverlap(slot Slot, unavailability Unavailability) string {
	dateStart := unavailability.Range.Start
	dateEnd := unavailability.Range.End

	slotStart := slot.Range.Start
	slotEnd := slot.Range.End

	slotStartingAfter := slotStart.After(dateStart)
	slotStartingBefore := slotStart.Before(dateStart)
	isStartingSame := slotStart.Equal(dateStart)
	unavailabilityStartingBeforeEnd := dateStart.Before(slotEnd)
	unavailabilityEndAfterStart := dateEnd.After(slotStart)
	slotEndingAfter := slotEnd.After(dateEnd)
	slotEndingBefore := slotEnd.Before(dateEnd)
	isEndingSame := slotEnd.Equal(dateEnd)

	// fmt.Printf("bookingStart: %v // slotStart:%v // slotEnd:%s // slotStartingBefore:%v // slotEndingAfter:%v // slotStartingAfter:%v // isStartingSame:%v // slotEndingBefore:%v // unavailabilityEndAfterStart:%v // unavailabilityEndAfterStart:%v \n\n",
	//	booking.DateStart, slot.Range.Start, slot.Range.End, slotStartingBefore, slotEndingAfter, slotStartingAfter, isStartingSame, slotEndingBefore, unavailabilityEndAfterStart, unavailabilityStartingBeforeEnd)

	contain := slotStartingBefore && slotEndingAfter
	overlap := (slotStartingAfter || isStartingSame) && (slotEndingBefore || isEndingSame)
	overlapStart := (slotStartingAfter || isStartingSame) && (isEndingSame || unavailabilityEndAfterStart)
	overlapEnd := slotStartingBefore && (isEndingSame || unavailabilityStartingBeforeEnd)

	hashFct := map[string]bool{
		"contain":      contain,
		"overlap":      overlap,
		"overlapStart": overlapStart && !contain && !overlap,
		"overlapEnd":   overlapEnd && !contain && !overlap,
	}

	for idx := range hashFct {
		if hashFct[idx] {
			return idx
		}
	}
	return ""
}

var UPDATE_SLOT = map[string]func(head **Availabilities, previous *Availabilities, current *Availabilities, unavailability Unavailability){
	"contain":      SplitSlotWithUnavailability,
	"overlap":      RemoveAvailability,
	"overlapStart": UpdateAvailabilityStart,
	"overlapEnd":   UpdateAvailabilityEnd,
}

func UpdateAvailabilityStart(availabilities **Availabilities, previous *Availabilities, availability *Availabilities, unavailability Unavailability) {
	availability.Slot.Range.Start = unavailability.Range.End
}

func UpdateAvailabilityEnd(availabilities **Availabilities, previous *Availabilities, availability *Availabilities, unavailability Unavailability) {
	availability.Slot.Range.End = unavailability.Range.Start
}

func RemoveAvailability(availabilities **Availabilities, previous *Availabilities, availability *Availabilities, unavailability Unavailability) {
	if previous == nil || *availabilities == availability {
		*availabilities = availability.Next
		return
	}
	previous.Next = availability.Next
}

func SplitSlotWithUnavailability(availabilities **Availabilities, previous *Availabilities, availability *Availabilities, unavailability Unavailability) {
	tmpNext := availability.Next
	tmpEnd := availability.Slot.Range.End

	availability.Slot.Range.End = unavailability.Range.Start
	availability.Next = new(Availabilities)
	availability.Next.Next = tmpNext
	availability.Next.Slot = availability.Slot
	availability.Next.Slot.Range.End = tmpEnd
	availability.Next.Slot.Range.Start = unavailability.Range.End
}

func CleanSlots(rawSlots Slots, unavailabilities Unavailabilities) ([]Slot, error) {
	availabilities := new(Availabilities)
	headAvailabilities := availabilities
	var slots []Slot
	availabilities.Slot = rawSlots[0]
	var previous *Availabilities

	for i := range rawSlots {
		if len(rawSlots)-1 > i {
			availability := Availabilities{
				Slot: rawSlots[i+1],
			}
			availabilities.Next = &availability
			availabilities = availabilities.Next
		}
	}
	cursor := headAvailabilities
	for ; cursor != nil; cursor = cursor.Next {
		wasOverlaped := false
		for _, unavailability := range unavailabilities {
			overlap := SlotAndUnavailabiltyOverlap(cursor.Slot, unavailability)
			if overlap != "" && UPDATE_SLOT[overlap] != nil {
				UPDATE_SLOT[overlap](&headAvailabilities, previous, cursor, unavailability)
				// Booking > slot
				if overlap == "overlap" {
					wasOverlaped = true
					break
				}
			}
		}
		if !wasOverlaped {
			previous = cursor
		}
	}

	cursor2 := headAvailabilities
	for ; cursor2 != nil; cursor2 = cursor2.Next {
		slots = append(slots, cursor2.Slot)
	}
	return slots, nil
}

func (slot Slot) toUnavailability() Unavailability {
	return Unavailability{
		Range:         slot.Range,
		HairdresserId: slot.HairdresserId,
	}
}

func (slots Slots) toUnavailabilities() Unavailabilities {
	var unavailabilities Unavailabilities
	for _, slot := range slots {
		unavailabilities = append(unavailabilities, slot.toUnavailability())
	}
	return unavailabilities
}

func (slots *Slots) RemoveSlotAtIndex(index int) {
	slotsValue := *slots
	*slots = append(slotsValue[:index], slotsValue[index+1:]...)
}

func (slot Slot) Touch(otherSlot Slot) bool {
	return slot.Range.End.Equal(otherSlot.Range.Start) ||
		slot.Range.Start.Equal(otherSlot.Range.End)
}

func (slot *Slot) MergeWithSlot(touchingSlot Slot) {
	if slot.Range.Start.After(touchingSlot.Range.Start) {
		slot.Range.Start = touchingSlot.Range.Start
	}
	if slot.Range.End.Before(touchingSlot.Range.End) {
		slot.Range.End = touchingSlot.Range.End
	}
}

func (slot *Slot) foundAndMergeTouchingSlots(slots *Slots, slotIndex int) {
	slotList := *slots
	for i := 0; i < len(slotList); i++ {
		if i != slotIndex && slotList[i].Discount == slot.Discount &&
			slotList[i].Touch(*slot) {
			slot.MergeWithSlot(slotList[i])
			slots.RemoveSlotAtIndex(i)
			i--
		}
	}
}

// MergeSlots only merge touching slots, ⚠️slots should not overlap ⚠️
func MergeSlots(slotsA Slots, slotsB Slots) Slots {
	allSlots := append(slotsA, slotsB...)
	for i := 0; i < len(allSlots); i++ {
		allSlots[i].foundAndMergeTouchingSlots(&allSlots, i)
	}
	return allSlots
}
