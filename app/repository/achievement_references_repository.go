package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"UAS/app/models"
	"github.com/google/uuid"
)

type AchievementReferenceRepository interface {
	// CRUD operations
	CreateReference(ref *models.AchievementReference) error
	GetReferenceByID(id uuid.UUID) (*models.AchievementReference, error)
	GetReferenceByMongoID(mongoID string) (*models.AchievementReference, error)
	UpdateReference(ref *models.AchievementReference) error
	DeleteReference(id uuid.UUID) error
	
	// Status management (sesuai SRS)
	SubmitForVerification(id uuid.UUID) error
	VerifyAchievement(id uuid.UUID, verifiedBy uuid.UUID) error
	RejectAchievement(id uuid.UUID, verifiedBy uuid.UUID, rejectionNote string) error
	
	// Query operations
	GetReferencesByStudentID(studentID uuid.UUID, status string) ([]models.AchievementReference, error)
	GetReferencesByAdvisor(advisorID uuid.UUID, status string) ([]models.AchievementReference, error)
	GetAllReferences(status string, limit, offset int) ([]models.AchievementReference, int, error)
	CheckOwnership(achievementID, studentID uuid.UUID) (bool, error)
}

type achievementReferenceRepo struct {
	DB *sql.DB
}

func NewAchievementReferenceRepository(db *sql.DB) AchievementReferenceRepository {
	return &achievementReferenceRepo{DB: db}
}

func (r *achievementReferenceRepo) CreateReference(ref *models.AchievementReference) error {
	ref.ID = uuid.New()
	ref.CreatedAt = time.Now()
	ref.UpdatedAt = time.Now()
	
	// Default status draft sesuai SRS FR-003
	if ref.Status == "" {
		ref.Status = models.AchievementStatusDraft
	}
	
	query := `
		INSERT INTO achievement_references (
			id, student_id, mongo_achievement_id, status, 
			submitted_at, verified_at, verified_by, rejection_note,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	
	_, err := r.DB.Exec(query,
		ref.ID,
		ref.StudentID,
		ref.MongoAchievementID,
		ref.Status,
		ref.SubmittedAt,
		ref.VerifiedAt,
		ref.VerifiedBy,
		ref.RejectionNote,
		ref.CreatedAt,
		ref.UpdatedAt,
	)
	
	return err
}

func (r *achievementReferenceRepo) GetReferenceByID(id uuid.UUID) (*models.AchievementReference, error) {
	var ref models.AchievementReference
	var submittedAt, verifiedAt sql.NullTime
	var verifiedBy sql.NullString
	var rejectionNote sql.NullString
	
	query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		WHERE id = $1
	`
	
	err := r.DB.QueryRow(query, id).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&submittedAt,
		&verifiedAt,
		&verifiedBy,
		&rejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	// Handle nullable fields
	if submittedAt.Valid {
		ref.SubmittedAt = &submittedAt.Time
	}
	if verifiedAt.Valid {
		ref.VerifiedAt = &verifiedAt.Time
	}
	if verifiedBy.Valid {
		parsedUUID, _ := uuid.Parse(verifiedBy.String)
		ref.VerifiedBy = &parsedUUID
	}
	if rejectionNote.Valid {
		ref.RejectionNote = &rejectionNote.String
	}
	
	return &ref, nil
}

