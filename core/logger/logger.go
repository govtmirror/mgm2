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
		//construct new, replacing prefix
		lp := logger.(logPrefixer)
		return logPrefixer{lp.prefix + "." + prefix, lp.log}
	default:
		//wrap logger
		return logPrefixer{prefix, logger}
	}
}

type logPrefixer struct {
	prefix string
	log    Log
}

func (lp logPrefixer) Debug(format string, v ...interface{}) {
	var args []interface{}
	args = append(args, lp.prefix)
	args = append(args, v...)
	lp.log.Debug("[%v] "+format, args...)
}

func (lp logPrefixer) Error(format string, v ...interface{}) {
	var args []interface{}
	args = append(args, lp.prefix)
	args = append(args, v...)
	lp.log.Error("[%v] "+format, args...)
}

func (lp logPrefixer) Fatal(format string, v ...interface{}) {
	var args []interface{}
	args = append(args, lp.prefix)
	args = append(args, v...)
	lp.log.Fatal("[%v] "+format, args...)
}

func (lp logPrefixer) Info(format string, v ...interface{}) {
	var args []interface{}
	args = append(args, lp.prefix)
	args = append(args, v...)
	lp.log.Info("[%v] "+format, args...)
}

func (lp logPrefixer) Trace(format string, v ...interface{}) {
	var args []interface{}
	args = append(args, lp.prefix)
	args = append(args, v...)
	lp.log.Trace("[%v] "+format, args...)
}

func (lp logPrefixer) Warn(format string, v ...interface{}) {
	var args []interface{}
	args = append(args, lp.prefix)
	args = append(args, v...)
	lp.log.Warn("[%v] "+format, args...)
}
