package manager

import (
	"errors"
	"strconv"
	"time"
)

type Booking struct {
	Id                   int64                `json:"id"`
	SalonId              int                  `json:"salonId,omitempty"`
	UserId               int                  `json:"userId,omitempty"`
	DateStart            time.Time            `json:"dateStart,omitempty"`
	DateEnd              time.Time            `json:"dateEnd,omitempty"`
	TransactionId        string               `json:"transactionId,omitempty"`
	CommissionAmount     float64              `json:"commissionAmount,omitempty"`
	PaymentType          string               `json:"paymentType,omitempty"`
	PayoutId             string               `json:"payoutId,omitempty"`
	Discount             float64              `json:"discount,omitempty"`
	Origin               string               `json:"origin,omitempty"`
	CancellationReason   string               `json:"cancellationReason,omitempty"`
	CancellationOrigin   string               `json:"cancellationOrigin,omitempty"`
	Insurance            float64              `json:"insurance,omitempty"`
	Donation             Donation             `json:"donation,omitempty"`
	Voucher              Voucher              `json:"voucher,omitempty"`
	Gift                 Gift                 `json:"gift,omitempty"`
	Refund               float64              `json:"refund,omitempty"`
	Status               string               `json:"status,omitempty"`
	PartnerRef           PartnerRef           `json:"partnerRef,omitempty"`
	SnapshotPackages     []SnapshotPackages   `json:"snapshotPackages,omitempty"`
	CommissionConditions CommissionConditions `json:"commissionConditions,omitempty"`
	CreatedAt            string               `json:"createdAt,omitempty"`
	UpdatedAt            string               `json:"updatedAt,omitempty"`
}

type SnapshotPackages struct {
	Id             int64         `json:"id"`
	Price          SnapshotPrice `json:"price"`
	Name           string        `json:"name"`
	HairdresserId  int           `json:"hairdresserId"`
	Description    string        `json:"description"`
	MainCategoryId int           `json:"mainCategoryId"`
	AllowPromo     bool          `json:"allowPromo"`
	PartnerRef     PartnerRef    `json:"partnerRef"`
}

type CommissionConditions struct {
	IsNewClient *bool `json:"isNewClient"`
}

type Donation struct {
	Amount     float64 `json:"amount"`
	DonationId string  `json:"donation"`
}

type Voucher struct {
	RedemptionId string  `json:"redemptionId"`
	Code         string  `json:"code"`
	Amount       float64 `json:"amount"`
	Type         string  `json:"type"`
}

type Gift struct {
	RedemptionIds []string `json:"redemptionIds"`
	TotalAmount   float64  `json:"totalAmount"`
}

type Variant struct {
	Id   int    `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type SnapshotPrice struct {
	Variant       Variant    `json:"variant"`
	Price         float64    `json:"price"`
	Currency      string     `json:"currency"`
	Duration      string     `json:"duration"`
	PackageNumber string     `json:"packageNumber"`
	PartnerRef    PartnerRef `json:"partnerRef"`
}

type Bookings []Booking

func (booking Booking) ToUnavailability() (Unavailability, error) {
	if len(booking.SnapshotPackages) < 1 {
		return Unavailability{}, errors.New("Empty snapshot packages in booking")
	}
	duration, errorDuration := strconv.Atoi(booking.SnapshotPackages[0].Price.Duration)
	if errorDuration != nil {
		return Unavailability{}, errorDuration
	}
	// @TODO check multiple snapshot
	dateEnd := booking.DateStart.Add(time.Minute * time.Duration(duration))
	return Unavailability{
		Range: Range{
			Start: booking.DateStart,
			End:   dateEnd,
		},
		HairdresserId: booking.SnapshotPackages[0].HairdresserId,
	}, nil
}

func (bookings Bookings) ToUnavailabilities() (Unavailabilities, error) {
	var unavailabilities Unavailabilities
	for _, booking := range bookings {
		unavailability, err := booking.ToUnavailability()
		if err != nil {
			return unavailabilities, err
		}
		unavailabilities = append(unavailabilities, unavailability)
	}
	return unavailabilities, nil
}
