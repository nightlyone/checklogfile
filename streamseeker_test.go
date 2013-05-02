package checklogfile

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

var positions = [...]struct {
	description string
	offset      int64
	want        int64
	whence      int
}{
	{
		description: "Getting position",
		whence:      1, offset: 0, want: 20,
	},
	{
		description: "Seeking 100 bytes forward",
		whence:      1, offset: 100, want: 120,
	},
	{
		description: "Seeking to 200 bytes",
		whence:      0, offset: 200, want: 200,
	},
	{
		description: "Rewind to beginning",
		whence:      0, offset: 0, want: 0,
	},
}

var files = map[string]string{
	"bz2": "testdata/unattended-upgrades.log.bz2",
	"gz":  "testdata/unattended-upgrades.log.gz",
	"":    "testdata/unattended-upgrades.log",
}

func TestCompressorSeeker(t *testing.T) {
	want_line := []byte("2013-02-20 11:05:29,")
	if n := len(want_line); n != 20 {
		t.Fatalf("testsetup broken. want_line should be %d bytes, got %d bytes", 20, n)
		return
	}
	for ext, f := range files {
		fp, err := os.Open(f)
		if err != nil {
			t.Fatal("testdata missing. Error: ", err)
			return
		}
		// This also ensures we have the right interface
		var cw ReadSeekCloser
		cw = NewCompressorSeekWrapper(fp, ext)
		if cw == nil {
			t.Errorf("%s:cannot open compressing seeker", ext)
			continue
		}
		defer cw.Close()

		got_line := make([]byte, len(want_line))

		n, err := cw.Read(got_line)
		if err != nil {
			t.Error(ext, ": testdata too small (need at least 200 bytes) Error ", err)
			continue
		} else if n < 20 {
			t.Errorf("%s: testdata too small (need at least 200 bytes, got %v) Error %v", ext, n, err)
			continue
		}

		if string(want_line) != string(got_line) {
			t.Errorf("%s: decompressor b0rken: want %s, got %s", ext, want_line, got_line)
			continue
		}
		for i, p := range positions {
			t.Logf("%d:%s", i, p.description)
			if pos, err := cw.Seek(p.offset, p.whence); err != nil {
				t.Errorf("%s:%d:err: %s", ext, i, err)
			} else if pos != p.want {
				t.Errorf("%s:%d: got %d, want %d", ext, i, pos, p.want)
			} else {
				t.Logf("%s:%d: ok, got %d", ext, i, pos)
			}
		}
	}
}

func BenchmarkGzipWrapper(b *testing.B) {
	fp, err := os.Open("testdata/unattended-upgrades.log.gz")
	if err != nil {
		b.Fatal("testdata missing. Error: ", err)
		return
	}
	// This also ensures we have the right interface
	var cw ReadSeekCloser
	cw = NewCompressorSeekWrapper(fp, "gz")
	if cw == nil {
		b.Fatal("cannot open compressing seeker")
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		b.StartTimer()
		n, _ := io.Copy(ioutil.Discard, cw)
		b.StopTimer()
		if err != nil {
			b.Fatal("invalid testdata: ", err)
			return
		}
		b.SetBytes(n)
		cw.Seek(0, 0)
	}
}
