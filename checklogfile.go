package checklogfile

import (
	"bufio"
	"io"
	"regexp"
)

// accept and io stream, we can read from, seek into and close
type ReadSeekCloser interface {
	io.Reader
	io.Closer
	io.Seeker
}

// Result to be reported to a monitoring system
type MonitoringResult int

const (
	MonitorOk MonitoringResult = iota // No remaining issues
	MonitorWarning
	MonitorCritical
	MonitorUnknown // Cannot determine state
	MonitorCount // Upper bound for arrays based on levels
)

var monitoringResults = [MonitorCount]string{
	MonitorOk:       "OK",
	MonitorCritical: "CRITICAL",
	MonitorWarning:  "WARNING",
	MonitorUnknown:  "UNKNOWN",
}

func (m MonitoringResult) String() string {
	return monitoringResults[m]
}

// Abstracts away a logfile event classificator
type Logfile struct {
	patterns [MonitorCount][]*regexp.Regexp
	r        ReadSeekCloser
	buffer   *bufio.Reader
	offset   int64
}

// Start managing logfile r at start. If start is invalid, we seek to the current end.
func NewLogFile(r ReadSeekCloser, start int64) *Logfile {
	l := &Logfile{
		r:      r,
		offset: start,
		buffer: bufio.NewReader(r),
	}
	_, err := l.r.Seek(start, 0)
	// hmm, maybe there has been a log rotation?
	if err != nil {
		l.offset, err = l.r.Seek(0, 0)
		if err != nil {
			panic(err)
		}
	}
	return l
}

// Add a regexp pattern to trigger monitoring alert level.
// Logfile lines are matched in order of apperance,
// but which pattern is applied first is not specified.
func (l *Logfile) AddPattern(level MonitoringResult, pattern string) error {
	if level < 0 || level > MonitorCount {
		panic("invalid monitor level")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	l.patterns[level] = append(l.patterns[level], re)
	return nil
}

// Scan logfile until end for supplied patterns.
// In res you can find the monitoring level to be reported,
// counts contains useful statistics and err is any error that accours.
// io.EOF is not reported as error, since this function is supposed to scan
// right until io.EOF in the good case.
// Empty files generate MonitorOk, unreadable files MonitorUnknown,
// everything else depends on which pattern the last read line matches.
func (l *Logfile) Scan() (res MonitoringResult, counts [MonitorCount]int64, err error) {
	var line []byte
	var matched, read int64
	res = MonitorUnknown
	for line, err = l.buffer.ReadBytes('\n'); err == nil; line, err = l.buffer.ReadBytes('\n') {
		for i := range l.patterns {
			for _, pattern := range l.patterns[i] {
				if pattern.Match(line) {
					matched += 1
					counts[i] += 1
					res = MonitoringResult(i)
					break
				}
			}
		}
		read += 1
	}
	if err == io.EOF {
		err = nil
	}
	if err != nil || read == 0 {
		res = MonitorUnknown
	}
	if read > 0 && matched == 0 {
		res = MonitorOk
	}
	if offset, err := l.r.Seek(0, 1); err == nil {
		offset -= int64(l.buffer.Buffered())
		l.offset = offset
	}
	return
}

// Returns current logfile offset. We will start here for the next Scan.
func (l *Logfile) Offset() int64 { return l.offset }

// Properly close underlying streams.
func (l *Logfile) Close() (err error) {
	err = l.r.Close()
	l.buffer = nil
	return
}
