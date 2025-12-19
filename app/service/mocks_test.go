package service

import (

	"context"
	"UAS/app/models"
	"github.com/google/uuid"
	
)

type MockAchievementReferenceRepository struct {
	CreateReferenceFn           func(ref *models.AchievementReference) error
	GetReferenceByIDFn          func(id uuid.UUID) (*models.AchievementReference, error)
	GetReferenceByMongoIDFn     func(mongoID string) (*models.AchievementReference, error)
	UpdateReferenceFn           func(ref *models.AchievementReference) error
	DeleteReferenceFn           func(id uuid.UUID) error
	SoftDeleteFn                func(id uuid.UUID) error
	SubmitForVerificationFn     func(id uuid.UUID) error
	VerifyAchievementFn         func(id uuid.UUID, verifiedBy uuid.UUID) error
	RejectAchievementFn         func(id uuid.UUID, verifiedBy uuid.UUID, rejectionNote string) error
	GetReferencesByStudentIDFn  func(studentID uuid.UUID, status string) ([]models.AchievementReference, error)
	GetReferencesByAdvisorFn    func(advisorID uuid.UUID, status string) ([]models.AchievementReference, error)
	GetAllReferencesFn          func(status string, limit, offset int) ([]models.AchievementReference, int, error)
	CheckOwnershipFn            func(achievementID, studentID uuid.UUID) (bool, error)
}

func (m *MockAchievementReferenceRepository) CreateReference(ref *models.AchievementReference) error {
	if m.CreateReferenceFn != nil {
		return m.CreateReferenceFn(ref)
	}
	return nil
}
func (m *MockAchievementReferenceRepository) GetReferenceByID(id uuid.UUID) (*models.AchievementReference, error) {
	if m.GetReferenceByIDFn != nil {
		return m.GetReferenceByIDFn(id)
	}
	return nil, nil
}
func (m *MockAchievementReferenceRepository) GetReferenceByMongoID(mongoID string) (*models.AchievementReference, error) {
	if m.GetReferenceByMongoIDFn != nil {
		return m.GetReferenceByMongoIDFn(mongoID)
	}
	return nil, nil
}
func (m *MockAchievementReferenceRepository) UpdateReference(ref *models.AchievementReference) error {
	if m.UpdateReferenceFn != nil {
		return m.UpdateReferenceFn(ref)
	}
	return nil
}
func (m *MockAchievementReferenceRepository) DeleteReference(id uuid.UUID) error {
	if m.DeleteReferenceFn != nil {
		return m.DeleteReferenceFn(id)
	}
	return nil
}
func (m *MockAchievementReferenceRepository) SoftDelete(id uuid.UUID) error {
	if m.SoftDeleteFn != nil {
		return m.SoftDeleteFn(id)
	}
	return nil
}
func (m *MockAchievementReferenceRepository) SubmitForVerification(id uuid.UUID) error {
	if m.SubmitForVerificationFn != nil {
		return m.SubmitForVerificationFn(id)
	}
	return nil
}
func (m *MockAchievementReferenceRepository) VerifyAchievement(id uuid.UUID, verifiedBy uuid.UUID) error {
	if m.VerifyAchievementFn != nil {
		return m.VerifyAchievementFn(id, verifiedBy)
	}
	return nil
}
func (m *MockAchievementReferenceRepository) RejectAchievement(id uuid.UUID, verifiedBy uuid.UUID, rejectionNote string) error {
	if m.RejectAchievementFn != nil {
		return m.RejectAchievementFn(id, verifiedBy, rejectionNote)
	}
	return nil
}
func (m *MockAchievementReferenceRepository) GetReferencesByStudentID(studentID uuid.UUID, status string) ([]models.AchievementReference, error) {
	if m.GetReferencesByStudentIDFn != nil {
		return m.GetReferencesByStudentIDFn(studentID, status)
	}
	return nil, nil
}
func (m *MockAchievementReferenceRepository) GetReferencesByAdvisor(advisorID uuid.UUID, status string) ([]models.AchievementReference, error) {
	if m.GetReferencesByAdvisorFn != nil {
		return m.GetReferencesByAdvisorFn(advisorID, status)
	}
	return nil, nil
}
func (m *MockAchievementReferenceRepository) GetAllReferences(status string, limit, offset int) ([]models.AchievementReference, int, error) {
	if m.GetAllReferencesFn != nil {
		return m.GetAllReferencesFn(status, limit, offset)
	}
	return nil, 0, nil
}
func (m *MockAchievementReferenceRepository) CheckOwnership(achievementID, studentID uuid.UUID) (bool, error) {
	if m.CheckOwnershipFn != nil {
		return m.CheckOwnershipFn(achievementID, studentID)
	}
	return false, nil
}

