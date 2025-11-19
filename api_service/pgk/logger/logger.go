package logger

import "log"

func Info(message string) {
	log.Printf("[INFO] %s", message)
}

func Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func Error(err error, message string) {
	log.Printf("[ERROR] %s: %v", message, err)
}

func Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func Warn(message string) {
	log.Printf("[WARN] %s", message)
}

func Warnf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

func Debug(message string) {
	log.Printf("[DEBUG] %s", message)
}

func Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

func WorkerInfo(workerID int, message string) {
	log.Printf("[INFO] Worker %d: %s", workerID, message)
}

func WorkerInfof(workerID int, format string, args ...interface{}) {
	allArgs := append([]interface{}{workerID}, args...)
	log.Printf("[INFO] Worker %d: "+format, allArgs...)
}

func WorkerError(workerID int, err error, message string) {
	log.Printf("[ERROR] Worker %d: %s: %v", workerID, message, err)
}

func WorkerErrorf(workerID int, format string, args ...interface{}) {
	allArgs := append([]interface{}{workerID}, args...)
	log.Printf("[ERROR] Worker %d: "+format, allArgs...)
}
