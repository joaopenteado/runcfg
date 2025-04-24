package runcfg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Job contains environment variables and metadata available to Cloud Run jobs.
// For more details see:
// https://cloud.google.com/run/docs/container-contract#services-env-vars
type Job struct {
	// Metadata contains information from the instance metadata server.
	Metadata

	// Name is the name of the Cloud Run job being run.
	// Read from `CLOUD_RUN_JOB` environment variable.
	Name string

	// Execution is the name of the Cloud Run execution being run.
	// Read from `CLOUD_RUN_EXECUTION` environment variable.
	Execution string

	// TaskIndex is the index of this task.
	// Starts at 0 for the first task and increments by 1 for every
	// successive task, up to the maximum number of tasks minus 1.
	// If you set --parallelism to greater than 1, tasks might not
	// follow the index order. For example, it would be possible for
	// task 2 to start before task 1.
	// Read from `CLOUD_RUN_TASK_INDEX` environment variable.
	TaskIndex uint

	// TaskAttempt is the number of times this task has been retried.
	// Starts at 0 for the first attempt and increments by 1 for every
	// successive retry, up to the maximum retries value.
	// Read from `CLOUD_RUN_TASK_ATTEMPT` environment variable.
	TaskAttempt uint

	// TaskCount is the total number of tasks defined in the --tasks parameter.
	// Read from `CLOUD_RUN_TASK_COUNT` environment variable.
	TaskCount uint
}

// LoadJob loads configuration for a Cloud Run job, including both environment
// variables and metadata from the metadata server. It returns a Job containing
// the loaded configuration or an error if the loading process fails.
//
// The ctx parameter is used for metadata server requests. The opts parameter
// allows specifying options to configure the loading process.
//
// LoadJob will return ErrEnvironmentProcess if environment variable processing
// fails, or ErrMetadataFetch if metadata server requests fail.
func LoadJob(ctx context.Context, opts ...LoadOption) (*Job, error) {
	cfg := Job{
		Name:      os.Getenv("CLOUD_RUN_JOB"),
		Execution: os.Getenv("CLOUD_RUN_EXECUTION"),
	}

	if taskIdx := os.Getenv("CLOUD_RUN_TASK_INDEX"); taskIdx != "" {
		idx, err := strconv.ParseUint(taskIdx, 10, 32)
		if err != nil {
			return nil, errors.Join(ErrEnvironmentProcess, fmt.Errorf("invalid CLOUD_RUN_TASK_INDEX value: %w", err))
		}
		cfg.TaskIndex = uint(idx)
	}

	if taskAttempt := os.Getenv("CLOUD_RUN_TASK_ATTEMPT"); taskAttempt != "" {
		attempt, err := strconv.ParseUint(taskAttempt, 10, 32)
		if err != nil {
			return nil, errors.Join(ErrEnvironmentProcess, fmt.Errorf("invalid CLOUD_RUN_TASK_ATTEMPT value: %w", err))
		}
		cfg.TaskAttempt = uint(attempt)
	}

	if taskCount := os.Getenv("CLOUD_RUN_TASK_COUNT"); taskCount != "" {
		count, err := strconv.ParseUint(taskCount, 10, 32)
		if err != nil {
			return nil, errors.Join(ErrEnvironmentProcess, fmt.Errorf("invalid CLOUD_RUN_TASK_COUNT value: %w", err))
		}
		cfg.TaskCount = uint(count)
	}

	loadOpts := defaultOptions()
	for _, opt := range opts {
		opt(&loadOpts)
	}

	metadata, err := LoadMetadata(ctx, loadOpts.requiredMetadata)
	if err != nil {
		return nil, errors.Join(ErrMetadataFetch, err)
	}

	cfg.Metadata = *metadata
	return &cfg, nil
}
