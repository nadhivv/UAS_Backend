package models

import "github.com/google/uuid"

type AchievementStats struct {
	TotalByType            []StatItem          `json:"total_by_type"`
	TotalByPeriod          []StatItem          `json:"total_by_period"`
	TopStudents            []TopStudentStat    `json:"top_students,omitempty"`
	CompetitionDistribution []StatItem         `json:"competition_distribution"`
}

type StatItem struct {
	Key   string `json:"key"`
	Total int    `json:"total"`
}

type TopStudentStat struct {
	StudentID   uuid.UUID `json:"student_id"`
	StudentName string    `json:"student_name"`
	TotalAchievements int `json:"total_achievements"`
	TotalPoints int      `json:"total_points"`
}
