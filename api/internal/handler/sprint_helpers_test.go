package handler

import (
	"testing"
	"time"

	"github.com/cfegela/flyhalf/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "Monday",
			date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // Monday
			expected: false,
		},
		{
			name:     "Tuesday",
			date:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), // Tuesday
			expected: false,
		},
		{
			name:     "Wednesday",
			date:     time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), // Wednesday
			expected: false,
		},
		{
			name:     "Thursday",
			date:     time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), // Thursday
			expected: false,
		},
		{
			name:     "Friday",
			date:     time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), // Friday
			expected: false,
		},
		{
			name:     "Saturday",
			date:     time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), // Saturday
			expected: true,
		},
		{
			name:     "Sunday",
			date:     time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), // Sunday
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWeekend(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountWorkingDays(t *testing.T) {
	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		expected  int
	}{
		{
			name:      "single weekday",
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),  // Monday
			endDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),  // Monday
			expected:  1,
		},
		{
			name:      "Monday to Friday (5 days)",
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),  // Monday
			endDate:   time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),  // Friday
			expected:  5,
		},
		{
			name:      "full week including weekend",
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),  // Monday
			endDate:   time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),  // Sunday
			expected:  5, // Only weekdays count
		},
		{
			name:      "two weeks (14 days)",
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),  // Monday
			endDate:   time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), // Sunday (2 weeks later)
			expected:  10, // 10 weekdays
		},
		{
			name:      "weekend only",
			startDate: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC),  // Saturday
			endDate:   time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),  // Sunday
			expected:  0, // No weekdays
		},
		{
			name:      "Friday to Monday (includes weekend)",
			startDate: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),  // Friday
			endDate:   time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),  // Monday
			expected:  2, // Friday and Monday only
		},
		{
			name:      "typical 2-week sprint (14 days)",
			startDate: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),  // Monday
			endDate:   time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC), // Sunday (14 days)
			expected:  10, // 10 weekdays
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countWorkingDays(tt.startDate, tt.endDate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateIdealBurndown(t *testing.T) {
	tests := []struct {
		name        string
		startDate   time.Time
		endDate     time.Time
		totalPoints int
		validate    func(t *testing.T, points []BurndownPoint)
	}{
		{
			name:        "2-week sprint starting Monday (50 points)",
			startDate:   time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),  // Monday
			endDate:     time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC), // Sunday (14 days)
			totalPoints: 50,
			validate: func(t *testing.T, points []BurndownPoint) {
				// Should have 11 points: day before (if not weekend) + 10 working days
				assert.GreaterOrEqual(t, len(points), 10)

				// First point should be total points
				assert.Equal(t, 50, points[0].Points)

				// Last point should be 0 or close to 0
				assert.LessOrEqual(t, points[len(points)-1].Points, 5)

				// Points should decrease monotonically (or stay same on weekends)
				for i := 1; i < len(points); i++ {
					assert.LessOrEqual(t, points[i].Points, points[i-1].Points)
				}
			},
		},
		{
			name:        "single working day sprint",
			startDate:   time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),  // Monday
			endDate:     time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),  // Monday
			totalPoints: 10,
			validate: func(t *testing.T, points []BurndownPoint) {
				assert.GreaterOrEqual(t, len(points), 1)
				// Should start at 10 and end at 0
				assert.Equal(t, 10, points[0].Points)
			},
		},
		{
			name:        "weekend sprint (edge case)",
			startDate:   time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC),  // Saturday
			endDate:     time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),  // Sunday
			totalPoints: 20,
			validate: func(t *testing.T, points []BurndownPoint) {
				// May have a day before point if it's not a weekend
				// But should not include Saturday or Sunday
				for _, point := range points {
					date, _ := time.Parse("2006-01-02", point.Date)
					weekday := date.Weekday()
					assert.False(t, weekday == time.Saturday || weekday == time.Sunday,
						"burndown should not include weekend dates")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := generateIdealBurndown(tt.startDate, tt.endDate, tt.totalPoints)
			tt.validate(t, points)
		})
	}
}

func TestGenerateIdealBurndownNoWeekends(t *testing.T) {
	// 2-week sprint starting Monday
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)  // Monday
	endDate := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)   // Sunday
	totalPoints := 50

	points := generateIdealBurndown(startDate, endDate, totalPoints)

	// Verify no weekend dates in burndown
	for _, point := range points {
		date, err := time.Parse("2006-01-02", point.Date)
		assert.NoError(t, err)
		weekday := date.Weekday()
		assert.False(t, weekday == time.Saturday || weekday == time.Sunday,
			"ideal burndown should skip weekends: found %s", point.Date)
	}
}

