package service

import (
	"context"
	"time"

	"UAS/app/models"

	"github.com/google/uuid"
)

// =======================
// AchievementReferenceRepository Mock
// =======================
type MockAchievementReferenceRepository struct {
	CreateReferenceFn          func(ref *models.AchievementReference) error
	GetReferenceByIDFn         func(id uuid.UUID) (*models.AchievementReference, error)
	GetReferenceByMongoIDFn    func(mongoID string) (*models.AchievementReference, error)
	UpdateReferenceFn          func(ref *models.AchievementReference) error
	DeleteReferenceFn          func(id uuid.UUID) error
	SoftDeleteFn               func(id uuid.UUID) error
	SubmitForVerificationFn    func(id uuid.UUID) error
	VerifyAchievementFn        func(id, verifiedBy uuid.UUID) error
	RejectAchievementFn        func(id, verifiedBy uuid.UUID, note string) error
	GetReferencesByStudentIDFn func(studentID uuid.UUID, status string) ([]models.AchievementReference, error)
	GetReferencesByAdvisorFn   func(advisorID uuid.UUID, status string) ([]models.AchievementReference, error)
	GetAllReferencesFn         func(status string, limit, offset int) ([]models.AchievementReference, int, error)
	CheckOwnershipFn           func(achievementID, studentID uuid.UUID) (bool, error)
}

func (m *MockAchievementReferenceRepository) CreateReference(ref *models.AchievementReference) error {
	if m.CreateReferenceFn == nil { panic("Mock Error: CreateReferenceFn is nil") }
	return m.CreateReferenceFn(ref)
}

func (m *MockAchievementReferenceRepository) GetReferenceByID(id uuid.UUID) (*models.AchievementReference, error) {
	if m.GetReferenceByIDFn == nil { panic("Mock Error: GetReferenceByIDFn is nil") }
	return m.GetReferenceByIDFn(id)
}

func (m *MockAchievementReferenceRepository) GetReferenceByMongoID(mongoID string) (*models.AchievementReference, error) {
	if m.GetReferenceByMongoIDFn == nil { panic("Mock Error: GetReferenceByMongoIDFn is nil") }
	return m.GetReferenceByMongoIDFn(mongoID)
}

func (m *MockAchievementReferenceRepository) UpdateReference(ref *models.AchievementReference) error {
	if m.UpdateReferenceFn == nil { panic("Mock Error: UpdateReferenceFn is nil") }
	return m.UpdateReferenceFn(ref)
}

func (m *MockAchievementReferenceRepository) DeleteReference(id uuid.UUID) error {
	if m.DeleteReferenceFn == nil { panic("Mock Error: DeleteReferenceFn is nil") }
	return m.DeleteReferenceFn(id)
}

func (m *MockAchievementReferenceRepository) SoftDelete(id uuid.UUID) error {
	if m.SoftDeleteFn == nil { panic("Mock Error: SoftDeleteFn is nil") }
	return m.SoftDeleteFn(id)
}

func (m *MockAchievementReferenceRepository) SubmitForVerification(id uuid.UUID) error {
	if m.SubmitForVerificationFn == nil { panic("Mock Error: SubmitForVerificationFn is nil") }
	return m.SubmitForVerificationFn(id)
}

func (m *MockAchievementReferenceRepository) VerifyAchievement(id, verifiedBy uuid.UUID) error {
	if m.VerifyAchievementFn == nil { panic("Mock Error: VerifyAchievementFn is nil") }
	return m.VerifyAchievementFn(id, verifiedBy)
}

func (m *MockAchievementReferenceRepository) RejectAchievement(id, verifiedBy uuid.UUID, note string) error {
	if m.RejectAchievementFn == nil { panic("Mock Error: RejectAchievementFn is nil") }
	return m.RejectAchievementFn(id, verifiedBy, note)
}

