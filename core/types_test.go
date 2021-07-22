package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CalendarMonthTestSuite struct {
	suite.Suite
}

func TestCalendarMonthTestSuite(t *testing.T) {
	suite.Run(t, new(CalendarMonthTestSuite))
}

// -- SUITE

func (suite *CalendarMonthTestSuite) Test_GIVEN_aCalendarMonth_WHEN_calculatingFirstDay_THEN_firstDayIsFirstOfMonth() {
	// GIVEN
	calendarMonth := MakeCalendarMonth(2020, time.January)

	// WHEN
	firstDay := calendarMonth.FirstDay()

	// THEN
	assert.Equal(suite.T(), 2020, firstDay.Year())
	assert.Equal(suite.T(), time.January, firstDay.Month())
	assert.Equal(suite.T(), 1, firstDay.Day())

	assert.Equal(suite.T(), 0, firstDay.Hour())
	assert.Equal(suite.T(), 0, firstDay.Minute())
	assert.Equal(suite.T(), 0, firstDay.Second())
	assert.Equal(suite.T(), 0, firstDay.Nanosecond())
	assert.Equal(suite.T(), time.UTC, firstDay.Location())
}

func (suite *CalendarMonthTestSuite) Test_GIVEN_aCalendarMonth_WHEN_calculatingLastDay_THEN_lastDayIsLastDayOfMonth() {
	// THEN
	assert.Equal(suite.T(), 31, MakeCalendarMonth(2020, time.January).LastDay().Day())
	assert.Equal(suite.T(), 29, MakeCalendarMonth(2020, time.February).LastDay().Day())
	assert.Equal(suite.T(), 31, MakeCalendarMonth(2020, time.March).LastDay().Day())
	assert.Equal(suite.T(), 30, MakeCalendarMonth(2020, time.April).LastDay().Day())
	assert.Equal(suite.T(), 31, MakeCalendarMonth(2020, time.May).LastDay().Day())
	assert.Equal(suite.T(), 30, MakeCalendarMonth(2020, time.June).LastDay().Day())
	assert.Equal(suite.T(), 31, MakeCalendarMonth(2020, time.July).LastDay().Day())
	assert.Equal(suite.T(), 31, MakeCalendarMonth(2020, time.August).LastDay().Day())
	assert.Equal(suite.T(), 30, MakeCalendarMonth(2020, time.September).LastDay().Day())
	assert.Equal(suite.T(), 31, MakeCalendarMonth(2020, time.October).LastDay().Day())
	assert.Equal(suite.T(), 30, MakeCalendarMonth(2020, time.November).LastDay().Day())
	assert.Equal(suite.T(), 31, MakeCalendarMonth(2020, time.December).LastDay().Day())
	assert.Equal(suite.T(), 28, MakeCalendarMonth(2021, time.February).LastDay().Day())
}

func (suite *CalendarMonthTestSuite) Test_GIVEN_aCalendarMonth_WHEN_calculatingNextMonth_THEN_nextMonthIsCorrect() {
	// THEN
	assert.Equal(suite.T(), "2020-02", MakeCalendarMonth(2020, time.January).NextMonth().String())
	assert.Equal(suite.T(), "2020-03", MakeCalendarMonth(2020, time.February).NextMonth().String())
	assert.Equal(suite.T(), "2020-04", MakeCalendarMonth(2020, time.March).NextMonth().String())
	assert.Equal(suite.T(), "2020-05", MakeCalendarMonth(2020, time.April).NextMonth().String())
	assert.Equal(suite.T(), "2020-06", MakeCalendarMonth(2020, time.May).NextMonth().String())
	assert.Equal(suite.T(), "2020-07", MakeCalendarMonth(2020, time.June).NextMonth().String())
	assert.Equal(suite.T(), "2020-08", MakeCalendarMonth(2020, time.July).NextMonth().String())
	assert.Equal(suite.T(), "2020-09", MakeCalendarMonth(2020, time.August).NextMonth().String())
	assert.Equal(suite.T(), "2020-10", MakeCalendarMonth(2020, time.September).NextMonth().String())
	assert.Equal(suite.T(), "2020-11", MakeCalendarMonth(2020, time.October).NextMonth().String())
	assert.Equal(suite.T(), "2020-12", MakeCalendarMonth(2020, time.November).NextMonth().String())
	assert.Equal(suite.T(), "2021-01", MakeCalendarMonth(2020, time.December).NextMonth().String())
}

func (suite *CalendarMonthTestSuite) Test_GIVEN_aCalendarMonth_WHEN_calculatingPreviousMonth_THEN_previousMonthIsCorrect() {
	// THEN
	assert.Equal(suite.T(), "2019-12", MakeCalendarMonth(2020, time.January).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-01", MakeCalendarMonth(2020, time.February).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-02", MakeCalendarMonth(2020, time.March).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-03", MakeCalendarMonth(2020, time.April).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-04", MakeCalendarMonth(2020, time.May).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-05", MakeCalendarMonth(2020, time.June).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-06", MakeCalendarMonth(2020, time.July).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-07", MakeCalendarMonth(2020, time.August).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-08", MakeCalendarMonth(2020, time.September).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-09", MakeCalendarMonth(2020, time.October).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-10", MakeCalendarMonth(2020, time.November).PreviousMonth().String())
	assert.Equal(suite.T(), "2020-11", MakeCalendarMonth(2020, time.December).PreviousMonth().String())
}
