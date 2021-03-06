package checklogfile

import (
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
)

var ErrStreamSeek = errors.New("cannot seek to this position in stream")

// We can skip IO and rewind in streams and thus support the io.Seeker interface
type CompressorSeekWrapper struct {
	f           ReadSeekCloser
	compressor  io.Reader
	offset      int64
	compression string
}

func getCompressor(compression string, r io.Reader) (io.Reader, error) {
	switch compression {
	case "gz", ".gz":
		return gzip.NewReader(r)
	case "bz2", ".bz2":
		return bzip2.NewReader(r), nil
	}
	return r, nil
}

// wrapper around compressors that supports streams
func NewCompressorSeekWrapper(backend ReadSeekCloser, compression string) *CompressorSeekWrapper {
	compressor, err := getCompressor(compression, backend)
	if err != nil {
		return nil
	}
	return &CompressorSeekWrapper{
		compressor:  compressor,
		f:           backend,
		compression: compression,
	}
}

// We can skip IO and rewind in streams and thus support the io.Seeker interface
func (c *CompressorSeekWrapper) Seek(offset int64, whence int) (ret int64, err error) {
	// shortcut if compressors suddenly learn to seek or we have no compressor in between
	if seeker, ok := c.compressor.(io.Seeker); ok {
		ret, err := seeker.Seek(offset, whence)
		c.offset = ret
		return ret, err
	}

	var newoffset int64
	switch whence {
	case io.SeekStart:
		newoffset = offset
	case io.SeekCurrent:
		newoffset = c.offset + offset
	case io.SeekEnd:
		return c.offset, ErrStreamSeek
	}
	if newoffset != 0 && newoffset < c.offset {
		return c.offset, ErrStreamSeek
	}
	if newoffset == c.offset {
		return c.offset, nil
	}

	// Rewind to zero
	if newoffset == 0 {
		if closer, ok := c.compressor.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				return 0, err
			}
		}
		if _, err := c.f.Seek(0, io.SeekStart); err != nil {
			return 0, err
		}
		comp, err := getCompressor(c.compression, c.f)
		if err != nil {
			return 0, err
		}
		c.compressor = comp
		c.offset = 0
		return 0, nil
	}
	// seek forward means read and skip
	_, err = io.CopyN(ioutil.Discard, c, newoffset-c.offset)
	return c.offset, err
}

// satisfy io.Reader interface
func (c *CompressorSeekWrapper) Read(p []byte) (n int, err error) {
	n, err = c.compressor.Read(p)
	c.offset += int64(n)
	return n, err
}

// Satisfy io.Closer interface
func (c *CompressorSeekWrapper) Close() error {
	if c.f == c.compressor {
		if closer, ok := c.compressor.(io.Closer); ok {
			return closer.Close()
		}
	}
	if closer, ok := c.compressor.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	if closer, ok := c.f.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}
