package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dsn string) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("🔥 Failed to connect DB:", err)
	}

	log.Println("✅ Database connected")
	DB = db
}

// EnsureSchema makes small, backward-compatible schema fixes for existing databases.
// This is intentionally safe to run multiple times.
func EnsureSchema() {
	if DB == nil {
		log.Fatal("🔥 DB is nil (call Connect first)")
	}

	// Columns added for invite-based onboarding / password activation.
	// Some existing DBs may not have these columns even though the Go model does.
	stmts := []string{
		// Timestamps (some older schemas might not have them)
		`ALTER TABLE "user" ADD COLUMN IF NOT EXISTS created_at timestamptz;`,
		`ALTER TABLE "user" ADD COLUMN IF NOT EXISTS updated_at timestamptz;`,

		`ALTER TABLE "user" ADD COLUMN IF NOT EXISTS invite_token varchar(255);`,
		`ALTER TABLE "user" ADD COLUMN IF NOT EXISTS token_expires_at timestamptz;`,
		`ALTER TABLE "user" ADD COLUMN IF NOT EXISTS is_password_set boolean NOT NULL DEFAULT false;`,
		`CREATE INDEX IF NOT EXISTS idx_user_invite_token ON "user"(invite_token);`,
	}

	for _, s := range stmts {
		if err := DB.Exec(s).Error; err != nil {
			log.Fatal("🔥 Schema ensure failed:", err)
		}
	}

	log.Println("✅ Schema ensured")
}
