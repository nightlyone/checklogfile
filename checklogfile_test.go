package checklogfile

import (
	"bytes"
	"io/ioutil"
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
	if err != nil {
		t.Fatal("testdata not available: ", err)
		return
	}
	lf := NewLogFile(fp, 0)
	for level, pa := range patterns {
		for _, p := range pa {
			lf.AddPattern(MonitoringResult(level), p)
		}
	}
	res, counts, err := lf.Scan()
	t.Log("Parsing result: counts = %+v, offset = %d", counts, lf.Offset())
	if err != nil {
		t.Errorf("unexpected error %v", err)
	} else if res != MonitorCritical {
		t.Errorf("got res = %s, expected %s", res, MonitorCritical)
	} else {
		t.Logf("res = %v")
	}
}

func BenchmarkUnattendedUpdate(b *testing.B) {
	contents, err := ioutil.ReadFile("testdata/unattended-upgrades.log")
	if err != nil {
		b.Fatal("testdata not available: ", err)
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		lf := NewLogFile(NewReadSeeker(contents), 0)
		for level, pa := range patterns {
			for _, p := range pa {
				lf.AddPattern(MonitoringResult(level), p)
			}
		}
		b.StartTimer()
		_, _, err = lf.Scan()
		b.StopTimer()
		if err != nil {
			b.Fatal("invalid testdata: ", err)
			return
		}
		b.SetBytes(lf.Offset())
	}
}

type closableReadSeeker struct {
	*bytes.Reader
}

func (closableReadSeeker) Close() error { return nil }
func NewReadSeeker(b []byte) *closableReadSeeker {
	return &closableReadSeeker{Reader: bytes.NewReader(b)}
}

