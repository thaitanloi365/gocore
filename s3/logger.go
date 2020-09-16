package s3

// Logger interface
type Logger interface {
	Printf(format string, v ...interface{})
}
