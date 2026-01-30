



CREATE TABLE super_admin (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE organization (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    timezone TEXT DEFAULT 'Asia/Kolkata',
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES super_admin(id),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_org_is_active ON organization(is_active);

CREATE TABLE "user" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organization(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    
    role TEXT NOT NULL CHECK (role IN ('org_admin', 'org_user')),
    is_active BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_user_org ON "user"(organization_id);
CREATE INDEX idx_user_email ON "user"(email);

CREATE TABLE contact (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    created_by UUID REFERENCES "user"(id),

    first_name TEXT,
    last_name TEXT,
    email TEXT,
    phone TEXT,

    -- Fixed preference fields
    budget_min NUMERIC(12,2),
    budget_max NUMERIC(12,2),
    property_type TEXT,
    bedrooms INTEGER,
    bathrooms INTEGER,
    square_feet INTEGER,
    preferred_location TEXT,

    notes TEXT,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_contact_org ON contact(organization_id);
CREATE INDEX idx_contact_email ON contact(email);
CREATE INDEX idx_contact_phone ON contact(phone);


CREATE TABLE audience (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    description TEXT,

    created_by UUID REFERENCES "user"(id),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE audience_contact (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audience_id UUID REFERENCES audience(id) ON DELETE CASCADE,
    contact_id UUID REFERENCES contact(id) ON DELETE CASCADE,
    UNIQUE(audience_id, contact_id)
);

CREATE TABLE email_template (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    subject TEXT NOT NULL,
    preheader TEXT,
    from_name TEXT,
    reply_to TEXT,

    html_body TEXT NOT NULL,
    plain_text_body TEXT,

    created_by UUID REFERENCES "user"(id),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE campaign (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    template_id UUID REFERENCES email_template(id),

    -- recipients (one OR many groups)
    audience_ids UUID[] DEFAULT NULL,            -- array of audiences
    contact_id UUID REFERENCES contact(id),      -- for single contact campaigns

    schedule_type TEXT NOT NULL CHECK (schedule_type IN ('once', 'recurring')),

    scheduled_at TIMESTAMP NOT NULL,

    -- recurring configs
    recurrence TEXT CHECK (recurrence IN ('daily', 'weekly', 'monthly')),
    recurrence_day_of_week INTEGER CHECK (recurrence_day_of_week BETWEEN 0 AND 6),
    recurrence_day_of_month INTEGER CHECK (recurrence_day_of_month BETWEEN 1 AND 31),
    recurrence_time TIME,

    last_run_at TIMESTAMP,

    status TEXT NOT NULL DEFAULT 'draft'
        CHECK(status IN ('draft','scheduled','running','paused','completed')),

    created_by UUID REFERENCES "user"(id),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CHECK (
        (audience_ids IS NOT NULL AND contact_id IS NULL) OR
        (audience_ids IS NULL AND contact_id IS NOT NULL)
    )
);

CREATE INDEX idx_campaign_org ON campaign(organization_id);
CREATE INDEX idx_campaign_status ON campaign(status);
CREATE INDEX idx_campaign_scheduled_at ON campaign(scheduled_at);

CREATE TABLE campaign_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID REFERENCES campaign(id) ON DELETE CASCADE,
    contact_id UUID REFERENCES contact(id),

    recipient_email TEXT NOT NULL,
    subject TEXT NOT NULL,

    status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'sent', 'failed')),
    error_message TEXT,

    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE notification (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES "user"(id) ON DELETE CASCADE,

    notification_type TEXT NOT NULL CHECK (
        notification_type IN ('agent_added', 'agent_removed', 'campaign_sent', 'csv_import_completed', 'csv_import_failed')
    ),

    title TEXT NOT NULL,
    message TEXT NOT NULL,

    related_user_id UUID REFERENCES "user"(id) ON DELETE SET NULL,
    related_campaign_id UUID REFERENCES campaign(id) ON DELETE SET NULL,

    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE background_job_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    job_type TEXT NOT NULL CHECK (job_type IN ('csv_import', 'campaign_run', 'campaign_scheduler')),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,

    reference_id UUID,   -- e.g., campaign_id

    status TEXT NOT NULL DEFAULT 'queued'
        CHECK(status IN ('queued','running','success','failed')),

    total_records INTEGER,
    processed_records INTEGER DEFAULT 0,
    error_message TEXT,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO super_admin (name, email, password_hash)
VALUES (
  'Atharv',
  'atharv@example.com',
  '$2a$14$f5OWdjt96eJd8q39kXjXFuYcIddsPRNl3/KrpWYayJG5E3yxwKqm6'
);

SELECT * FROM super_admin;

ALTER TABLE organization
ADD COLUMN updated_at TIMESTAMP DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_contact_org ON contact(organization_id);
CREATE INDEX IF NOT EXISTS idx_contact_property_type ON contact(property_type);
CREATE INDEX IF NOT EXISTS idx_contact_bedrooms ON contact(bedrooms);
CREATE INDEX IF NOT EXISTS idx_contact_bathrooms ON contact(bathrooms);
CREATE INDEX IF NOT EXISTS idx_contact_location ON contact(preferred_location);
CREATE INDEX IF NOT EXISTS idx_contact_budget_min ON contact(budget_min);
CREATE INDEX IF NOT EXISTS idx_contact_budget_max ON contact(budget_max);


ALTER TABLE campaign
ALTER COLUMN audience_ids TYPE jsonb
USING audience_ids::jsonb;

select * from audience_contact;

SELECT * FROM campaign_log ORDER BY created_at DESC;
SELECT * FROM background_job_log ORDER BY created_at DESC;

SELECT * FROM "user";

