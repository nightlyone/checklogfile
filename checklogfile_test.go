package checklogfile

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

var patterns = [MonitorCount][]string{
	MonitorCritical: {`^\d{4}-\d{2}-\d{2}\ \d{2}\:\d{2}\:\d{2}\,\d{3}\ ERROR\ ((error message:.*)|)`},
	MonitorWarning: {
		`^\d{4}-\d{2}-\d{2}\ \d{2}\:\d{2}\:\d{2}\,\d{3}\ INFO Packages that are upgraded:\ \$`,
		`^\d{4}-\d{2}-\d{2}\ \d{2}\:\d{2}\:\d{2}\,\d{3}\ WARNING package '.*' upgradable but fails to be marked for upgrade`,
	},
	MonitorOk: {`^\d{4}-\d{2}-\d{2}\ \d{2}\:\d{2}\:\d{2}\,\d{3}\ INFO No packages found that can be upgraded unattended\$`},
}

func TestUnattendedUpdate(t *testing.T) {
	testfiles := []string{
		"unattended-upgrades.log",
		"unattended-upgrades.log.1",
	}
	for _, file := range testfiles {
		fp, err := os.Open("testdata/" + file)
		if err != nil {
			t.Fatalf("testdata not available: %v", err)
		}
		lf := NewLogFile(fp, 0)
		defer lf.Close()
		for level, pa := range patterns {
			for _, p := range pa {
				if err := lf.AddPattern(MonitoringResult(level), p); err != nil {
					t.Fatalf("Cannot AddPattern(%s, %#q): %v", MonitoringResult(level), p, err)
				}
			}
		}
		res, counts, err := lf.Scan()
		t.Logf("Parsing result of %s: counts = %+v, offset = %d", file, counts, lf.Offset())
		switch {
		case err != nil:
			t.Errorf("%s: unexpected error %v", file, err)
		case res != MonitorCritical:
			t.Errorf("%s:got res = %s, want res = %s", file, res, MonitorCritical)
		default:
			t.Logf("%s:got res = %v, want res = %v", file, res, MonitorCritical)
		}
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
				_ = lf.AddPattern(MonitoringResult(level), p) // error checked in test already
			}
		}
		b.StartTimer()
		_, _, err := lf.Scan() // error checked in test already
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
