package standalone

import "errors"

var (
	ErrNoAvailableWorker                 = errors.New("no worker is available")
	ErrNoAvailableWorkerForConfiguration = errors.New("no worker is available for given configuration")

	ErrUpdatingWorker = errors.New("error during worker update")
)
