package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSprintCalculateStatus(t *testing.T) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		sprint         Sprint
		expectedStatus string
	}{
		{
			name: "closed sprint",
			sprint: Sprint{
				ID:        uuid.New(),
				StartDate: today.AddDate(0, 0, -10),
				EndDate:   today.AddDate(0, 0, -1),
				IsClosed:  true,
			},
			expectedStatus: "closed",
		},
		{
			name: "closed sprint in future (shouldn't happen but test behavior)",
			sprint: Sprint{
				ID:        uuid.New(),
				StartDate: today.AddDate(0, 0, 5),
				EndDate:   today.AddDate(0, 0, 10),
				IsClosed:  true,
			},
			expectedStatus: "closed",
		},
		{
			name: "upcoming sprint",
			sprint: Sprint{
				ID:        uuid.New(),
				StartDate: today.AddDate(0, 0, 5),
				EndDate:   today.AddDate(0, 0, 19),
				IsClosed:  false,
			},
			expectedStatus: "upcoming",
		},
		{
			name: "active sprint - started today",
			sprint: Sprint{
				ID:        uuid.New(),
				StartDate: today,
				EndDate:   today.AddDate(0, 0, 13),
				IsClosed:  false,
			},
			expectedStatus: "active",
		},
		{
			name: "active sprint - in progress",
			sprint: Sprint{
				ID:        uuid.New(),
				StartDate: today.AddDate(0, 0, -5),
				EndDate:   today.AddDate(0, 0, 8),
				IsClosed:  false,
			},
			expectedStatus: "active",
		},
		{
			name: "active sprint - ends today",
			sprint: Sprint{
				ID:        uuid.New(),
				StartDate: today.AddDate(0, 0, -13),
				EndDate:   today,
				IsClosed:  false,
			},
			expectedStatus: "active",
		},
		{
			name: "completed sprint",
			sprint: Sprint{
				ID:        uuid.New(),
				StartDate: today.AddDate(0, 0, -20),
				EndDate:   today.AddDate(0, 0, -1),
				IsClosed:  false,
			},
			expectedStatus: "completed",
		},
		{
			name: "completed sprint - ended yesterday",
			sprint: Sprint{
				ID:        uuid.New(),
				StartDate: today.AddDate(0, 0, -14),
				EndDate:   today.AddDate(0, 0, -1),
				IsClosed:  false,
			},
			expectedStatus: "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sprint.CalculateStatus()
			assert.Equal(t, tt.expectedStatus, tt.sprint.Status)
		})
	}
}

func TestSprintCalculateStatusWithTime(t *testing.T) {
	// Test that time of day doesn't affect date-only comparison
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Create sprint with times during the day (not midnight)
	sprint := Sprint{
		ID:        uuid.New(),
		StartDate: today.Add(10 * time.Hour),  // 10 AM today
		EndDate:   today.AddDate(0, 0, 13).Add(15 * time.Hour), // 3 PM in 13 days
		IsClosed:  false,
	}

	sprint.CalculateStatus()
	assert.Equal(t, "active", sprint.Status, "Sprint should be active regardless of time of day")
}

func TestSprintCalculateStatusEdgeCases(t *testing.T) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		startDate      time.Time
		endDate        time.Time
		isClosed       bool
		expectedStatus string
	}{
		{
			name:           "single day sprint - today",
			startDate:      today,
			endDate:        today,
			isClosed:       false,
			expectedStatus: "active",
		},
		{
			name:           "single day sprint - tomorrow",
			startDate:      today.AddDate(0, 0, 1),
			endDate:        today.AddDate(0, 0, 1),
			isClosed:       false,
			expectedStatus: "upcoming",
		},
		{
			name:           "single day sprint - yesterday",
			startDate:      today.AddDate(0, 0, -1),
			endDate:        today.AddDate(0, 0, -1),
			isClosed:       false,
			expectedStatus: "completed",
		},
		{
			name:           "closed overrides upcoming",
			startDate:      today.AddDate(0, 0, 10),
			endDate:        today.AddDate(0, 0, 23),
			isClosed:       true,
			expectedStatus: "closed",
		},
		{
			name:           "closed overrides completed",
			startDate:      today.AddDate(0, 0, -20),
			endDate:        today.AddDate(0, 0, -1),
			isClosed:       true,
			expectedStatus: "closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprint := Sprint{
				ID:        uuid.New(),
				StartDate: tt.startDate,
				EndDate:   tt.endDate,
				IsClosed:  tt.isClosed,
			}
			sprint.CalculateStatus()
			assert.Equal(t, tt.expectedStatus, sprint.Status)
		})
	}
}
