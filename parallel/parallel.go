package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"
)

type codes map[string][]string
type CodeSystem struct {
	mux   *sync.RWMutex
	codes codes
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan string)
	mux := sync.RWMutex{}
	wg := &sync.WaitGroup{}
	directory := "./tmp/testdata/"
	var data = make(codes)

	sys := CodeSystem{
		mux:   &mux,
		codes: data,
	}
	// load files
	files, err := loadDirectory(directory)
	if err != nil {
		fmt.Printf("error loading files: %v", err)
		panic(err)
	}

	// read files into memory
	for _, v := range files {
		fmt.Printf("accessing file: %v\n", v.Name())
		if !v.Type().IsDir() {
			wg.Add(1)
			go func(ctx context.Context) {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						sys.mux.Lock()
						err := readCSV(directory+v.Name(), sys, wg)
						if err != nil {
							mux.Unlock()
							done <- fmt.Sprintf("error reading file, %v: %v", v.Name(), err)
							cancel()
						}
						sys.mux.Unlock()
					}
				}
			}(ctx)
		}
		continue
	}

	// check codes for matches
	sys.mux.RLock()
	defer sys.mux.RUnlock()
	for _, v := range data {
		wg.Add(1)
		code := v[1]
		go func(ctx context.Context) {
			if codeExists(code, sys, wg) {
				done <- fmt.Sprintf("matching code found: %v", code)
				cancel()
			}
		}(ctx)
	}

	// receive matching code and report
	message := <-done
	fmt.Println(message)
}

// will load files in the target directory for reading
func loadDirectory(directory string) ([]fs.DirEntry, error) {
	files, err := os.ReadDir(directory)

	if err != nil {
		return nil, err
	}

	return files, nil
}

// will read the file and add the data to the map
func readCSV(file string, data CodeSystem, wg *sync.WaitGroup) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	reader := csv.NewReader(f)

	for {
		var d []string
		d, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// skip header row
		if d[1] == "code" {
			continue
		}

		data.codes[d[1]] = d

	}

	wg.Done()

	return nil
}

func codeExists(code string, data CodeSystem, wg *sync.WaitGroup) bool {
	_, exists := data.codes[code]

	wg.Done()
	return exists
}