// ===== MOCK AchievementRepository =====
type MockAchievementRepository struct {
	CreateAchievementFn     func(ctx context.Context, achievement *models.Achievement) (string, error)
	GetAchievementByIDFn    func(ctx context.Context, id string) (*models.Achievement, error)
	UpdateAchievementFn     func(ctx context.Context, id string, achievement *models.Achievement) error
	DeleteAchievementFn     func(ctx context.Context, id string) error
	FindAchievementsFn      func(ctx context.Context, studentIDs []string, achievementType, search string, page, limit int, sortBy, sortOrder string) ([]models.Achievement, int64, error)
	GetAchievementsByIDsFn  func(ctx context.Context, ids []string) ([]models.Achievement, error)
	AddAttachmentFn         func(ctx context.Context, achievementID string, attachment models.Attachment) error
	RemoveAttachmentFn      func(ctx context.Context, achievementID, fileName string) error
}

func (m *MockAchievementRepository) CreateAchievement(ctx context.Context, achievement *models.Achievement) (string, error) {
	if m.CreateAchievementFn != nil {
		return m.CreateAchievementFn(ctx, achievement)
	}
	return "", nil
}
func (m *MockAchievementRepository) GetAchievementByID(ctx context.Context, id string) (*models.Achievement, error) {
	if m.GetAchievementByIDFn != nil {
		return m.GetAchievementByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *MockAchievementRepository) UpdateAchievement(ctx context.Context, id string, achievement *models.Achievement) error {
	if m.UpdateAchievementFn != nil {
		return m.UpdateAchievementFn(ctx, id, achievement)
	}
	return nil
}
func (m *MockAchievementRepository) DeleteAchievement(ctx context.Context, id string) error {
	if m.DeleteAchievementFn != nil {
		return m.DeleteAchievementFn(ctx, id)
	}
	return nil
}
func (m *MockAchievementRepository) FindAchievements(ctx context.Context, studentIDs []string, achievementType, search string, page, limit int, sortBy, sortOrder string) ([]models.Achievement, int64, error) {
	if m.FindAchievementsFn != nil {
		return m.FindAchievementsFn(ctx, studentIDs, achievementType, search, page, limit, sortBy, sortOrder)
	}
	return nil, 0, nil
}
func (m *MockAchievementRepository) GetAchievementsByIDs(ctx context.Context, ids []string) ([]models.Achievement, error) {
	if m.GetAchievementsByIDsFn != nil {
		return m.GetAchievementsByIDsFn(ctx, ids)
	}
	return nil, nil
}
func (m *MockAchievementRepository) AddAttachment(ctx context.Context, achievementID string, attachment models.Attachment) error {
	if m.AddAttachmentFn != nil {
		return m.AddAttachmentFn(ctx, achievementID, attachment)
	}
	return nil
}
func (m *MockAchievementRepository) RemoveAttachment(ctx context.Context, achievementID, fileName string) error {
	if m.RemoveAttachmentFn != nil {
		return m.RemoveAttachmentFn(ctx, achievementID, fileName)
	}
	return nil
}

// --- MOCK ROLE REPOSITORY ---
type MockRoleRepository struct {
	GetByIDFn                    func(id uuid.UUID) (*models.Role, error)
	GetByNameFn                  func(name string) (*models.Role, error)
	GetAllFn                     func(page, limit int) ([]models.Role, int, error)
	GetTotalCountFn              func() (int, error)
	GetPermissionsByRoleIDFn     func(roleID uuid.UUID) ([]models.Permission, error)
	GetPermissionNamesByRoleIDFn func(roleID uuid.UUID) ([]string, error)
	AssignPermissionFn           func(roleID, permissionID uuid.UUID) error
	RemovePermissionFn           func(roleID, permissionID uuid.UUID) error
}

func (m *MockRoleRepository) GetByID(id uuid.UUID) (*models.Role, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(id)
	}
	return nil, nil
}

