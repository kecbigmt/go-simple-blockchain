/*
Package logch focuses on controlling multiple logs between goroutines. It provides functionality for unifying several channels corresponding log level, inputting logs into a specified channel and taking out them.
*/
package logch

import (
  "fmt"
  //"context"
)

// LogCh unifies channels corresponding 4 log levels(INFO, DEBUG, WARNING, ERROR). These channels are private fields. Logs can only be input by using Infof, Debugf, Warningf and Errorf methods. And they can only be taken out by InfoOut, DebugOut, WarningOut and ErrorOut methods.
type LogCh struct {
  info chan string
  debug chan string
  warning chan string
  error chan error
}

// NewLogCh returns the pointer for a initialized LogCh.
func NewLogCh() *LogCh {
  return &LogCh{
    info: make(chan string, 1),
    debug: make(chan string, 1),
    warning: make(chan string, 1),
    error: make(chan error, 1),
  }
}

// Infof inputs formatted text into LogCh for INFO level logging
func (ch *LogCh) Infof(format string, a ...interface{}) {
  log := fmt.Sprintf(format, a...)
  ch.info <- log
}

// Debugf inputs formatted text into LogCh for DEBUG level logging
func (ch *LogCh) Debugf(format string, a ...interface{}) {
  log := fmt.Sprintf(format, a...)
  ch.debug <- log
}

// Warningf inputs formatted text into LogCh for WARNING level logging
func (ch *LogCh) Warningf(format string, a ...interface{}) {
  log := fmt.Sprintf(format, a...)
  ch.warning <- log
}

// Errorf inputs formatted error into LogCh for ERROR level logging
func (ch *LogCh) Errorf(format string, a ...interface{}) {
  err := fmt.Errorf(format, a...)
  ch.error <- err
}


// InfoOut returns read-only channel to take out INFO level logs
func (ch *LogCh) InfoOut() <-chan string {
  return ch.info
}

// DebugOut returns read-only channel to take out DEBUG level logs
func (ch *LogCh) DebugOut() <-chan string {
  return ch.debug
}

// WarningOut returns read-only channel to take out WARNING level logs
func (ch *LogCh) WarningOut() <-chan string {
  return ch.warning
}

// ErrorOut returns read-only channel to take out ERROR level logs
func (ch *LogCh) ErrorOut() <-chan error {
  return ch.error
}

// InfoClose closes the INFO channel in LogCh
func (ch *LogCh) InfoClose() {
  close(ch.info)
  return
}

// InfoClose closes the DEBUG channel in LogCh
func (ch *LogCh) DebugClose() {
  close(ch.debug)
  return
}

// InfoClose closes the WARNING channel in LogCh
func (ch *LogCh) WarningClose() {
  close(ch.warning)
  return
}

// InfoClose closes the ERROR channel in LogCh
func (ch *LogCh) ErrorClose(){
  close(ch.error)
  return
}

// AllClose closes all channels in LogCh
func (ch *LogCh) AllClose(){
  ch.InfoClose()
  ch.DebugClose()
  ch.WarningClose()
  ch.ErrorClose()
  return
}
