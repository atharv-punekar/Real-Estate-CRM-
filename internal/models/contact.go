package models

import "time"

type Contact struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string `gorm:"type:uuid"`
	CreatedBy      string `gorm:"type:uuid"`

	FirstName string
	LastName  string
	Email     string
	Phone     string

	BudgetMin         float64
	BudgetMax         float64
	PropertyType      string
	Bedrooms          int
	Bathrooms         int
	SquareFeet        int
	PreferredLocation string

	IsActive  bool `gorm:"default:true"`
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ContactFilter struct {
	OrganizationID string
	PropertyType   []string
	Bedrooms       []int
	Bathrooms      []int
	Locations      []string
	MinBudget      float64
	MaxBudget      float64
}

func (Contact) TableName() string {
	return "contact"
}
