//go:generate stringer -type=TaskStatus
package taskmodel

import "errors"

type TaskStatus int

const (
	NEW TaskStatus = iota
	WAITING
	SUBMITTED
	PROCESSING
	DONE
	ERROR
	FAILED
)

func TaskStatusFromString(status string) (TaskStatus, error) {
	switch status {
	case "NEW":
		return NEW, nil
	case "WAITING":
		return WAITING, nil
	case "SUBMITTED":
		return SUBMITTED, nil
	case "PROCESSING":
		return PROCESSING, nil
	case "DONE":
		return DONE, nil
	case "ERROR":
		return ERROR, nil
	case "FAILED":
		return FAILED, nil
	default:
		return -1, errors.New("unknown task status")
	}
}
