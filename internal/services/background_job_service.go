package services

import (
	"bytes"
	"log"
	"time"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
)

type BackgroundJobService struct {
	jobRepo         *repository.BackgroundJobRepository
	contactRepo     *repository.ContactRepository
	campaignRepo    *repository.CampaignRepository
	campaignLogRepo *repository.CampaignLogRepository
	templateRepo    *repository.EmailTemplateRepository
	emailService    *EmailService
	notifService    *NotificationService
	contactService  *ContactService
}

func NewBackgroundJobService() *BackgroundJobService {
	return &BackgroundJobService{
		jobRepo:         &repository.BackgroundJobRepository{},
		contactRepo:     &repository.ContactRepository{},
		campaignRepo:    &repository.CampaignRepository{},
		campaignLogRepo: &repository.CampaignLogRepository{},
		templateRepo:    &repository.EmailTemplateRepository{},
		emailService:    NewEmailService(),
		notifService:    NewNotificationService(),
		contactService:  NewContactService(),
	}
}

// StartJob updates job status from queued to running
func (s *BackgroundJobService) StartJob(jobID string) error {
	return s.jobRepo.UpdateStatus(jobID, "running", "")
}

// FinishJob updates job status from running to success
func (s *BackgroundJobService) FinishJob(jobID string) error {
	return s.jobRepo.UpdateStatus(jobID, "success", "")
}

// FailJob updates job status to failed with error message
func (s *BackgroundJobService) FailJob(jobID string, errorMsg string) error {
	return s.jobRepo.UpdateStatus(jobID, "failed", errorMsg)
}

// UpdateProgress updates the processed records counter
func (s *BackgroundJobService) UpdateProgress(jobID string, count int) error {
	return s.jobRepo.UpdateProgress(jobID, count)
}

// ProcessCSVImport processes a CSV import job
func (s *BackgroundJobService) ProcessCSVImport(jobID, orgID, userID string, csvData []byte) {
	// Start the job (queued → running)
	s.StartJob(jobID)

	// Parse CSV
	reader := bytes.NewReader(csvData)
	contacts, err := s.contactService.ParseCSV(reader, orgID, userID)
	if err != nil {
		s.FailJob(jobID, err.Error())
		s.notifService.NotifyCSVImportFailed(orgID, userID, err.Error())
		return
	}

	// Update total records
	job, _ := s.jobRepo.FindByID(jobID)
	if job != nil {
		totalRecords := len(contacts)
		job.TotalRecords = &totalRecords
		s.jobRepo.Update(job)
	}

	// Bulk create contacts
	successCount, skipCount, err := s.contactService.BulkCreateContacts(contacts)
	if err != nil {
		s.FailJob(jobID, err.Error())
		s.notifService.NotifyCSVImportFailed(orgID, userID, err.Error())
		return
	}

	// Update job progress and finish
	if job != nil {
		job.ProcessedRecords = &successCount
		s.jobRepo.Update(job)
	}
	s.FinishJob(jobID)

	// Notify user
	s.notifService.NotifyCSVImportCompleted(orgID, userID, successCount)
	log.Printf("CSV Import completed: %d imported, %d skipped", successCount, skipCount)
}

