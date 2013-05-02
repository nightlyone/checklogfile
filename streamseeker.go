package checklogfile

import (
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
)

var ErrStreamSeek = errors.New("Cannot seek to this position in stream")

// We can skip IO and rewind in streams and thus support the io.Seeker interface
type CompressorSeekWrapper struct {
	f           ReadSeekCloser
	compressor  io.Reader
	offset      int64
	compression string
}

func getCompressor(compression string, r io.Reader) (io.Reader, error) {
	switch compression {
	case "gz":
		return gzip.NewReader(r)
	case "bz2":
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
	case 0:
		newoffset = offset
	case 1:
		newoffset = c.offset + offset
	case 2:
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
		if _, err := c.f.Seek(0, 0); err != nil {
			return 0, err
		}
		if comp, err := getCompressor(c.compression, c.f); err != nil {
			return 0, err
		} else {
			c.compressor = comp
			c.offset = 0
			return 0, nil
		}
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
