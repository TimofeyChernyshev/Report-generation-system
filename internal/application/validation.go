package application

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

var (
	reTimeHHMM = regexp.MustCompile(`^\d{2}:\d{2}$`)
)

var (
	errEmptyID         error = errors.New("empty ID")
	errWrongTimeFormat error = errors.New("wrong time format")
	errExitBeforeCome  error = errors.New("exiting time before coming time")
)

// Проверяет корректность данных о сотрудниках и объединяет разрозненные записи в одну
func validateAndNormalizeEmplData(data []domain.EmplRawData) ([]domain.EmplRawData, []error) {
	var errs []error
	normalized := make(map[string]domain.EmplRawData)

	for i, e := range data {
		// Проверка ID
		if e.ID == "" {
			errs = append(errs, fmt.Errorf("record %d: %w", i, errEmptyID))
		}

		// Проверка времени
		if e.ComingTime != "" && !isValidTime(e.ComingTime) {
			errs = append(errs, fmt.Errorf("record %d: %w: coming time '%s'", i, errWrongTimeFormat, e.ComingTime))
		}
		if e.ExitingTime != "" && !isValidTime(e.ExitingTime) {
			errs = append(errs, fmt.Errorf("record %d: %w: exiting time '%s'", i, errWrongTimeFormat, e.ExitingTime))
		}
		come, _ := time.Parse("15:04", e.ComingTime)
		exit, _ := time.Parse("15:04", e.ExitingTime)
		if e.ComingTime != "" && e.ExitingTime != "" && come.After(exit) {
			errs = append(errs, fmt.Errorf("record: %d: %w", i, errExitBeforeCome))
		}

		// Проверка требуемого рабочего диапазона
		if e.WorkingTime != "" {
			times := regexp.MustCompile(`^(\d{2}:\d{2})-(\d{2}:\d{2})$`).FindStringSubmatch(e.WorkingTime)
			if len(times) != 3 || !isValidTime(times[1]) || !isValidTime(times[2]) {
				errs = append(errs, fmt.Errorf("record %d: %w(HH:MM-HH:MM): working time '%s'", i, errWrongTimeFormat, e.WorkingTime))
			}
		}

		if existing, ok := normalized[e.ID]; ok {
			merged := mergeEmplRecords(existing, e, &errs, i)
			normalized[e.ID] = merged
		} else {
			normalized[e.ID] = e
		}
	}

	var result []domain.EmplRawData
	for _, v := range normalized {
		result = append(result, v)
	}

	return result, errs
}

// mergeEmplRecords объединяет две записи одного сотрудника
func mergeEmplRecords(a, b domain.EmplRawData, errs *[]error, index int) domain.EmplRawData {
	if a.Name == "" && b.Name != "" {
		a.Name = b.Name
	} else if a.Name != "" && b.Name != "" && a.Name != b.Name {
		*errs = append(*errs, fmt.Errorf("name conflict with ID %s (record %d): '%s' vs '%s'", a.ID, index, a.Name, b.Name))
	}

	if a.Email == "" {
		a.Email = b.Email
	}
	if a.PhoneNum == "" {
		a.PhoneNum = b.PhoneNum
	}
	if a.WorkingTime == "" {
		a.WorkingTime = b.WorkingTime
	}

	// если в одной записи не было времени прихода/ухода, подставляем из другой
	// берем самое раннее время прихода для дня и самое позднее время ухода
	if a.ComingTime == "" {
		a.ComingTime = b.ComingTime
	} else if a.ComingTime != "" && b.ComingTime != "" {
		aComingTime, _ := time.Parse("15:04", a.ComingTime)
		bComingTime, _ := time.Parse("15:04", b.ComingTime)
		if aComingTime.After(bComingTime) {
			a.ComingTime = b.ComingTime
		}
	}

	if a.ExitingTime == "" {
		a.ExitingTime = b.ExitingTime
	} else if a.ExitingTime != "" && b.ExitingTime != "" {
		aExitingTime, _ := time.Parse("15:04", a.ExitingTime)
		bExitingTime, _ := time.Parse("15:04", b.ExitingTime)
		if aExitingTime.Before(bExitingTime) {
			a.ExitingTime = b.ExitingTime
		}
	}

	return a
}

func isValidTime(value string) bool {
	if !reTimeHHMM.MatchString(value) {
		return false
	}
	t, err := time.Parse("15:04", value)
	if err != nil {
		return false
	}
	h, m := t.Hour(), t.Minute()
	return h >= 0 && h < 24 && m >= 0 && m < 60
}
