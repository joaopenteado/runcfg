package runcfg

import (
	"errors"
	"os"
	"strconv"
)

type jobLoadOptions struct {
	defaultName        string
	defaultExecution   string
	defaultTaskIndex   uint
	defaultTaskAttempt uint
	defaultTaskCount   uint
}

func defaultJobLoadOptions() jobLoadOptions {
	return jobLoadOptions{
		defaultName:        "",
		defaultExecution:   "",
		defaultTaskIndex:   0,
		defaultTaskAttempt: 0,
		defaultTaskCount:   1,
	}
}

type JobLoadOption func(*jobLoadOptions)

// WithDefaultJobName specifies the default name to use if the CLOUD_RUN_JOB
// environment variable is not set. If multiple names are provided, the first
// non-empty name will be used.
func WithDefaultJobName(names ...string) JobLoadOption {
	return func(o *jobLoadOptions) {
		for _, name := range names {
			if name != "" {
				o.defaultName = name
				break
			}
		}
	}
}

// WithDefaultExecution specifies the default execution name to use if the
// CLOUD_RUN_EXECUTION environment variable is not set. If multiple execution
// names are provided, the first non-empty execution name will be used.
func WithDefaultExecution(executions ...string) JobLoadOption {
	return func(o *jobLoadOptions) {
		for _, execution := range executions {
			if execution != "" {
				o.defaultExecution = execution
				break
			}
		}
	}
}

// WithDefaultTaskIndex specifies the default task index to use if the
// CLOUD_RUN_TASK_INDEX environment variable is not set.
func WithDefaultTaskIndex(taskIndex uint) JobLoadOption {
	return func(o *jobLoadOptions) {
		o.defaultTaskIndex = taskIndex
	}
}

// WithDefaultTaskAttempt specifies the default task attempt to use if the
// CLOUD_RUN_TASK_ATTEMPT environment variable is not set.
func WithDefaultTaskAttempt(taskAttempt uint) JobLoadOption {
	return func(o *jobLoadOptions) {
		o.defaultTaskAttempt = taskAttempt
	}
}

// WithDefaultTaskCount specifies the default task count to use if the
// CLOUD_RUN_TASK_COUNT environment variable is not set.
func WithDefaultTaskCount(taskCount uint) JobLoadOption {
	return func(o *jobLoadOptions) {
		o.defaultTaskCount = taskCount
	}
}

// Job contains environment variables available to Cloud Run jobs.
// For more details see:
// https://cloud.google.com/run/docs/container-contract#services-env-vars
type Job struct {
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

// LoadJob loads configuration for a Cloud Run job from environment variables.
// It returns a Job containing the loaded configuration or ErrEnvironmentProcess
// if environment variable processing fails.
func LoadJob(opts ...JobLoadOption) (*Job, error) {
	loadOpts := defaultJobLoadOptions()
	for _, opt := range opts {
		opt(&loadOpts)
	}

	cfg := Job{
		Name:      os.Getenv("CLOUD_RUN_JOB"),
		Execution: os.Getenv("CLOUD_RUN_EXECUTION"),
	}

	if cfg.Name == "" {
		cfg.Name = loadOpts.defaultName
	}
	if cfg.Execution == "" {
		cfg.Execution = loadOpts.defaultExecution
	}

	if taskIdx := os.Getenv("CLOUD_RUN_TASK_INDEX"); taskIdx != "" {
		idx, err := strconv.ParseUint(taskIdx, 10, 32)
		if err != nil {
			return nil, errors.Join(ErrEnvironmentProcess, errors.New("invalid CLOUD_RUN_TASK_INDEX value"), err)
		}
		cfg.TaskIndex = uint(idx)
	} else {
		cfg.TaskIndex = loadOpts.defaultTaskIndex
	}

	if taskAttempt := os.Getenv("CLOUD_RUN_TASK_ATTEMPT"); taskAttempt != "" {
		attempt, err := strconv.ParseUint(taskAttempt, 10, 32)
		if err != nil {
			return nil, errors.Join(ErrEnvironmentProcess, errors.New("invalid CLOUD_RUN_TASK_ATTEMPT value"), err)
		}
		cfg.TaskAttempt = uint(attempt)
	} else {
		cfg.TaskAttempt = loadOpts.defaultTaskAttempt
	}

	if taskCount := os.Getenv("CLOUD_RUN_TASK_COUNT"); taskCount != "" {
		count, err := strconv.ParseUint(taskCount, 10, 32)
		if err != nil {
			return nil, errors.Join(ErrEnvironmentProcess, errors.New("invalid CLOUD_RUN_TASK_COUNT value"), err)
		}
		cfg.TaskCount = uint(count)
	} else {
		cfg.TaskCount = loadOpts.defaultTaskCount
	}

	return &cfg, nil
}