func (m *MockAchievementReferenceRepository) GetReferencesByStudentID(studentID uuid.UUID, status string) ([]models.AchievementReference, error) {
	if m.GetReferencesByStudentIDFn == nil { panic("Mock Error: GetReferencesByStudentIDFn is nil") }
	return m.GetReferencesByStudentIDFn(studentID, status)
}

func (m *MockAchievementReferenceRepository) GetReferencesByAdvisor(advisorID uuid.UUID, status string) ([]models.AchievementReference, error) {
	if m.GetReferencesByAdvisorFn == nil { panic("Mock Error: GetReferencesByAdvisorFn is nil") }
	return m.GetReferencesByAdvisorFn(advisorID, status)
}

func (m *MockAchievementReferenceRepository) GetAllReferences(status string, limit, offset int) ([]models.AchievementReference, int, error) {
	if m.GetAllReferencesFn == nil { panic("Mock Error: GetAllReferencesFn is nil") }
	return m.GetAllReferencesFn(status, limit, offset)
}

func (m *MockAchievementReferenceRepository) CheckOwnership(achievementID, studentID uuid.UUID) (bool, error) {
	if m.CheckOwnershipFn == nil { panic("Mock Error: CheckOwnershipFn is nil") }
	return m.CheckOwnershipFn(achievementID, studentID)
}

// =======================
// AchievementRepository Mock
// =======================
type MockAchievementRepository struct {
	CreateAchievementFn func(ctx context.Context, achievement *models.Achievement) (string, error)
	GetAchievementByIDFn func(ctx context.Context, id string) (*models.Achievement, error)
	UpdateAchievementFn func(ctx context.Context, id string, achievement *models.Achievement) error
	DeleteAchievementFn func(ctx context.Context, id string) error
	FindAchievementsFn  func(ctx context.Context, studentIDs []string, achievementType, search string, page, limit int, sortBy, sortOrder string) ([]models.Achievement, int64, error)
	GetAchievementsByIDsFn func(ctx context.Context, ids []string) ([]models.Achievement, error)
	AddAttachmentFn     func(ctx context.Context, achievementID string, attachment models.Attachment) error
	RemoveAttachmentFn  func(ctx context.Context, achievementID, fileName string) error
}

func (m *MockAchievementRepository) CreateAchievement(ctx context.Context, achievement *models.Achievement) (string, error) {
	if m.CreateAchievementFn == nil { panic("Mock Error: CreateAchievementFn is nil") }
	return m.CreateAchievementFn(ctx, achievement)
}

func (m *MockAchievementRepository) GetAchievementByID(ctx context.Context, id string) (*models.Achievement, error) {
	if m.GetAchievementByIDFn == nil { panic("Mock Error: GetAchievementByIDFn is nil") }
	return m.GetAchievementByIDFn(ctx, id)
}

func (m *MockAchievementRepository) UpdateAchievement(ctx context.Context, id string, achievement *models.Achievement) error {
	if m.UpdateAchievementFn == nil { panic("Mock Error: UpdateAchievementFn is nil") }
	return m.UpdateAchievementFn(ctx, id, achievement)
}

func (m *MockAchievementRepository) DeleteAchievement(ctx context.Context, id string) error {
	if m.DeleteAchievementFn == nil { panic("Mock Error: DeleteAchievementFn is nil") }
	return m.DeleteAchievementFn(ctx, id)
}

func (m *MockAchievementRepository) FindAchievements(ctx context.Context, studentIDs []string, achievementType, search string, page, limit int, sortBy, sortOrder string) ([]models.Achievement, int64, error) {
	if m.FindAchievementsFn == nil { panic("Mock Error: FindAchievementsFn is nil") }
	return m.FindAchievementsFn(ctx, studentIDs, achievementType, search, page, limit, sortBy, sortOrder)
}

func (m *MockAchievementRepository) GetAchievementsByIDs(ctx context.Context, ids []string) ([]models.Achievement, error) {
	if m.GetAchievementsByIDsFn == nil { panic("Mock Error: GetAchievementsByIDsFn is nil") }
	return m.GetAchievementsByIDsFn(ctx, ids)
}

