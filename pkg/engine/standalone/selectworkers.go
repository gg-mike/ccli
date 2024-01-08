package standalone

import (
	"sort"

	"github.com/gg-mike/ccli/pkg/model"
)

func SelectWorker(cfg model.PipelineConfig, workers []model.Worker) (model.Worker, error) {
	if len(workers) == 0 {
		return model.Worker{}, ErrNoAvailableWorker
	}

	workers = filterWorkers(workers, cfg.System, cfg.Image)
	if len(workers) == 0 {
		return model.Worker{}, ErrNoAvailableWorkerForConfiguration
	}

	return sortWorkers(workers)[0], nil
}

func filterWorkers(workers []model.Worker, system, image string) []model.Worker {
	filteredWorkers := []model.Worker{}
	for _, worker := range workers {
		if worker.ActiveBuilds < worker.Capacity &&
			(system == "" || (system != "" && worker.System == system)) &&
			(image == "" || (image != "" && !worker.IsStatic)) {
			filteredWorkers = append(filteredWorkers, worker)
		}
	}
	return filteredWorkers
}

// First element after sorting should be the best candidate
func sortWorkers(workers []model.Worker) []model.Worker {
	sort.Slice(workers, func(i, j int) bool {
		iw, jw := workers[i], workers[j]
		switch {
		case iw.Strategy != jw.Strategy:
			return iw.Strategy > jw.Strategy
		case iw.ActiveBuilds != jw.ActiveBuilds:
			return iw.ActiveBuilds < jw.ActiveBuilds
		default:
			return iw.Name < jw.Name
		}
	})
	return workers
}
