package repository

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"

	"UAS/app/models"
)

type StudentRepository interface {
	GetByUserID(userID string) (*models.Student, error)
	GetByID(id uuid.UUID) (*models.Student, error)  // TAMBAHKAN INI
	Create(student models.Student) (uuid.UUID, error)
	GetAll() ([]models.Student, error)
	GetAllByAdvisorID(advisorID string) ([]models.Student, error)
}

type studentRepo struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepo{DB: db}
}

func (r *studentRepo) GetByUserID(userID string) (*models.Student, error) {
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