// ProcessCampaignRun executes a campaign and sends emails
func (s *BackgroundJobService) ProcessCampaignRun(jobID, campaignID string) {
	// Start the job (queued → running)
	s.StartJob(jobID)

	// Get campaign (scheduler context - no org restriction needed)
	campaign, err := s.campaignRepo.FindByIDOnly(campaignID)
	if err != nil {
		s.FailJob(jobID, "Campaign not found")
		return
	}

	// Safety check 1: Skip if already completed
	if campaign.Status == "completed" {
		s.FailJob(jobID, "Campaign already completed")
		log.Printf("Skipping already completed campaign: %s", campaignID)
		return
	}

	// Safety check 2: Skip if organization_id is invalid
	if campaign.OrganizationID == "" {
		s.FailJob(jobID, "Invalid organization_id")
		log.Printf("Skipping campaign with empty org_id: %s", campaignID)
		return
	}

	// Safety check 3: For one-time campaigns, check if already run
	if campaign.ScheduleType == "once" && campaign.LastRunAt != nil {
		s.FailJob(jobID, "One-time campaign already executed")
		log.Printf("Skipping one-time campaign that already ran: %s", campaignID)
		// Update status to completed if not already
		if campaign.Status != "completed" {
			campaign.Status = "completed"
			s.campaignRepo.Update(campaign)
		}
		return
	}

	// Update campaign status to running
	s.campaignRepo.UpdateStatus(campaignID, "running")

	// Get template
	template, err := s.templateRepo.FindByID(campaign.TemplateID, campaign.OrganizationID)
	if err != nil {
		s.FailJob(jobID, "Template not found")
		s.campaignRepo.UpdateStatus(campaignID, "failed")
		return
	}

	// Get recipient contacts
	contactIDs, err := s.campaignRepo.GetRecipientContacts(campaign)
	if err != nil {
		s.FailJob(jobID, "Failed to get recipients")
		s.campaignRepo.UpdateStatus(campaignID, "failed")
		return
	}

	if len(contactIDs) == 0 {
		s.FailJob(jobID, "No recipients found")
		s.campaignRepo.UpdateStatus(campaignID, "completed")
		return
	}

	// Get contact details
	contacts, err := s.contactRepo.FindByIDs(contactIDs, campaign.OrganizationID)
	if err != nil {
		s.FailJob(jobID, "Failed to fetch contact details")
		s.campaignRepo.UpdateStatus(campaignID, "failed")
		return
	}

	// Update job total records
	job, _ := s.jobRepo.FindByID(jobID)
	if job != nil {
		totalRecords := len(contacts)
		job.TotalRecords = &totalRecords
		s.jobRepo.Update(job)
	}

	// Send emails to each contact
	sentCount := 0
	for _, contact := range contacts {
		// Skip contacts without email
		if contact.Email == "" {
			continue
		}

		// Create campaign log entry
		campaignLog := models.CampaignLog{
			CampaignID:     campaignID,
			ContactID:      contact.ID,
			RecipientEmail: contact.Email,
			Subject:        template.Subject,
			Status:         "queued",
		}
		s.campaignLogRepo.Create(&campaignLog)

		// Substitute template variables
		variables := map[string]string{
			"first_name": contact.FirstName,
			"last_name":  contact.LastName,
			"email":      contact.Email,
			"phone":      contact.Phone,
		}

		htmlBody := SubstituteTemplateVariables(template.HtmlBody, variables)
		plainTextBody := SubstituteTemplateVariables(template.PlainTextBody, variables)
		subject := SubstituteTemplateVariables(template.Subject, variables)

		// Send email (no fromName/replyTo - using SMTP_FROM from config)
		err := s.emailService.SendCampaignEmail(
			contact.Email,
			subject,
			htmlBody,
			plainTextBody,
		)

		if err != nil {
			// Update log as failed
			s.campaignLogRepo.UpdateStatus(campaignLog.ID, "failed", err.Error())
		} else {
			// Update log as sent
			s.campaignLogRepo.UpdateStatus(campaignLog.ID, "sent", "")
			sentCount++
		}
	}

	// Update job progress and finish
	if job != nil {
		job.ProcessedRecords = &sentCount
		s.jobRepo.Update(job)
	}
	s.FinishJob(jobID)

	// Update campaign
	now := time.Now()
	campaign.LastRunAt = &now

	// For one-time campaigns, mark as completed
	// Update status after success
	if campaign.ScheduleType == "once" {
		campaign.Status = "completed"
	} else {
		campaign.Status = "scheduled"
	}

	s.campaignRepo.Update(campaign)

	// Notify user
	s.notifService.NotifyCampaignSent(campaign.OrganizationID, campaign.CreatedBy, campaignID, sentCount)
	log.Printf("Campaign %s completed: %d emails sent", campaignID, sentCount)
}