func (m *MockRoleRepository) GetByName(name string) (*models.Role, error) {
	if m.GetByNameFn != nil {
		return m.GetByNameFn(name)
	}
	return nil, nil
}

func (m *MockRoleRepository) GetAll(page, limit int) ([]models.Role, int, error) {
	if m.GetAllFn != nil {
		return m.GetAllFn(page, limit)
	}
	return nil, 0, nil
}

func (m *MockRoleRepository) GetTotalCount() (int, error) {
	if m.GetTotalCountFn != nil {
		return m.GetTotalCountFn()
	}
	return 0, nil
}

func (m *MockRoleRepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]models.Permission, error) {
	if m.GetPermissionsByRoleIDFn != nil {
		return m.GetPermissionsByRoleIDFn(roleID)
	}
	return nil, nil
}

func (m *MockRoleRepository) GetPermissionNamesByRoleID(roleID uuid.UUID) ([]string, error) {
	if m.GetPermissionNamesByRoleIDFn != nil {
		return m.GetPermissionNamesByRoleIDFn(roleID)
	}
	return nil, nil
}

func (m *MockRoleRepository) AssignPermission(roleID, permissionID uuid.UUID) error {
	if m.AssignPermissionFn != nil {
		return m.AssignPermissionFn(roleID, permissionID)
	}
	return nil
}

func (m *MockRoleRepository) RemovePermission(roleID, permissionID uuid.UUID) error {
	if m.RemovePermissionFn != nil {
		return m.RemovePermissionFn(roleID, permissionID)
	}
	return nil
}