func (m *MockAchievementRepository) AddAttachment(ctx context.Context, achievementID string, attachment models.Attachment) error {
	if m.AddAttachmentFn == nil { panic("Mock Error: AddAttachmentFn is nil") }
	return m.AddAttachmentFn(ctx, achievementID, attachment)
}

func (m *MockAchievementRepository) RemoveAttachment(ctx context.Context, achievementID, fileName string) error {
	if m.RemoveAttachmentFn == nil { panic("Mock Error: RemoveAttachmentFn is nil") }
	return m.RemoveAttachmentFn(ctx, achievementID, fileName)
}

// =======================
// LecturerRepository Mock
// =======================
type MockLecturerRepository struct {
	GetByIDFn            func(id uuid.UUID) (*models.Lecturer, error)
	GetByUserIDFn        func(userID uuid.UUID) (*models.Lecturer, error)
	GetByLecturerIDFn    func(lecturerID string) (*models.Lecturer, error)
	CreateFn             func(lecturer models.Lecturer) (uuid.UUID, error)
	UpdateFn             func(id uuid.UUID, req *models.UpdateLecturerRequest) error
	GetAllFn             func(page, limit int) ([]models.Lecturer, int, error)
	GetTotalCountFn      func() (int, error)
	GetWithUserDetailsFn func(page, limit int) ([]models.LecturerResponse, int, error)
	GetAdviseesCountFn   func(lecturerID uuid.UUID) (int, error)
	GetAdviseesFn        func(lecturerID uuid.UUID, page, limit int) ([]models.Student, int, error)
	SearchByNameFn       func(name string, page, limit int) ([]models.LecturerResponse, int, error)
	GetByDepartmentFn    func(department string, page, limit int) ([]models.Lecturer, int, error)
}

func (m *MockLecturerRepository) GetByID(id uuid.UUID) (*models.Lecturer, error) {
	if m.GetByIDFn == nil { panic("Mock Error: GetByIDFn is nil") }
	return m.GetByIDFn(id)
}

func (m *MockLecturerRepository) GetByUserID(userID uuid.UUID) (*models.Lecturer, error) {
	if m.GetByUserIDFn == nil { panic("Mock Error: GetByUserIDFn is nil") }
	return m.GetByUserIDFn(userID)
}

func (m *MockLecturerRepository) GetByLecturerID(lecturerID string) (*models.Lecturer, error) {
	if m.GetByLecturerIDFn == nil { panic("Mock Error: GetByLecturerIDFn is nil") }
	return m.GetByLecturerIDFn(lecturerID)
}

func (m *MockLecturerRepository) Create(lecturer models.Lecturer) (uuid.UUID, error) {
	if m.CreateFn == nil { panic("Mock Error: CreateFn is nil") }
	return m.CreateFn(lecturer)
}

func (m *MockLecturerRepository) Update(id uuid.UUID, req *models.UpdateLecturerRequest) error {
	if m.UpdateFn == nil { panic("Mock Error: UpdateFn is nil") }
	return m.UpdateFn(id, req)
}

func (m *MockLecturerRepository) GetAll(page, limit int) ([]models.Lecturer, int, error) {
	if m.GetAllFn == nil { panic("Mock Error: GetAllFn is nil") }
	return m.GetAllFn(page, limit)
}

func (m *MockLecturerRepository) GetTotalCount() (int, error) {
	if m.GetTotalCountFn == nil { panic("Mock Error: GetTotalCountFn is nil") }
	return m.GetTotalCountFn()
}

func (m *MockLecturerRepository) GetWithUserDetails(page, limit int) ([]models.LecturerResponse, int, error) {
	if m.GetWithUserDetailsFn == nil { panic("Mock Error: GetWithUserDetailsFn is nil") }
	return m.GetWithUserDetailsFn(page, limit)
}

func (m *MockLecturerRepository) GetAdviseesCount(lecturerID uuid.UUID) (int, error) {
	if m.GetAdviseesCountFn == nil { panic("Mock Error: GetAdviseesCountFn is nil") }
	return m.GetAdviseesCountFn(lecturerID)
}