func TestGenerateActualBurndown(t *testing.T) {
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)  // Monday
	endDate := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)   // Sunday
	totalPoints := 50

	size5 := 5
	size10 := 10
	size15 := 15
	size20 := 20

	tickets := []*model.Ticket{
		{
			ID:        uuid.New(),
			Status:    "closed",
			Size:      &size10,
			UpdatedAt: time.Date(2024, 1, 9, 14, 0, 0, 0, time.UTC), // Closed on Tuesday
		},
		{
			ID:        uuid.New(),
			Status:    "closed",
			Size:      &size15,
			UpdatedAt: time.Date(2024, 1, 11, 10, 0, 0, 0, time.UTC), // Closed on Thursday
		},
		{
			ID:        uuid.New(),
			Status:    "in_progress",
			Size:      &size20,
			UpdatedAt: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC), // Not closed
		},
		{
			ID:        uuid.New(),
			Status:    "closed",
			Size:      &size5,
			UpdatedAt: time.Date(2024, 1, 16, 16, 0, 0, 0, time.UTC), // Closed on Tuesday next week
		},
	}

	points := generateActualBurndown(startDate, endDate, totalPoints, tickets)

	// Verify structure
	assert.NotEmpty(t, points)

	// First point should be total points (day before or first day)
	assert.Equal(t, totalPoints, points[0].Points)

	// Verify no weekend dates
	for _, point := range points {
		date, err := time.Parse("2006-01-02", point.Date)
		assert.NoError(t, err)
		weekday := date.Weekday()
		assert.False(t, weekday == time.Saturday || weekday == time.Sunday,
			"actual burndown should skip weekends: found %s", point.Date)
	}

	// Points should decrease as tickets are completed
	// After Jan 9: 50 - 10 = 40
	// After Jan 11: 40 - 15 = 25
	// After Jan 16: 25 - 5 = 20 (20 points remain from in_progress ticket)
}

func TestGenerateActualBurndownNoTickets(t *testing.T) {
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)  // Monday
	endDate := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)   // Sunday
	totalPoints := 50

	points := generateActualBurndown(startDate, endDate, totalPoints, []*model.Ticket{})

	// Should still generate burndown with no progress
	assert.NotEmpty(t, points)

	// All points should remain at totalPoints (no tickets closed)
	for _, point := range points {
		assert.Equal(t, totalPoints, point.Points)
	}
}

func TestGenerateActualBurndownAllTicketsClosed(t *testing.T) {
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)  // Monday
	endDate := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)   // Sunday
	totalPoints := 30

	size10 := 10
	size20 := 20

	tickets := []*model.Ticket{
		{
			ID:        uuid.New(),
			Status:    "closed",
			Size:      &size10,
			UpdatedAt: time.Date(2024, 1, 9, 14, 0, 0, 0, time.UTC),
		},
		{
			ID:        uuid.New(),
			Status:    "closed",
			Size:      &size20,
			UpdatedAt: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		},
	}

	points := generateActualBurndown(startDate, endDate, totalPoints, tickets)

	// Find a point after all tickets are closed (e.g., Jan 12)
	var foundPoint *BurndownPoint
	for _, point := range points {
		if point.Date >= "2024-01-12" {
			foundPoint = &point
			break
		}
	}

	assert.NotNil(t, foundPoint)
	assert.Equal(t, 0, foundPoint.Points, "all tickets closed, should reach 0 points")
}

func TestGenerateActualBurndownTicketsWithoutSize(t *testing.T) {
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	totalPoints := 20

	size10 := 10

	tickets := []*model.Ticket{
		{
			ID:        uuid.New(),
			Status:    "closed",
			Size:      &size10,
			UpdatedAt: time.Date(2024, 1, 9, 14, 0, 0, 0, time.UTC),
		},
		{
			ID:        uuid.New(),
			Status:    "closed",
			Size:      nil, // No size
			UpdatedAt: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		},
	}

	points := generateActualBurndown(startDate, endDate, totalPoints, tickets)

	// Should handle nil size as 0 points
	assert.NotEmpty(t, points)

	// After Jan 10, should have 20 - 10 = 10 points remaining (second ticket has no size)
	var jan10Point *BurndownPoint
	for _, point := range points {
		if point.Date >= "2024-01-10" {
			jan10Point = &point
			break
		}
	}

	assert.NotNil(t, jan10Point)
	assert.Equal(t, 10, jan10Point.Points)
}