func (r *achievementReferenceRepo) GetReferenceByMongoID(mongoID string) (*models.AchievementReference, error) {
	var ref models.AchievementReference
	var submittedAt, verifiedAt sql.NullTime
	var verifiedBy sql.NullString
	var rejectionNote sql.NullString
	
	query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		WHERE mongo_achievement_id = $1
	`
	
	err := r.DB.QueryRow(query, mongoID).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&submittedAt,
		&verifiedAt,
		&verifiedBy,
		&rejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	if submittedAt.Valid {
		ref.SubmittedAt = &submittedAt.Time
	}
	if verifiedAt.Valid {
		ref.VerifiedAt = &verifiedAt.Time
	}
	if verifiedBy.Valid {
		parsedUUID, _ := uuid.Parse(verifiedBy.String)
		ref.VerifiedBy = &parsedUUID
	}
	if rejectionNote.Valid {
		ref.RejectionNote = &rejectionNote.String
	}
	
	return &ref, nil
}

func (r *achievementReferenceRepo) UpdateReference(ref *models.AchievementReference) error {
	ref.UpdatedAt = time.Now()
	
	query := `
		UPDATE achievement_references 
		SET status = $1, submitted_at = $2, verified_at = $3, 
		    verified_by = $4, rejection_note = $5, updated_at = $6
		WHERE id = $7
	`
	
	_, err := r.DB.Exec(query,
		ref.Status,
		ref.SubmittedAt,
		ref.VerifiedAt,
		ref.VerifiedBy,
		ref.RejectionNote,
		ref.UpdatedAt,
		ref.ID,
	)
	
	return err
}

func (r *achievementReferenceRepo) DeleteReference(id uuid.UUID) error {
	query := `DELETE FROM achievement_references WHERE id = $1`
	
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("achievement reference not found")
	}
	
	return nil
}

// FR-004: Submit untuk Verifikasi
func (r *achievementReferenceRepo) SubmitForVerification(id uuid.UUID) error {
	// Cek status saat ini
	ref, err := r.GetReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}
	
	// Validasi: hanya draft yang bisa disubmit sesuai SRS
	if ref.Status != models.AchievementStatusDraft {
		return fmt.Errorf("cannot submit achievement with status: %s", ref.Status)
	}
	
	now := time.Now()
	query := `
		UPDATE achievement_references 
		SET status = $1, submitted_at = $2, updated_at = $3
		WHERE id = $4
	`
	
	_, err = r.DB.Exec(query, 
		models.AchievementStatusSubmitted,
		now,
		now,
		id,
	)
	
	return err
}

// FR-007: Verify Prestasi
func (r *achievementReferenceRepo) VerifyAchievement(id uuid.UUID, verifiedBy uuid.UUID) error {
	// Cek status saat ini
	ref, err := r.GetReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}
	
	// Validasi: hanya submitted yang bisa diverifikasi sesuai SRS
	if ref.Status != models.AchievementStatusSubmitted {
		return fmt.Errorf("cannot verify achievement with status: %s", ref.Status)
	}
	
	now := time.Now()
	query := `
		UPDATE achievement_references 
		SET status = $1, verified_at = $2, verified_by = $3, updated_at = $4
		WHERE id = $5
	`
	
	_, err = r.DB.Exec(query, 
		models.AchievementStatusVerified,
		now,
		verifiedBy,
		now,
		id,
	)
	
	return err
}

// FR-008: Reject Prestasi
func (r *achievementReferenceRepo) RejectAchievement(id uuid.UUID, verifiedBy uuid.UUID, rejectionNote string) error {
	// Cek status saat ini
	ref, err := r.GetReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}
	
	// Validasi: hanya submitted yang bisa ditolak sesuai SRS
	if ref.Status != models.AchievementStatusSubmitted {
		return fmt.Errorf("cannot reject achievement with status: %s", ref.Status)
	}
	
	now := time.Now()
	query := `
		UPDATE achievement_references 
		SET status = $1, verified_at = $2, verified_by = $3, 
		    rejection_note = $4, updated_at = $5
		WHERE id = $6
	`
	
	_, err = r.DB.Exec(query, 
		models.AchievementStatusRejected,
		now,
		verifiedBy,
		rejectionNote,
		now,
		id,
	)
	
	return err
}

func (r *achievementReferenceRepo) GetReferencesByStudentID(studentID uuid.UUID, status string) ([]models.AchievementReference, error) {
	var whereClause string
	var args []interface{}
	
	args = append(args, studentID)
	whereClause = "WHERE student_id = $1"
	
	if status != "" {
		whereClause += " AND status = $2"
		args = append(args, status)
	}
	
	query := fmt.Sprintf(`
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		%s
		ORDER BY created_at DESC
	`, whereClause)
	
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var references []models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		var submittedAt, verifiedAt sql.NullTime
		var verifiedBy sql.NullString
		var rejectionNote sql.NullString
		
		err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&submittedAt,
			&verifiedAt,
			&verifiedBy,
			&rejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if submittedAt.Valid {
			ref.SubmittedAt = &submittedAt.Time
		}
		if verifiedAt.Valid {
			ref.VerifiedAt = &verifiedAt.Time
		}
		if verifiedBy.Valid {
			parsedUUID, _ := uuid.Parse(verifiedBy.String)
			ref.VerifiedBy = &parsedUUID
		}
		if rejectionNote.Valid {
			ref.RejectionNote = &rejectionNote.String
		}
		
		references = append(references, ref)
	}
	
	return references, nil
}

func (r *achievementReferenceRepo) GetReferencesByAdvisor(advisorID uuid.UUID, status string) ([]models.AchievementReference, error) {
	var whereClause string
	var args []interface{}
	
	args = append(args, advisorID)
	whereClause = `
		WHERE student_id IN (
			SELECT id FROM students WHERE advisor_id = $1
		)
	`
	
	if status != "" {
		whereClause += " AND status = $2"
		args = append(args, status)
	}
	
	query := fmt.Sprintf(`
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		%s
		ORDER BY created_at DESC
	`, whereClause)
	
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var references []models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		var submittedAt, verifiedAt sql.NullTime
		var verifiedBy sql.NullString
		var rejectionNote sql.NullString
		
		err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&submittedAt,
			&verifiedAt,
			&verifiedBy,
			&rejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if submittedAt.Valid {
			ref.SubmittedAt = &submittedAt.Time
		}
		if verifiedAt.Valid {
			ref.VerifiedAt = &verifiedAt.Time
		}
		if verifiedBy.Valid {
			parsedUUID, _ := uuid.Parse(verifiedBy.String)
			ref.VerifiedBy = &parsedUUID
		}
		if rejectionNote.Valid {
			ref.RejectionNote = &rejectionNote.String
		}
		
		references = append(references, ref)
	}
	
	return references, nil
}

func (r *achievementReferenceRepo) GetAllReferences(status string, limit, offset int) ([]models.AchievementReference, int, error) {
	var whereClause string
	var args []interface{}
	
	if status != "" {
		whereClause = "WHERE status = $1"
		args = append(args, status)
	}
	
	// Count total
	countQuery := "SELECT COUNT(*) FROM achievement_references"
	if whereClause != "" {
		countQuery += " " + whereClause
	}
	
	var total int
	err := r.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated data
	query := fmt.Sprintf(`
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)
	
	args = append(args, limit, offset)
	
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var references []models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		var submittedAt, verifiedAt sql.NullTime
		var verifiedBy sql.NullString
		var rejectionNote sql.NullString
		
		err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&submittedAt,
			&verifiedAt,
			&verifiedBy,
			&rejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		
		if err != nil {
			return nil, 0, err
		}
		
		if submittedAt.Valid {
			ref.SubmittedAt = &submittedAt.Time
		}
		if verifiedAt.Valid {
			ref.VerifiedAt = &verifiedAt.Time
		}
		if verifiedBy.Valid {
			parsedUUID, _ := uuid.Parse(verifiedBy.String)
			ref.VerifiedBy = &parsedUUID
		}
		if rejectionNote.Valid {
			ref.RejectionNote = &rejectionNote.String
		}
		
		references = append(references, ref)
	}
	
	return references, total, nil
}

func (r *achievementReferenceRepo) CheckOwnership(achievementID, studentID uuid.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM achievement_references 
		WHERE id = $1 AND student_id = $2
	`
	
	err := r.DB.QueryRow(query, achievementID, studentID).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}