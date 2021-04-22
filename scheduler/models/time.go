package models

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseTime(timeStr string) (int, int, int, error) {
	parts := strings.Split(timeStr, "-")

	if len(parts) != 2 && len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("failed to parse time, possible wrong separator")
	}

	var (
		hour, minute, second int64
		err                  error
	)

	hour, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse hours: %w", err)
	}
	if hour > 24 || hour < 0 {
		return 0, 0, 0, fmt.Errorf("wrong hour value %v, should be => 0 and <= 24", hour)
	}

	minute, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse minutes: %w", err)
	}
	if minute > 60 || minute < 0 {
		return 0, 0, 0, fmt.Errorf("wrong minute value %v, should be => 0 and <= 60", minute)
	}

	if len(parts) == 3 {
		second, err = strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("failed to parse minutes: %w", err)
		}
		if second > 60 || second < 0 {
			return 0, 0, 0, fmt.Errorf("wrong second value %v, should be => 0 and <= 60", second)
		}
	}

	return int(hour), int(minute), int(second), nil
}
