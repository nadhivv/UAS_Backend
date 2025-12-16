package repository

import (
	"context"
	"time"

	"UAS/app/models"
	"UAS/database"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
)

type ReportRepository interface {
	GetStatistics(
		ctx context.Context,
		actorID uuid.UUID,
		scope string,
		startDate *time.Time,
		endDate *time.Time,
	) (*models.AchievementStats, error)
}

type reportRepo struct{}

func NewReportRepository() ReportRepository {
	return &reportRepo{}
}

func (r *reportRepo) GetStatistics(
	ctx context.Context,
	actorID uuid.UUID,
	scope string,
	startDate *time.Time,
	endDate *time.Time,
) (*models.AchievementStats, error) {

	stats := &models.AchievementStats{
		TotalByType:             []models.StatItem{},
		TotalByPeriod:           []models.StatItem{},
		TopStudents:             []models.TopStudentStat{},
		CompetitionDistribution: []models.StatItem{},
	}

	var studentIDs []uuid.UUID

	switch scope {
	case "student":
		studentIDs = []uuid.UUID{actorID}

	case "lecturer":
		rows, err := database.PgDB.QueryContext(
			ctx,
			`SELECT id FROM students WHERE advisor_id = $1`,
			actorID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err == nil {
				studentIDs = append(studentIDs, id)
			}
		}

	case "all":
		rows, err := database.PgDB.QueryContext(
			ctx,
			`SELECT id FROM students`,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err == nil {
				studentIDs = append(studentIDs, id)
			}
		}
	}

	if len(studentIDs) == 0 {
		return stats, nil
	}

	query := `
		SELECT TO_CHAR(verified_at, 'YYYY-MM') AS period, COUNT(*)
		FROM achievement_references
		WHERE status = 'verified'
		AND student_id = ANY($1)
	`
	args := []interface{}{pq.Array(studentIDs)}
	param := 2

	if startDate != nil {
		query += " AND verified_at >= $" + string(rune('0'+param))
		args = append(args, *startDate)
		param++
	}

	if endDate != nil {
		query += " AND verified_at <= $" + string(rune('0'+param))
		args = append(args, *endDate)
		param++
	}

	query += " GROUP BY period ORDER BY period"

	rows, err := database.PgDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.StatItem
		if err := rows.Scan(&item.Key, &item.Total); err == nil {
			stats.TotalByPeriod = append(stats.TotalByPeriod, item)
		}
	}

	if scope == "all" || scope == "lecturer" {
		topQuery := `
			SELECT 
				s.id,
				u.full_name,
				COUNT(ar.id),
				COUNT(ar.id) * 10
			FROM students s
			JOIN users u ON u.id = s.user_id
			JOIN achievement_references ar ON ar.student_id = s.id
			WHERE ar.status = 'verified'
		`

		var params []interface{}
		if scope == "lecturer" {
			topQuery += " AND s.advisor_id = $1"
			params = append(params, actorID)
		}

		topQuery += `
			GROUP BY s.id, u.full_name
			ORDER BY COUNT(ar.id) DESC
			LIMIT 10
		`

		rows, err := database.PgDB.QueryContext(ctx, topQuery, params...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var s models.TopStudentStat
				if err := rows.Scan(
					&s.StudentID,
					&s.StudentName,
					&s.TotalAchievements,
					&s.TotalPoints,
				); err == nil {
					stats.TopStudents = append(stats.TopStudents, s)
				}
			}
		}
	}

	var studentIDStrings []string
	for _, id := range studentIDs {
		studentIDStrings = append(studentIDStrings, id.String())
	}

	typePipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDStrings}}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cur, err := database.MongoDB.Collection("achievements").Aggregate(ctx, typePipeline)
	if err == nil {
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var row struct {
				Key   string `bson:"_id"`
				Total int    `bson:"total"`
			}
			if err := cur.Decode(&row); err == nil {
				stats.TotalByType = append(stats.TotalByType, models.StatItem{
					Key:   row.Key,
					Total: row.Total,
				})
			}
		}
	}

	compPipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "achievementType", Value: "competition"},
			{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDStrings}}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$details.competitionLevel"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	compCur, err := database.MongoDB.Collection("achievements").Aggregate(ctx, compPipeline)
	if err == nil {
		defer compCur.Close(ctx)
		for compCur.Next(ctx) {
			var row struct {
				Key   string `bson:"_id"`
				Total int    `bson:"total"`
			}
			if err := compCur.Decode(&row); err == nil {
				stats.CompetitionDistribution = append(
					stats.CompetitionDistribution,
					models.StatItem{Key: row.Key, Total: row.Total},
				)
			}
		}
	}

	return stats, nil
}
