package logger

// Log is the system logging interface
type Log interface {
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
}

//Wrap places a new, or replaces a prefix on a logger object
func Wrap(prefix string, logger Log) Log {
	switch logger.(type) {
	case logPrefixer:
		logger.Info("pulling log from logPrefixer, new prefix: %s", prefix)
		//construct new, replacing prefix
		lp := logger.(logPrefixer)
		return logPrefixer{prefix, lp.log}
	default:
		logger.Info("Wrapping log interface, new prefix: %s", prefix)
		//wrap logger
		return logPrefixer{prefix, logger}
	}
}

type logPrefixer struct {
	prefix string
	log    Log
}

func (lp logPrefixer) Debug(format string, v ...interface{}) {
	lp.log.Debug(lp.prefix+" "+format, v)
}

func (lp logPrefixer) Error(format string, v ...interface{}) {
	lp.log.Error(lp.prefix+" "+format, v)
}

func (lp logPrefixer) Fatal(format string, v ...interface{}) {
	lp.log.Fatal(lp.prefix+" "+format, v)
}

func (lp logPrefixer) Info(format string, v ...interface{}) {
	lp.log.Info(lp.prefix+" "+format, v)
}

func (lp logPrefixer) Trace(format string, v ...interface{}) {
	lp.log.Trace(lp.prefix+" "+format, v)
}

func (lp logPrefixer) Warn(format string, v ...interface{}) {
	lp.log.Warn(lp.prefix+" "+format, v)
}
