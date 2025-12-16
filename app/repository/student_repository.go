package repository

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"

	"UAS/app/models"
)

type StudentRepository interface {
	GetByUserID(userID uuid.UUID) (*models.Student, error)
	GetByID(id uuid.UUID) (*models.Student, error) 
	Create(student models.Student) (uuid.UUID, error)
	GetAll() ([]models.Student, error)
	GetAllByAdvisorID(advisorID string) ([]models.Student, error)
	UpdateAdvisor(studentID uuid.UUID, advisorID *uuid.UUID) error
	RemoveAdvisor(studentID uuid.UUID) error
}

type studentRepo struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepo{DB: db}
}

func (r *studentRepo) GetByUserID(userID uuid.UUID) (*models.Student, error) {
	var s models.Student
	err := r.DB.QueryRow(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students WHERE user_id=$1
	`, userID).Scan(&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// TAMBAHKAN METHOD INI
func (r *studentRepo) GetByID(id uuid.UUID) (*models.Student, error) {
	var s models.Student
	err := r.DB.QueryRow(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students WHERE id=$1
	`, id).Scan(&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepo) Create(student models.Student) (uuid.UUID, error) {
	if student.ID == uuid.Nil {
		student.ID = uuid.New()
	}
	
	_, err := r.DB.Exec(`
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, 
		student.ID,
		student.UserID,
		student.StudentID,
		student.ProgramStudy,
		student.AcademicYear,
		student.AdvisorID,
	)
	
	if err != nil {
		return uuid.Nil, err
	}
	
	return student.ID, nil
}

func (r *studentRepo) GetAll() ([]models.Student, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	return students, nil
}

func (r *studentRepo) GetAllByAdvisorID(advisorID string) ([]models.Student, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students WHERE advisor_id=$1 ORDER BY created_at DESC
	`, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	return students, nil
}

func (r *studentRepo) UpdateAdvisor(studentID uuid.UUID, advisorID *uuid.UUID) error {
	var result sql.Result
	var err error
	
	if advisorID == nil {
		result, err = r.DB.Exec(`
			UPDATE students 
			SET advisor_id = NULL 
			WHERE id = $1
		`, studentID)
	} else {
		// Set advisor
		result, err = r.DB.Exec(`
			UPDATE students 
			SET advisor_id = $1 
			WHERE id = $2
		`, advisorID, studentID)
	}
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("student not found")
	}
	
	return nil
}

func (r *studentRepo) RemoveAdvisor(studentID uuid.UUID) error {
	return r.UpdateAdvisor(studentID, nil)
}