package domain

import "time"

func NewEmplCompleteData(e EmplRawData) EmplCompleteData {
	return EmplCompleteData{
		ID:            e.ID,
		Name:          e.Name,
		Email:         e.Email,
		PhoneNum:      e.PhoneNum,
		WorkedTime:    0,
		LateComeTime:  0,
		EarlyExitTime: 0,
	}
}

func (e EmplRawData) CalculateWorkedTime() float64 {
	if e.ComingTime == "" && e.ExitingTime == "" {
		return 0
	}
	comingTime, _ := time.Parse("15:04", e.ComingTime)
	exitingTime, _ := time.Parse("15:04", e.ExitingTime)
	workedTimeH := exitingTime.Hour() - comingTime.Hour()
	workedTimeM := exitingTime.Minute() - comingTime.Minute()
	return float64(workedTimeH) + (float64(workedTimeM) / 60.0)
}

func (e EmplRawData) CalculateLateComeTime() float64 {
	expectedComingTime, _ := time.Parse("15:04", e.WorkingTime[:5])
	var actualComingTime time.Time
	if e.ComingTime == "" {
		actualComingTime, _ = time.Parse("15:04", e.WorkingTime[6:])
	} else {
		actualComingTime, _ = time.Parse("15:04", e.ComingTime)
	}

	if actualComingTime.After(expectedComingTime) {
		lateH := actualComingTime.Hour() - expectedComingTime.Hour()
		lateM := actualComingTime.Minute() - expectedComingTime.Minute()
		return float64(lateH) + (float64(lateM) / 60.0)
	}
	return 0
}

func (e EmplRawData) CalculateEarlyExitTime() float64 {
	expectedExitingTime, _ := time.Parse("15:04", e.WorkingTime[6:])
	var actualExitingTime time.Time
	if e.ExitingTime == "" {
		actualExitingTime, _ = time.Parse("15:04", e.WorkingTime[:5])
	} else {
		actualExitingTime, _ = time.Parse("15:04", e.ExitingTime)
	}

	if actualExitingTime.Before(expectedExitingTime) {
		earlyH := expectedExitingTime.Hour() - actualExitingTime.Hour()
		earlyM := expectedExitingTime.Minute() - actualExitingTime.Minute()
		return float64(earlyH) + (float64(earlyM) / 60.0)
	}
	return 0
}
