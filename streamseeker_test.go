package checklogfile

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestCompressorSeeker(t *testing.T) {
	files := map[string]string{
		"bz2": "testdata/unattended-upgrades.log.bz2",
		"gz":  "testdata/unattended-upgrades.log.gz",
		"":    "testdata/unattended-upgrades.log",
	}
	tests := [...]struct {
		description string
		offset      int64
		want        int64
		whence      int
	}{
		{description: "Getting position", whence: io.SeekCurrent, offset: 0, want: 20},
		{description: "Seeking 100 bytes forward", whence: io.SeekCurrent, offset: 100, want: 120},
		{description: "Seeking to 200 bytes", whence: io.SeekStart, offset: 200, want: 200},
		{description: "Rewind to beginning", whence: io.SeekStart, offset: 0, want: 0},
	}
	want := []byte("2013-02-20 11:05:29,")
	for ext, f := range files {
		f := f
		ext := ext
		t.Run(fmt.Sprintf("with ext: %q", ext), func(t *testing.T) {
			fp, err := os.Open(f)
			if err != nil {
				t.Fatal("testdata missing. Error: ", err)
			}
			// This also ensures we have the right interface
			var cw ReadSeekCloser = NewCompressorSeekWrapper(fp, ext)
			if cw == nil {
				_ = fp.Close() // Cannot use defer, because cw.Close closes fp as well.
				t.Fatal("cannot create compressing seeker")
			}
			defer func() { _ = cw.Close() }()

			if _, ok := cw.(ReadSeekCloser); !ok {
				t.Fatalf("returned type is not a ReadSeekCloser: got %T, want ReadSeekCloser", cw)
			}

			got := make([]byte, len(want))
			n, err := cw.Read(got)
			switch {
			case err != nil:
				t.Fatalf("cannot read data: %v", err)
			case n < len(want):
				t.Fatalf("got %d, want %d bytes", n, len(want))
			case string(got) != string(want):
				t.Fatalf("data corruption: got %q, want %q", got, want)
			}

			for _, tt := range tests {
				tt := tt
				t.Run(tt.description, func(t *testing.T) {
					pos, err := cw.Seek(tt.offset, tt.whence)
					switch {
					case err != nil:
						t.Errorf("got error %s, want nil error", err)
					case pos != tt.want:
						t.Errorf("got position %d, want %d", pos, tt.want)
					default:
					}
				})
			}
		})
	}
}

func BenchmarkWrappers(b *testing.B) {
	files := map[string]string{
		"bz2": "testdata/unattended-upgrades.log.bz2",
		"gz":  "testdata/unattended-upgrades.log.gz",
		"":    "testdata/unattended-upgrades.log",
	}
	for ext, f := range files {
		ext := ext
		f := f
		b.Run(fmt.Sprintf("with ext: %q", ext), func(b *testing.B) {
			fp, err := os.Open(f)
			if err != nil {
				b.Fatal("testdata missing. Error: ", err)
			}
			cw := NewCompressorSeekWrapper(fp, ext)
			if cw == nil {
				_ = fp.Close() // Cannot use defer, because cw.Close closes fp as well.
				b.Fatal("cannot open compressing seeker")
			}
			defer func() { _ = cw.Close() }()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StartTimer()
				n, err := io.Copy(ioutil.Discard, cw)
				if err != nil {
					b.Fatalf("invalid testdata: %v", err)
				}
				b.SetBytes(n)
				b.StopTimer()
				if _, err := cw.Seek(0, io.SeekStart); err != nil {
					b.Fatalf("cannot rewind testdata: %v", err)
				}
			}
		})
	}
}
