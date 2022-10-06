package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

type SparkyCompressor struct {
	FileName string
	Path     string
}

func main() {
	fn := "example.txt"
	p := "./tmp/"

	sc := NewSparkyCompressor(fn, p)

	f := sc.open()
	r := sc.SparkyCompress(f)

	compressedContent, err := io.ReadAll(r)
	if err != nil {
		panic(fmt.Errorf("failed to read compression stream: %w", err))
	}
	fmt.Println("compression complete, compressed bytes: %v", compressedContent)
}

func NewSparkyCompressor(fileName, path string) *SparkyCompressor {
	return &SparkyCompressor{
		FileName: fileName,
		Path:     path,
	}
}

func (s *SparkyCompressor) SparkyCompress(rc io.ReadCloser) io.Reader {
	defer rc.Close()

	buf, err := io.ReadAll(rc)
	if err != nil {
		fmt.Printf("failed to read uncompressed file to buffer: %v", err)
	}

	f, err := s.createCompressedFile()
	if err != nil {
		fmt.Printf("failed to create file: %v", err)
	}

	gw := gzip.NewWriter(f)
	_, err = gw.Write(buf)
	if err != nil {
		fmt.Printf("failed to compress buffer into target file: %v", err)
	}
	gw.Close()

	d, err := os.Open("./tmp/example.gz")
	if err != nil {
		fmt.Printf("failed to open compressed file: %v", err)
	}

	cbuf, err := io.ReadAll(d)
	if err != nil {
		fmt.Printf("failed to read compressed file to buffer: %v", err)
	}
	r := bytes.NewReader(cbuf)

	return r
}

func (s *SparkyCompressor) open() io.ReadCloser {
	fmt.Println("open file")
	f, err := os.Open(s.Path + s.FileName)
	if err != nil {
		panic(fmt.Errorf("failed to open file: %w", err))
	}

	return f
}

func (s *SparkyCompressor) createCompressedFile() (*os.File, error) {
	name := strings.Replace(s.FileName, ".txt", ".gz", -1)

	return os.Create(s.Path + name)
}