func (m *MockLecturerRepository) GetAdvisees(lecturerID uuid.UUID, page, limit int) ([]models.Student, int, error) {
	if m.GetAdviseesFn == nil { panic("Mock Error: GetAdviseesFn is nil") }
	return m.GetAdviseesFn(lecturerID, page, limit)
}

func (m *MockLecturerRepository) SearchByName(name string, page, limit int) ([]models.LecturerResponse, int, error) {
	if m.SearchByNameFn == nil { panic("Mock Error: SearchByNameFn is nil") }
	return m.SearchByNameFn(name, page, limit)
}

func (m *MockLecturerRepository) GetByDepartment(department string, page, limit int) ([]models.Lecturer, int, error) {
	if m.GetByDepartmentFn == nil { panic("Mock Error: GetByDepartmentFn is nil") }
	return m.GetByDepartmentFn(department, page, limit)
}

// =======================
// ReportRepository Mock
// =======================
type MockReportRepository struct {
	GetStatisticsFn func(ctx context.Context, actorID uuid.UUID, scope string, startDate, endDate *time.Time) (*models.AchievementStats, error)
}

func (m *MockReportRepository) GetStatistics(ctx context.Context, actorID uuid.UUID, scope string, startDate, endDate *time.Time) (*models.AchievementStats, error) {
	if m.GetStatisticsFn == nil { panic("Mock Error: GetStatisticsFn is nil") }
	return m.GetStatisticsFn(ctx, actorID, scope, startDate, endDate)
}

// =======================
// RoleRepository Mock
// =======================
type MockRoleRepository struct {
	GetByIDFn                func(id uuid.UUID) (*models.Role, error)
	GetByNameFn              func(name string) (*models.Role, error)
	GetAllFn                 func(page, limit int) ([]models.Role, int, error)
	GetTotalCountFn          func() (int, error)
	GetPermissionsByRoleIDFn func(roleID uuid.UUID) ([]models.Permission, error)
	GetPermissionNamesByRoleIDFn func(roleID uuid.UUID) ([]string, error)
	AssignPermissionFn       func(roleID, permissionID uuid.UUID) error
	RemovePermissionFn       func(roleID, permissionID uuid.UUID) error
}

func (m *MockRoleRepository) GetByID(id uuid.UUID) (*models.Role, error) {
	if m.GetByIDFn == nil { panic("Mock Error: GetByIDFn is nil") }
	return m.GetByIDFn(id)
}

func (m *MockRoleRepository) GetByName(name string) (*models.Role, error) {
	if m.GetByNameFn == nil { panic("Mock Error: GetByNameFn is nil") }
	return m.GetByNameFn(name)
}

func (m *MockRoleRepository) GetAll(page, limit int) ([]models.Role, int, error) {
	if m.GetAllFn == nil { panic("Mock Error: GetAllFn is nil") }
	return m.GetAllFn(page, limit)
}

func (m *MockRoleRepository) GetTotalCount() (int, error) {
	if m.GetTotalCountFn == nil { panic("Mock Error: GetTotalCountFn is nil") }
	return m.GetTotalCountFn()
}

func (m *MockRoleRepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]models.Permission, error) {
	if m.GetPermissionsByRoleIDFn == nil { panic("Mock Error: GetPermissionsByRoleIDFn is nil") }
	return m.GetPermissionsByRoleIDFn(roleID)
}

func (m *MockRoleRepository) GetPermissionNamesByRoleID(roleID uuid.UUID) ([]string, error) {
	if m.GetPermissionNamesByRoleIDFn == nil { panic("Mock Error: GetPermissionNamesByRoleIDFn is nil") }
	return m.GetPermissionNamesByRoleIDFn(roleID)
}

func (m *MockRoleRepository) AssignPermission(roleID, permissionID uuid.UUID) error {
	if m.AssignPermissionFn == nil { panic("Mock Error: AssignPermissionFn is nil") }
	return m.AssignPermissionFn(roleID, permissionID)
}

