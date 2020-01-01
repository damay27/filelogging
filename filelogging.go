package filelogging

import (
	"os"
	"sync"
)

//LogSeverityLevel is used to specify the severity of a message in a log file
type LogSeverityLevel uint8

const (
	//STATUS has a integer value of zero
	STATUS = iota
	//WARNING has an integer value of 1
	WARNING
	//ERROR has an integer value of 2
	ERROR
)

//LogFile structure for working with log files.
//Must be opened before used and closed when no longer needed
type LogFile struct {
	file  *os.File
	mutex sync.Mutex //Is already initialized to unlocked state
}

//Error for internal use only. Used when the number of bytes written does not
//match the number of bytes sent to the file.
type logWriteError struct{}

func (lwe logWriteError) Error() string {
	return "The number of bytes written to the file does not match the message length"
}

//OpenLogFile will open a file at the given path for writing logs into. If the
//file does not already exist if will be created. The fule is opened such that
//all data being sent to it is appended to the file.
func (lf *LogFile) OpenLogFile(filePath string) error {
	var err error
	lf.file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	return err
}

//Log will write the given message to the log with the given severity level.
func (lf *LogFile) Log(message string, logLevel LogSeverityLevel) error {

	var count int
	var err error

	//Prepend the correct text to the method based on the severity level
	switch logLevel {
	case WARNING:
		message = "WARNING: " + message
	case ERROR:
		message = "ERROR: " + message
	}

	message += "\n"

	lf.mutex.Lock()
	count, err = lf.file.Write([]byte(message))

	//Error checking for the file write
	if err != nil {
		lf.mutex.Unlock()
		return err
	} else if count < len(message) {
		lf.mutex.Unlock()
		var countErr logWriteError
		return countErr
	}

	//Always sync the file so that no data will be lost in the event of a
	//program crash
	err = lf.file.Sync()
	lf.mutex.Unlock()
	if err != nil {
		return err
	}

	return nil
}

//CloseLogFile waits to log the file and the closes it. Any attept to write a log
//to the file after this function is called will result in an error.
func (lf *LogFile) CloseLogFile() error {
	lf.mutex.Lock()
	err := lf.file.Close()
	lf.mutex.Unlock()
	return err
}
