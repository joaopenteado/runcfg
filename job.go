package runcfg

import (
	"context"
	"errors"
	"os"
	"strconv"
)

// Job contains environment variables available to Cloud Run jobs.
// For more details see the [container runtime contract for Jobs].
//
// [container runtime contract for Jobs]: https://cloud.google.com/run/docs/container-contract#jobs-env-vars
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

func defaultJob() *Job {
	return &Job{
		TaskCount: 1,
	}
}

type JobLoadOption func(*Job)

// WithDefaultJobName specifies the default name to use if the CLOUD_RUN_JOB
// environment variable is not set. If multiple names are provided, the first
// non-empty name will be used.
func WithDefaultJobName(names ...string) JobLoadOption {
	return func(o *Job) {
		for _, name := range names {
			if name != "" {
				o.Name = name
				break
			}
		}
	}
}

// WithDefaultExecution specifies the default execution name to use if the
// CLOUD_RUN_EXECUTION environment variable is not set. If multiple execution
// names are provided, the first non-empty execution name will be used.
func WithDefaultExecution(executions ...string) JobLoadOption {
	return func(o *Job) {
		for _, execution := range executions {
			if execution != "" {
				o.Execution = execution
				break
			}
		}
	}
}

// WithDefaultTaskIndex specifies the default task index to use if the
// CLOUD_RUN_TASK_INDEX environment variable is not set.
func WithDefaultTaskIndex(taskIndex uint) JobLoadOption {
	return func(o *Job) {
		o.TaskIndex = taskIndex
	}
}

// WithDefaultTaskAttempt specifies the default task attempt to use if the
// CLOUD_RUN_TASK_ATTEMPT environment variable is not set.
func WithDefaultTaskAttempt(taskAttempt uint) JobLoadOption {
	return func(o *Job) {
		o.TaskAttempt = taskAttempt
	}
}

// WithDefaultTaskCount specifies the default task count to use if the
// CLOUD_RUN_TASK_COUNT environment variable is not set.
func WithDefaultTaskCount(taskCount uint) JobLoadOption {
	return func(o *Job) {
		o.TaskCount = taskCount
	}
}

// LoadJob loads configuration for a Cloud Run job from environment variables.
// It returns a Job containing the loaded configuration or ErrEnvironmentProcess
// if environment variable processing fails. Use options to specify default
// values for the job.
func LoadJob(opts ...JobLoadOption) (*Job, error) {
	// Default values
	j := defaultJob()

	// Apply options
	for _, opt := range opts {
		opt(j)
	}

	// Reload configuration from the environment
	if err := j.Reload(); err != nil {
		return nil, err
	}

	return j, nil
}

// Reload reloads the configuration for a Cloud Run job from environment
// variables. It returns ErrEnvironmentProcess if environment variable
// processing fails. It does not overwrite values already set in the Job
// struct if they are not set in the environment.
func (j *Job) Reload() error {
	if name := os.Getenv("CLOUD_RUN_JOB"); name != "" {
		j.Name = name
	}

	if execution := os.Getenv("CLOUD_RUN_EXECUTION"); execution != "" {
		j.Execution = execution
	}

	if taskIdx := os.Getenv("CLOUD_RUN_TASK_INDEX"); taskIdx != "" {
		idx, err := strconv.ParseUint(taskIdx, 10, 32)
		if err != nil {
			return errors.Join(ErrEnvironmentProcess, errors.New("invalid CLOUD_RUN_TASK_INDEX value"), err)
		}
		j.TaskIndex = uint(idx)
	}

	if taskAttempt := os.Getenv("CLOUD_RUN_TASK_ATTEMPT"); taskAttempt != "" {
		attempt, err := strconv.ParseUint(taskAttempt, 10, 32)
		if err != nil {
			return errors.Join(ErrEnvironmentProcess, errors.New("invalid CLOUD_RUN_TASK_ATTEMPT value"), err)
		}
		j.TaskAttempt = uint(attempt)
	}

	if taskCount := os.Getenv("CLOUD_RUN_TASK_COUNT"); taskCount != "" {
		count, err := strconv.ParseUint(taskCount, 10, 32)
		if err != nil {
			return errors.Join(ErrEnvironmentProcess, errors.New("invalid CLOUD_RUN_TASK_COUNT value"), err)
		}
		j.TaskCount = uint(count)
	}

	return nil
}

// EnvDecode implements the [envconfig.DecoderCtx] interface from
// github.com/sethvargo/go-envconfig. This ensures that [envconfig.Process] will
// return errors derived from [ErrEnvironmentProcess] if the environment
// variables are invalid.
//
// The behavior of this function is equivalent to initializing a Job with
// the default values and then calling [Job.Reload]. However, values already
// set in the Job struct prior to calling this function are not overridden
// by the defaults, only by the reloaded values from the environment.
//
// [envconfig.DecoderCtx]: https://pkg.go.dev/github.com/sethvargo/go-envconfig#DecoderCtx
// [envconfig.Process]: https://pkg.go.dev/github.com/sethvargo/go-envconfig#Process
func (j *Job) EnvDecode(ctx context.Context, val string) error {
	defaults := defaultJob()

	// Name, Execution, TaskIndex and TaskAttempt default values are zero and
	// can be skipped.

	if j.TaskCount == 0 {
		j.TaskCount = defaults.TaskCount
	}

	return j.Reload()
}
