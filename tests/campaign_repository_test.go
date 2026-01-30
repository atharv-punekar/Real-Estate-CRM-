package tests

import (
	"testing"
	"time"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestCampaignRepository_FindByID_Success(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.CampaignRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create campaign
	campaign := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Test Campaign",
		TemplateID:     uuid.New().String(),
		ScheduleType:   "once",
		ScheduledAt:    time.Now().Add(1 * time.Hour),
		Status:         "scheduled",
		CreatedBy:      uuid.New().String(),
		AudienceIDs:    datatypes.JSONSlice[string]{},
	}
	db.Create(&campaign)

	// Find campaign
	found, err := repo.FindByID(campaign.ID, org.ID.String())

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, campaign.ID, found.ID)
	assert.Equal(t, campaign.Name, found.Name)
}

func TestCampaignRepository_FindByID_NotFound(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.CampaignRepository{}

	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Try to find non-existent campaign
	fakeID := uuid.New().String()
	found, err := repo.FindByID(fakeID, org.ID.String())

	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestCampaignRepository_FindScheduledCampaigns(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.CampaignRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create scheduled campaign (in the past, ready to run)
	pastTime := time.Now().Add(-1 * time.Hour)
	campaign1 := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Scheduled Campaign",
		TemplateID:     uuid.New().String(),
		ScheduleType:   "once",
		ScheduledAt:    pastTime,
		Status:         "scheduled",
		CreatedBy:      uuid.New().String(),
		AudienceIDs:    datatypes.JSONSlice[string]{},
	}
	db.Create(&campaign1)

	// Create future scheduled campaign (not ready yet)
	futureTime := time.Now().Add(2 * time.Hour)
	campaign2 := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Future Campaign",
		TemplateID:     uuid.New().String(),
		ScheduleType:   "once",
		ScheduledAt:    futureTime,
		Status:         "scheduled",
		CreatedBy:      uuid.New().String(),
		AudienceIDs:    datatypes.JSONSlice[string]{},
	}
	db.Create(&campaign2)

	// Create completed campaign
	campaign3 := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Completed Campaign",
		TemplateID:     uuid.New().String(),
		ScheduleType:   "once",
		ScheduledAt:    pastTime,
		Status:         "completed",
		CreatedBy:      uuid.New().String(),
		AudienceIDs:    datatypes.JSONSlice[string]{},
	}
	db.Create(&campaign3)

	// Find scheduled campaigns
	currentTime := time.Now()
	campaigns, err := repo.FindScheduledCampaigns(currentTime)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(campaigns), 1)
	// Should only find campaign1 (scheduled and in the past)
}

func TestCampaignRepository_GetRecipientContacts_SingleContact(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.CampaignRepository{}

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

	// Create campaign with single contact
	campaign := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Test Campaign",
		TemplateID:     uuid.New().String(),
		ContactID:      &contact.ID,
		ScheduleType:   "once",
		ScheduledAt:    time.Now().Add(1 * time.Hour),
		Status:         "scheduled",
		CreatedBy:      uuid.New().String(),
		AudienceIDs:    datatypes.JSONSlice[string]{},
	}
	db.Create(&campaign)

	// Get recipient contacts
	contactIDs, err := repo.GetRecipientContacts(&campaign)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(contactIDs))
	assert.Equal(t, contact.ID, contactIDs[0])
}

func TestCampaignRepository_GetRecipientContacts_MultipleAudiences(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.CampaignRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create contacts
	contact1 := models.Contact{
		OrganizationID: org.ID.String(),
		CreatedBy:      uuid.New().String(),
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@test.com",
	}
	db.Create(&contact1)

	contact2 := models.Contact{
		OrganizationID: org.ID.String(),
		CreatedBy:      uuid.New().String(),
		FirstName:      "Jane",
		LastName:       "Smith",
		Email:          "jane@test.com",
	}
	db.Create(&contact2)

	// Create audience
	audience := models.Audience{
		OrganizationID: org.ID.String(),
		Name:           "Test Audience",
		CreatedBy:      uuid.New().String(),
	}
	db.Create(&audience)

	// Link contacts to audience
	db.Create(&models.AudienceContact{
		AudienceID: audience.ID,
		ContactID:  contact1.ID,
	})
	db.Create(&models.AudienceContact{
		AudienceID: audience.ID,
		ContactID:  contact2.ID,
	})

	// Create campaign with audience
	campaign := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Test Campaign",
		TemplateID:     uuid.New().String(),
		ScheduleType:   "once",
		ScheduledAt:    time.Now().Add(1 * time.Hour),
		Status:         "scheduled",
		CreatedBy:      uuid.New().String(),
		AudienceIDs:    datatypes.JSONSlice[string]{audience.ID},
	}
	db.Create(&campaign)

	// Get recipient contacts
	contactIDs, err := repo.GetRecipientContacts(&campaign)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(contactIDs), 2)
}

func TestCampaignRepository_UpdateStatus(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.CampaignRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create campaign
	campaign := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Test Campaign",
		TemplateID:     uuid.New().String(),
		ScheduleType:   "once",
		ScheduledAt:    time.Now().Add(1 * time.Hour),
		Status:         "scheduled",
		CreatedBy:      uuid.New().String(),
		AudienceIDs:    datatypes.JSONSlice[string]{},
	}
	db.Create(&campaign)

	// Update status
	err := repo.UpdateStatus(campaign.ID, "running")

	assert.NoError(t, err)

	// Verify status updated
	var updated models.Campaign
	db.First(&updated, "id = ?", campaign.ID)
	assert.Equal(t, "running", updated.Status)
}

func TestCampaignRepository_UpdateLastRunAt(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	repo := &repository.CampaignRepository{}

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create campaign
	campaign := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Test Campaign",
		TemplateID:     uuid.New().String(),
		ScheduleType:   "recurring",
		ScheduledAt:    time.Now().Add(1 * time.Hour),
		Status:         "scheduled",
		CreatedBy:      uuid.New().String(),
		AudienceIDs:    datatypes.JSONSlice[string]{},
	}
	db.Create(&campaign)

	// Update last run at
	runTime := time.Now()
	err := repo.UpdateLastRunAt(campaign.ID, runTime)

	assert.NoError(t, err)

	// Verify last_run_at updated
	var updated models.Campaign
	db.First(&updated, "id = ?", campaign.ID)
	assert.NotNil(t, updated.LastRunAt)
}
