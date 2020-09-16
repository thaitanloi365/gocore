package sendgrid

// Logger interface
type Logger interface {
	Printf(format string, v ...interface{})
}