// --- MOCK USER REPOSITORY ---
type MockUserRepository struct {
	GetByIDFn              func(id uuid.UUID) (*models.User, error)
	GetByEmailFn           func(email string) (*models.User, error)
	GetByUsernameFn        func(username string) (*models.User, error)
	GetByUsernameOrEmailFn func(identifier string) (*models.User, error)
	CreateFn               func(user *models.User) (uuid.UUID, error)
	UpdateFn               func(id uuid.UUID, req *models.UpdateUserRequest) error
	UpdatePasswordFn       func(id uuid.UUID, hashedPassword string) error
	SoftDeleteFn           func(id uuid.UUID) error
	HardDeleteFn           func(id uuid.UUID) error
	GetAllFn               func(page, limit int) ([]models.User, int, error)
	GetInactiveUsersFn     func(page, limit int) ([]models.User, int, error)
	GetAllWithInactiveFn   func(page, limit int) ([]models.User, int, error)
	SearchByNameFn         func(name string, page, limit int) ([]models.User, int, error)
	GetByRoleFn            func(roleID uuid.UUID, page, limit int) ([]models.User, int, error)
	GetUsersCountByRoleFn  func() (map[uuid.UUID]int, error)
	GetTotalActiveCountFn  func() (int, error)
	GetTotalInactiveCountFn func() (int, error)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	return m.GetByIDFn(id)
}
func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	return m.GetByEmailFn(email)
}
func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	return m.GetByUsernameFn(username)
}
func (m *MockUserRepository) GetByUsernameOrEmail(identifier string) (*models.User, error) {
	return m.GetByUsernameOrEmailFn(identifier)
}
func (m *MockUserRepository) Create(user *models.User) (uuid.UUID, error) {
	return m.CreateFn(user)
}
func (m *MockUserRepository) Update(id uuid.UUID, req *models.UpdateUserRequest) error {
	return m.UpdateFn(id, req)
}
func (m *MockUserRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	return m.UpdatePasswordFn(id, hashedPassword)
}
func (m *MockUserRepository) SoftDelete(id uuid.UUID) error {
	return m.SoftDeleteFn(id)
}
func (m *MockUserRepository) HardDelete(id uuid.UUID) error {
	return m.HardDeleteFn(id)
}
func (m *MockUserRepository) GetAll(page, limit int) ([]models.User, int, error) {
	return m.GetAllFn(page, limit)
}
func (m *MockUserRepository) GetInactiveUsers(page, limit int) ([]models.User, int, error) {
	return m.GetInactiveUsersFn(page, limit)
}
func (m *MockUserRepository) GetAllWithInactive(page, limit int) ([]models.User, int, error) {
	return m.GetAllWithInactiveFn(page, limit)
}
func (m *MockUserRepository) SearchByName(name string, page, limit int) ([]models.User, int, error) {
	return m.SearchByNameFn(name, page, limit)
}
func (m *MockUserRepository) GetByRole(roleID uuid.UUID, page, limit int) ([]models.User, int, error) {
	return m.GetByRoleFn(roleID, page, limit)
}
func (m *MockUserRepository) GetUsersCountByRole() (map[uuid.UUID]int, error) {
	return m.GetUsersCountByRoleFn()
}
func (m *MockUserRepository) GetTotalActiveCount() (int, error) {
	return m.GetTotalActiveCountFn()
}
func (m *MockUserRepository) GetTotalInactiveCount() (int, error) {
	return m.GetTotalInactiveCountFn()
}

// ===== MOCK STUDENT REPOSITORY =====
type MockStudentRepository struct {
	GetByUserIDFn       func(userID uuid.UUID) (*models.Student, error)
	GetByIDFn           func(id uuid.UUID) (*models.Student, error)
	CreateFn            func(student models.Student) (uuid.UUID, error)
	GetAllFn            func() ([]models.Student, error)
	GetAllByAdvisorIDFn func(advisorID string) ([]models.Student, error)
	UpdateAdvisorFn     func(studentID uuid.UUID, advisorID *uuid.UUID) error
	RemoveAdvisorFn     func(studentID uuid.UUID) error
}

func (m *MockStudentRepository) GetByUserID(userID uuid.UUID) (*models.Student, error) {
	if m.GetByUserIDFn != nil {
		return m.GetByUserIDFn(userID)
	}
	return nil, nil
}

func (m *MockStudentRepository) GetByID(id uuid.UUID) (*models.Student, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(id)
	}
	return nil, nil
}

func (m *MockStudentRepository) Create(student models.Student) (uuid.UUID, error) {
	if m.CreateFn != nil {
		return m.CreateFn(student)
	}
	return uuid.Nil, nil
}

func (m *MockStudentRepository) GetAll() ([]models.Student, error) {
	if m.GetAllFn != nil {
		return m.GetAllFn()
	}
	return nil, nil
}

func (m *MockStudentRepository) GetAllByAdvisorID(advisorID string) ([]models.Student, error) {
	if m.GetAllByAdvisorIDFn != nil {
		return m.GetAllByAdvisorIDFn(advisorID)
	}
	return nil, nil
}

func (m *MockStudentRepository) UpdateAdvisor(studentID uuid.UUID, advisorID *uuid.UUID) error {
	if m.UpdateAdvisorFn != nil {
		return m.UpdateAdvisorFn(studentID, advisorID)
	}
	return nil
}

func (m *MockStudentRepository) RemoveAdvisor(studentID uuid.UUID) error {
	if m.RemoveAdvisorFn != nil {
		return m.RemoveAdvisorFn(studentID)
	}
	return nil
}

