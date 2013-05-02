package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/laziac/go-nagios/nagios"
	"github.com/nightlyone/checklogfile"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Verbose         bool     `short:"v" long:"verbose" description:"be verbose debug"`
	Logfile         string   `short:"f" long:"logfile" required:"true" description:"parse this logfile"`
	Tag             string   `short:"t" long:"tag" default:"default" description:"tag to use for reporting"`
	OffsetFile      string   `short:"o" long:"offsetfile" description:"file describing the offset we restart scaning for events"`
	OkPattern       []string `short:"k" long:"okpattern" description:"pattern for event clearing a previous event"`
	CriticalPattern []string `short:"c" long:"criticalpattern" description:"pattern for critical event"`
	WarningPattern  []string `short:"w" long:"warningpattern" description:"pattern for warning event"`
	UnknownPattern  []string `short:"u" long:"unknownpattern" description:"pattern meaning we don't know yet"`
}

func ProcessLog() (nagios.Status, error) {
	var fp checklogfile.ReadSeekCloser

	fp, err := os.Open(opts.Logfile)
	if err != nil {
		return nagios.UNKNOWN, err
	}
	ext := filepath.Ext(opts.Logfile)
	fp = checklogfile.NewCompressorSeekWrapper(fp, ext)
	offset := int64(0)
	lf := checklogfile.NewLogFile(fp, offset)
	defer lf.Close()

	if err := lf.AddPatterns(checklogfile.MonitorOk, opts.OkPattern); err != nil {
		return nagios.UNKNOWN, err
	}
	if err := lf.AddPatterns(checklogfile.MonitorCritical, opts.CriticalPattern); err != nil {
		return nagios.UNKNOWN, err
	}
	if err := lf.AddPatterns(checklogfile.MonitorWarning, opts.WarningPattern); err != nil {
		return nagios.UNKNOWN, err
	}
	if err := lf.AddPatterns(checklogfile.MonitorUnknown, opts.UnknownPattern); err != nil {
		return nagios.UNKNOWN, err
	}
	res, count, err := lf.Scan()
	lines := int64(0)
	for i, v := range count {
		lines += v
		level := strings.ToLower(nagios.Status(i).String())
		nagios.Perfdata(opts.Tag+"-"+level, float64(v), "", nil, nil)
	}
	return nagios.Status(res), err
}

var opts = &Options{}

func main() {
	args, err := flags.Parse(opts)

	if err != nil {
		nagios.Exit(nagios.UNKNOWN, err.Error())
	}

	for _, a := range args {
		fmt.Println(a)
	}
	state, err := ProcessLog()
	if err == nil {
		nagios.Exit(state, "")
	}
	nagios.Exit(state, err.Error())
}
