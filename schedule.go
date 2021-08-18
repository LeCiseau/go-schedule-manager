package manager

import (
	"time"
)

type Schedule struct {
	HairdresserId int       `json:"hairdresserId"`
	Discount      float64   `json:"discount"`
	Range         Range     `json:"range"`
	CreatedAt     time.Time `json:"createdAt,omitempty"`
}

type Schedules []Schedule

func (schedule Schedule) toSlot() Slot {
	return Slot{
		Range:         schedule.Range,
		Discount:      schedule.Discount,
		HairdresserId: schedule.HairdresserId,
	}
}

func (schedules Schedules) toSlots() Slots {
	var slots Slots
	for _, schedule := range schedules {
		slots = append(slots, schedule.toSlot())
	}
	return slots
}

func MergeSlotsAndSchedules(slots Slots, schedules Schedules) (Slots, error) {
	if len(schedules) == 0 {
		return slots, nil
	}
	tmpUnavailabilities := slots.ToUnavailabilities()
	tmpSlots := schedules.toSlots()
	schedulesWithoutSlots, errClean := CleanSlots(tmpSlots, tmpUnavailabilities)
	if errClean != nil {
		return Slots{}, errClean
	}

	newSlots := MergeSlots(slots, schedulesWithoutSlots)
	return newSlots, nil
}
