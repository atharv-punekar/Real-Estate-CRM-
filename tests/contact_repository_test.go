package tests

import (
	"testing"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestContactRepository_FindAllByOrg(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.ContactRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create contacts
	for i := 1; i <= 3; i++ {
		contact := models.Contact{
			OrganizationID: org.ID.String(),
			CreatedBy:      uuid.New().String(),
			FirstName:      "Contact",
			LastName:       string(rune(i)),
			Email:          "contact" + string(rune(i)) + "@test.com",
		}
		db.Create(&contact)
	}

	// Find all contacts
	contacts, total, err := repo.FindAllByOrg(org.ID.String(), 1, 10, "")

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(contacts), 3)
	assert.GreaterOrEqual(t, int(total), 3)
}

func TestContactRepository_FindByID(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.ContactRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create contact
	contact := models.Contact{
		OrganizationID: org.ID.String(),
		CreatedBy:      uuid.New().String(),
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@test.com",
	}
	db.Create(&contact)

	// Find contact
	found, err := repo.FindByID(contact.ID, org.ID.String())

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, contact.ID, found.ID)
	assert.Equal(t, contact.FirstName, found.FirstName)
}

func TestContactRepository_FindWithFilter(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.ContactRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create contacts with different attributes
	contact1 := models.Contact{
		OrganizationID:    org.ID.String(),
		CreatedBy:         uuid.New().String(),
		FirstName:         "John",
		LastName:          "Doe",
		Email:             "john@test.com",
		PropertyType:      "apartment",
		Bedrooms:          3,
		Bathrooms:         2,
		PreferredLocation: "Downtown",
		BudgetMin:         300000,
		BudgetMax:         500000,
	}
	db.Create(&contact1)

	contact2 := models.Contact{
		OrganizationID:    org.ID.String(),
		CreatedBy:         uuid.New().String(),
		FirstName:         "Jane",
		LastName:          "Smith",
		Email:             "jane@test.com",
		PropertyType:      "house",
		Bedrooms:          4,
		Bathrooms:         3,
		PreferredLocation: "Suburbs",
		BudgetMin:         500000,
		BudgetMax:         800000,
	}
	db.Create(&contact2)

	// Filter for apartments
	filter := models.ContactFilter{
		OrganizationID: org.ID.String(),
		PropertyType:   []string{"apartment"},
	}

	contacts, err := repo.FindWithFilter(filter)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(contacts), 1)
	// Should contain contact1
	found := false
	for _, c := range contacts {
		if c.ID == contact1.ID {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestContactRepository_Update(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.ContactRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create contact
	contact := models.Contact{
		OrganizationID: org.ID.String(),
		CreatedBy:      uuid.New().String(),
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@test.com",
	}
	db.Create(&contact)

	// Update contact
	contact.FirstName = "Jane"
	contact.LastName = "Smith"
	err := repo.Update(&contact)

	assert.NoError(t, err)

	// Verify update
	var updated models.Contact
	db.First(&updated, "id = ?", contact.ID)
	assert.Equal(t, "Jane", updated.FirstName)
	assert.Equal(t, "Smith", updated.LastName)
}

func TestContactRepository_Delete(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.ContactRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create contact
	contact := models.Contact{
		OrganizationID: org.ID.String(),
		CreatedBy:      uuid.New().String(),
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@test.com",
	}
	db.Create(&contact)

	// Delete contact
	err := repo.Delete(contact.ID, org.ID.String())

	assert.NoError(t, err)

	// Verify contact is soft deleted (is_active = false)
	var deleted models.Contact
	db.Unscoped().First(&deleted, "id = ?", contact.ID)
	assert.False(t, deleted.IsActive)
}