// ProcessCampaignScheduler checks for due campaigns and queues them
func (s *BackgroundJobService) ProcessCampaignScheduler() {
	log.Println("Running campaign scheduler...")

	currentTime := time.Now()

	// Find scheduled campaigns that are due
	campaigns, err := s.campaignRepo.FindScheduledCampaigns(currentTime)
	if err != nil {
		log.Printf("Error finding scheduled campaigns: %v", err)
		return
	}

	for _, campaign := range campaigns {
		// Safety check: Skip one-time campaigns that have already run
		if campaign.ScheduleType == "once" && campaign.LastRunAt != nil {
			log.Printf("Skipping already-completed one-time campaign: %s", campaign.ID)
			continue
		}

		// Create background job for campaign execution
		job := models.BackgroundJobLog{
			JobType:        "campaign_run",
			OrganizationID: campaign.OrganizationID,
			ReferenceID:    &campaign.ID,
			Status:         "queued",
		}
		if err := s.jobRepo.Create(&job); err != nil {
			log.Printf("Error creating job for campaign %s: %v", campaign.ID, err)
			continue
		}

		// Execute campaign in goroutine
		go s.ProcessCampaignRun(job.ID, campaign.ID)
	}

	// Handle recurring campaigns
	recurringCampaigns, err := s.campaignRepo.FindRecurringCampaigns()
	if err != nil {
		log.Printf("Error finding recurring campaigns: %v", err)
		return
	}

	for _, campaign := range recurringCampaigns {
		// Safety check 1: Skip if organization_id is invalid
		if campaign.OrganizationID == "" {
			log.Printf("Skipping recurring campaign with empty org_id: %s", campaign.ID)
			continue
		}

		// Safety check 2: Skip if status is completed (shouldn't happen but extra safety)
		if campaign.Status == "completed" {
			log.Printf("Skipping completed recurring campaign: %s", campaign.ID)
			continue
		}

		// Check if it's time to run based on recurrence settings
		if s.shouldRunRecurringCampaign(&campaign, currentTime) {
			// Create background job
			job := models.BackgroundJobLog{
				JobType:        "campaign_run",
				OrganizationID: campaign.OrganizationID,
				ReferenceID:    &campaign.ID,
				Status:         "queued",
			}
			if err := s.jobRepo.Create(&job); err != nil {
				log.Printf("Error creating job for recurring campaign %s: %v", campaign.ID, err)
				continue
			}

			// Execute campaign in goroutine
			go s.ProcessCampaignRun(job.ID, campaign.ID)
		}
	}

	log.Println("Campaign scheduler completed")
}

// shouldRunRecurringCampaign determines if a recurring campaign should run now
func (s *BackgroundJobService) shouldRunRecurringCampaign(campaign *models.Campaign, currentTime time.Time) bool {
	// If never run, check if it's past the scheduled time
	if campaign.LastRunAt == nil {
		return currentTime.After(campaign.ScheduledAt)
	}

	lastRun := *campaign.LastRunAt

	switch *campaign.Recurrence {
	case "daily":
		// Run if last run was more than 24 hours ago
		return currentTime.Sub(lastRun) >= 24*time.Hour
	case "weekly":
		// Run if it's the right day of week and at least 7 days since last run
		if campaign.RecurrenceDayOfWeek != nil {
			dayOfWeek := int(currentTime.Weekday())
			return dayOfWeek == *campaign.RecurrenceDayOfWeek && currentTime.Sub(lastRun) >= 7*24*time.Hour
		}
	case "monthly":
		// Run if it's the right day of month and at least 28 days since last run
		if campaign.RecurrenceDayOfMonth != nil {
			dayOfMonth := currentTime.Day()
			return dayOfMonth == *campaign.RecurrenceDayOfMonth && currentTime.Sub(lastRun) >= 28*24*time.Hour
		}
	}

	return false
}