func (m *MockRoleRepository) RemovePermission(roleID, permissionID uuid.UUID) error {
	if m.RemovePermissionFn == nil { panic("Mock Error: RemovePermissionFn is nil") }
	return m.RemovePermissionFn(roleID, permissionID)
}

// =======================
// StudentRepository Mock
// =======================
type MockStudentRepository struct {
	GetByUserIDFn         func(userID uuid.UUID) (*models.Student, error)
	GetByIDFn             func(id uuid.UUID) (*models.Student, error)
	CreateFn              func(student models.Student) (uuid.UUID, error)
	GetAllFn              func() ([]models.Student, error)
	GetAllByAdvisorIDFn   func(advisorID string) ([]models.Student, error)
	UpdateAdvisorFn       func(studentID uuid.UUID, advisorID *uuid.UUID) error
	RemoveAdvisorFn       func(studentID uuid.UUID) error
}

func (m *MockStudentRepository) GetByUserID(userID uuid.UUID) (*models.Student, error) {
	if m.GetByUserIDFn == nil { panic("Mock Error: GetByUserIDFn is nil") }
	return m.GetByUserIDFn(userID)
}

func (m *MockStudentRepository) GetByID(id uuid.UUID) (*models.Student, error) {
	if m.GetByIDFn == nil { panic("Mock Error: GetByIDFn is nil") }
	return m.GetByIDFn(id)
}

func (m *MockStudentRepository) Create(student models.Student) (uuid.UUID, error) {
	if m.CreateFn == nil { panic("Mock Error: CreateFn is nil") }
	return m.CreateFn(student)
}

func (m *MockStudentRepository) GetAll() ([]models.Student, error) {
	if m.GetAllFn == nil { panic("Mock Error: GetAllFn is nil") }
	return m.GetAllFn()
}

func (m *MockStudentRepository) GetAllByAdvisorID(advisorID string) ([]models.Student, error) {
	if m.GetAllByAdvisorIDFn == nil { panic("Mock Error: GetAllByAdvisorIDFn is nil") }
	return m.GetAllByAdvisorIDFn(advisorID)
}

func (m *MockStudentRepository) UpdateAdvisor(studentID uuid.UUID, advisorID *uuid.UUID) error {
	if m.UpdateAdvisorFn == nil { panic("Mock Error: UpdateAdvisorFn is nil") }
	return m.UpdateAdvisorFn(studentID, advisorID)
}

func (m *MockStudentRepository) RemoveAdvisor(studentID uuid.UUID) error {
	if m.RemoveAdvisorFn == nil { panic("Mock Error: RemoveAdvisorFn is nil") }
	return m.RemoveAdvisorFn(studentID)
}

// =======================
// UserRepository Mock
// =======================
type MockUserRepository struct {
	GetByIDFn                func(id uuid.UUID) (*models.User, error)
	GetByEmailFn             func(email string) (*models.User, error)
	GetByUsernameFn          func(username string) (*models.User, error)
	GetByUsernameOrEmailFn   func(identifier string) (*models.User, error)
	CreateFn                 func(user *models.User) (uuid.UUID, error)
	UpdateFn                 func(id uuid.UUID, req *models.UpdateUserRequest) error
	UpdatePasswordFn         func(id uuid.UUID, hashedPassword string) error
	SoftDeleteFn             func(id uuid.UUID) error
	HardDeleteFn             func(id uuid.UUID) error
	GetAllFn                 func(page, limit int) ([]models.User, int, error)
	GetInactiveUsersFn       func(page, limit int) ([]models.User, int, error)
	GetAllWithInactiveFn     func(page, limit int) ([]models.User, int, error)
	SearchByNameFn           func(name string, page, limit int) ([]models.User, int, error)
	GetByRoleFn              func(roleID uuid.UUID, page, limit int) ([]models.User, int, error)
	GetUsersCountByRoleFn    func() (map[uuid.UUID]int, error)
	GetTotalActiveCountFn    func() (int, error)
	GetTotalInactiveCountFn  func() (int, error)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
    if m.GetByIDFn == nil {
        return &models.User{}, nil
    }
    return m.GetByIDFn(id)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	if m.GetByEmailFn == nil { panic("Mock Error: GetByEmailFn is nil") }
	return m.GetByEmailFn(email)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	if m.GetByUsernameFn == nil { panic("Mock Error: GetByUsernameFn is nil") }
	return m.GetByUsernameFn(username)
}

