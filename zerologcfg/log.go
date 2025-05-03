package zerologcfg

import (
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	// This init function modifies global zerolog settings to align with Google
	// Cloud Logging conventions.
	// These changes affect all logging behavior globally in the application,
	// which may have side effects on downstream logging if other parts of the
	// application or libraries also use zerolog.
	//
	// Specifically:
	// - TimeFieldFormat is set to RFC3339Nano for precise timestamp formatting.
	// - LevelFieldName is set to "severity" to match Google Cloud Logging's
	//   expected field name.
	// - LevelFieldMarshalFunc is customized to map zerolog levels to Google
	//   Cloud Logging severity levels.
	//
	// For more details, refer to:
	// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.LevelFieldName = "severity"
	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		switch l {
		case zerolog.TraceLevel:
			return "DEFAULT"
		case zerolog.DebugLevel:
			return "DEBUG"
		case zerolog.InfoLevel:
			return "INFO"
		case zerolog.WarnLevel:
			return "WARNING"
		case zerolog.ErrorLevel:
			return "ERROR"
		case zerolog.FatalLevel:
			return "CRITICAL"
		case zerolog.PanicLevel:
			return "ALERT"
		case zerolog.NoLevel:
			return "DEFAULT"
		default:
			return "DEFAULT"
		}
	}
}

type cloudLoggingHook struct {
	ProjectID string
}

func Hook(projectID string) zerolog.Hook {
	return &cloudLoggingHook{ProjectID: projectID}
}

func (h cloudLoggingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	// https://cloud.google.com/logging/docs/structured-logging

	// CallerSkipFrameCount is the number of stack frames to skip to find the
	// caller. The first frame is the CloudLoggingHook, the second is the
	// zerolog.Logger, the third is the caller.
	const skipFrameCount = 3

	var file, line, function string
	if pc, filePath, lineNum, ok := runtime.Caller(skipFrameCount); ok {
		if f := runtime.FuncForPC(pc); f != nil {
			function = f.Name()
		}
		line = strconv.Itoa(lineNum)
		file = filePath[strings.LastIndexByte(filePath, '/')+1:]
	}
	e.Dict("logging.googleapis.com/sourceLocation", zerolog.Dict().
		Str("file", file).
		Str("line", line).
		Str("function", function))

	ctx := e.GetCtx()
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		e.Str("logging.googleapis.com/trace", "projects/"+h.ProjectID+"/traces/"+span.TraceID().String())
		e.Str("logging.googleapis.com/spanId", span.SpanID().String())
		e.Bool("logging.googleapis.com/trace_sampled", span.TraceFlags().IsSampled())
	}
}
