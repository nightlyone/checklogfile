package checklogfile

import (
	"os"
	"testing"
)

var patterns = [MonitorCount][]string{
	MonitorCritical: []string{`^\d{4}-\d{2}-\d{2}\ \d{2}\:\d{2}\:\d{2}\,\d{3}\ ERROR\ ((error message:.*)|)`},
	MonitorWarning: []string{
		`^\d{4}-\d{2}-\d{2}\ \d{2}\:\d{2}\:\d{2}\,\d{3}\ INFO Packages that are upgraded:\ \$`,
		`^\d{4}-\d{2}-\d{2}\ \d{2}\:\d{2}\:\d{2}\,\d{3}\ WARNING package '.*' upgradable but fails to be marked for upgrade`,
	},
	MonitorOk: []string{`^\d{4}-\d{2}-\d{2}\ \d{2}\:\d{2}\:\d{2}\,\d{3}\ INFO No packages found that can be upgraded unattended\$`},
}

func TestUnattendedUpdate(t *testing.T) {
	fp, err := os.Open("testdata/unattended-upgrades.log")
	lf := NewLogFile(fp, 0)
	for level, pa := range patterns {
		for _, p := range pa {
			lf.AddPattern(MonitoringResult(level), p)
		}
	}
	res, counts, err := lf.Scan()
	t.Logf("res = %v, counts = %+v, err = %v, offset = %d", res, counts, err, lf.Offset())
}
