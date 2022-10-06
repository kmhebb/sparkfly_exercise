package sparkyCompressor

import "io"

type SparkyCompressor interface {
	SparkyCompress(rc io.ReadCloser) *io.Reader
}