// ===== MOCK LECTURER REPOSITORY =====
type MockLecturerRepository struct {
	GetByIDFn              func(id uuid.UUID) (*models.Lecturer, error)
	GetByUserIDFn          func(userID uuid.UUID) (*models.Lecturer, error)
	GetByLecturerIDFn      func(lecturerID string) (*models.Lecturer, error)
	CreateFn               func(lecturer models.Lecturer) (uuid.UUID, error)
	UpdateFn               func(id uuid.UUID, req *models.UpdateLecturerRequest) error
	GetAllFn               func(page, limit int) ([]models.Lecturer, int, error)
	GetTotalCountFn        func() (int, error)
	GetWithUserDetailsFn   func(page, limit int) ([]models.LecturerResponse, int, error)
	GetAdviseesCountFn     func(lecturerID uuid.UUID) (int, error)
	GetAdviseesFn          func(lecturerID uuid.UUID, page, limit int) ([]models.Student, int, error)
	SearchByNameFn         func(name string, page, limit int) ([]models.LecturerResponse, int, error)
	GetByDepartmentFn      func(department string, page, limit int) ([]models.Lecturer, int, error)
}

func (m *MockLecturerRepository) GetByID(id uuid.UUID) (*models.Lecturer, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(id)
	}
	return nil, nil
}

func (m *MockLecturerRepository) GetByUserID(userID uuid.UUID) (*models.Lecturer, error) {
	if m.GetByUserIDFn != nil {
		return m.GetByUserIDFn(userID)
	}
	return nil, nil
}

func (m *MockLecturerRepository) GetByLecturerID(lecturerID string) (*models.Lecturer, error) {
	if m.GetByLecturerIDFn != nil {
		return m.GetByLecturerIDFn(lecturerID)
	}
	return nil, nil
}

func (m *MockLecturerRepository) Create(lecturer models.Lecturer) (uuid.UUID, error) {
	if m.CreateFn != nil {
		return m.CreateFn(lecturer)
	}
	return uuid.Nil, nil
}

func (m *MockLecturerRepository) Update(id uuid.UUID, req *models.UpdateLecturerRequest) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(id, req)
	}
	return nil
}

func (m *MockLecturerRepository) GetAll(page, limit int) ([]models.Lecturer, int, error) {
	if m.GetAllFn != nil {
		return m.GetAllFn(page, limit)
	}
	return nil, 0, nil
}

func (m *MockLecturerRepository) GetTotalCount() (int, error) {
	if m.GetTotalCountFn != nil {
		return m.GetTotalCountFn()
	}
	return 0, nil
}

func (m *MockLecturerRepository) GetWithUserDetails(page, limit int) ([]models.LecturerResponse, int, error) {
	if m.GetWithUserDetailsFn != nil {
		return m.GetWithUserDetailsFn(page, limit)
	}
	return nil, 0, nil
}

func (m *MockLecturerRepository) GetAdviseesCount(lecturerID uuid.UUID) (int, error) {
	if m.GetAdviseesCountFn != nil {
		return m.GetAdviseesCountFn(lecturerID)
	}
	return 0, nil
}

func (m *MockLecturerRepository) GetAdvisees(lecturerID uuid.UUID, page, limit int) ([]models.Student, int, error) {
	if m.GetAdviseesFn != nil {
		return m.GetAdviseesFn(lecturerID, page, limit)
	}
	return nil, 0, nil
}

func (m *MockLecturerRepository) SearchByName(name string, page, limit int) ([]models.LecturerResponse, int, error) {
	if m.SearchByNameFn != nil {
		return m.SearchByNameFn(name, page, limit)
	}
	return nil, 0, nil
}

func (m *MockLecturerRepository) GetByDepartment(department string, page, limit int) ([]models.Lecturer, int, error) {
	if m.GetByDepartmentFn != nil {
		return m.GetByDepartmentFn(department, page, limit)
	}
	return nil, 0, nil
}