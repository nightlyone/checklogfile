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

func TestGzipSeeker(t *testing.T) {
	fp, err := os.Open("testdata/unattended-upgrades.log.gz")
	if err != nil {
		t.Fatal("testdata missing. Error: ", err)
		return
	}
	// This also ensures we have the right interface
	var cw ReadSeekCloser
	cw = NewCompressorSeekWrapper(fp, "gz")
	if cw == nil {
		t.Errorf("cannot open compressing seeker")
		return
	}
	// do a bit io 
	_, err = io.CopyN(ioutil.Discard, cw, 20)
	if err != nil {
		t.Fatal("testdata too small (need at least 100 bytes) Error ", err)
		return
	}
	for i, p := range positions {
		t.Logf("%d:%s", i, p.description)
		if pos, err := cw.Seek(p.offset, p.whence); err != nil {
			t.Errorf("%d:err: %s", i, err)
		} else if pos != p.want {
			t.Errorf("%d: got %d, want %d", i, pos, p.want)
		} else {
			t.Logf("%d: ok, got %d", i, pos)
		}
	}
}

func TestBzip2Seeker(t *testing.T) {
	fp, err := os.Open("testdata/unattended-upgrades.log.bz2")
	if err != nil {
		t.Fatal("testdata missing. Error: ", err)
		return
	}
	// This also ensures we have the right interface
	var cw ReadSeekCloser
	cw = NewCompressorSeekWrapper(fp, "bz2")
	if cw == nil {
		t.Errorf("cannot open compressing seeker")
		return
	}
	// do a bit io 
	_, err = io.CopyN(ioutil.Discard, cw, 20)
	if err != nil {
		t.Fatal("testdata too small (need at least 100 bytes) Error ", err)
		return
	}
	for i, p := range positions {
		t.Logf("%d:%s", i, p.description)
		if pos, err := cw.Seek(p.offset, p.whence); err != nil {
			t.Errorf("%d:err: %s", i, err)
		} else if pos != p.want {
			t.Errorf("%d: got %d, want %d", i, pos, p.want)
		} else {
			t.Logf("%d: ok, got %d", i, pos)
		}
	}
}