func (m *MockUserRepository) GetByUsernameOrEmail(identifier string) (*models.User, error) {
	if m.GetByUsernameOrEmailFn == nil { panic("Mock Error: GetByUsernameOrEmailFn is nil") }
	return m.GetByUsernameOrEmailFn(identifier)
}

func (m *MockUserRepository) Create(user *models.User) (uuid.UUID, error) {
	if m.CreateFn == nil { panic("Mock Error: CreateFn is nil") }
	return m.CreateFn(user)
}

func (m *MockUserRepository) Update(id uuid.UUID, req *models.UpdateUserRequest) error {
	if m.UpdateFn == nil { panic("Mock Error: UpdateFn is nil") }
	return m.UpdateFn(id, req)
}

func (m *MockUserRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	if m.UpdatePasswordFn == nil { panic("Mock Error: UpdatePasswordFn is nil") }
	return m.UpdatePasswordFn(id, hashedPassword)
}

func (m *MockUserRepository) SoftDelete(id uuid.UUID) error {
	if m.SoftDeleteFn == nil { panic("Mock Error: SoftDeleteFn is nil") }
	return m.SoftDeleteFn(id)
}

func (m *MockUserRepository) HardDelete(id uuid.UUID) error {
	if m.HardDeleteFn == nil { panic("Mock Error: HardDeleteFn is nil") }
	return m.HardDeleteFn(id)
}

func (m *MockUserRepository) GetAll(page, limit int) ([]models.User, int, error) {
	if m.GetAllFn == nil { panic("Mock Error: GetAllFn is nil") }
	return m.GetAllFn(page, limit)
}

func (m *MockUserRepository) GetInactiveUsers(page, limit int) ([]models.User, int, error) {
	if m.GetInactiveUsersFn == nil { panic("Mock Error: GetInactiveUsersFn is nil") }
	return m.GetInactiveUsersFn(page, limit)
}

func (m *MockUserRepository) GetAllWithInactive(page, limit int) ([]models.User, int, error) {
	if m.GetAllWithInactiveFn == nil { panic("Mock Error: GetAllWithInactiveFn is nil") }
	return m.GetAllWithInactiveFn(page, limit)
}

func (m *MockUserRepository) SearchByName(name string, page, limit int) ([]models.User, int, error) {
	if m.SearchByNameFn == nil { panic("Mock Error: SearchByNameFn is nil") }
	return m.SearchByNameFn(name, page, limit)
}

func (m *MockUserRepository) GetByRole(roleID uuid.UUID, page, limit int) ([]models.User, int, error) {
	if m.GetByRoleFn == nil { panic("Mock Error: GetByRoleFn is nil") }
	return m.GetByRoleFn(roleID, page, limit)
}

func (m *MockUserRepository) GetUsersCountByRole() (map[uuid.UUID]int, error) {
	if m.GetUsersCountByRoleFn == nil { panic("Mock Error: GetUsersCountByRoleFn is nil") }
	return m.GetUsersCountByRoleFn()
}

func (m *MockUserRepository) GetTotalActiveCount() (int, error) {
	if m.GetTotalActiveCountFn == nil { panic("Mock Error: GetTotalActiveCountFn is nil") }
	return m.GetTotalActiveCountFn()
}

func (m *MockUserRepository) GetTotalInactiveCount() (int, error) {
	if m.GetTotalInactiveCountFn == nil { panic("Mock Error: GetTotalInactiveCountFn is nil") }
	return m.GetTotalInactiveCountFn()
}