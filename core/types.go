package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
)

type CalendarMonth struct {
	month time.Month
	year  uint
}

func MakeCalendarMonth(year uint, month time.Month) CalendarMonth {
	return CalendarMonth{
		year:  year,
		month: month,
	}
}

func MakeCalendarMonthFromDate(date time.Time) CalendarMonth {
	return MakeCalendarMonth(uint(date.Year()), date.Month())
}

func (cm CalendarMonth) Month() time.Month {
	return cm.month
}

func (cm CalendarMonth) Year() uint {
	return cm.year
}

func (cm CalendarMonth) FirstDay() time.Time {
	return time.Date(int(cm.year), cm.month, 1, 0, 0, 0, 0, time.UTC)
}

func (cm CalendarMonth) LastDay() time.Time {
	return cm.FirstDay().AddDate(0, 1, -1)
}

func (cm CalendarMonth) NextMonth() CalendarMonth {
	date := cm.FirstDay().AddDate(0, 1, 0)
	return MakeCalendarMonth(uint(date.Year()), date.Month())
}

func (cm CalendarMonth) PreviousMonth() CalendarMonth {
	date := cm.FirstDay().AddDate(0, -1, 0)
	return MakeCalendarMonth(uint(date.Year()), date.Month())
}

func (cm CalendarMonth) String() string {
	return fmt.Sprintf("%04d-%02d", cm.year, cm.month)
}

func makeCoreValidationError(code ErrorCode, errors *validate.Errors) error {
	if !errors.HasAny() {
		return nil
	}

	flatErrors := map[string]string{}
	for field, violations := range errors.Errors {
		flatErrors[field] = strings.Join(violations, ", ")
	}

	listErrors := []string{}
	for _, violations := range flatErrors {
		listErrors = append(listErrors, violations)
	}

	return NewErrorWithFields(code, strings.Join(listErrors, ", "), nil, flatErrors)
}
