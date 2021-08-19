package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// first, we define an interface for our data types
type RecordList interface {
	Unmarshal([]byte) error
	Process()
}

// mapping data is stored in the first format.
type Rec1Rec2Map struct {
	Id2      string `json:"rec2_id"`
	Rel_type int    `json:"rel_type"`
}

// define a struct for the first format
type Rec1 struct {
	Id     string        `json:"rec1_id"`
	Val1   bool          `json:"val1"`
	Val2   float64       `json:"val2"`
	Val3   int           `json:"val3"`
	R12Map []Rec1Rec2Map `json:"rec1_2_map"`
}

// Because we're passing structs, we have to
// attach Unmarshal to the struct as a method
func (r1 *Rec1) Unmarshal(br []byte) error {
	return json.Unmarshal(br, &r1)
}

// Handle arbitrary record processing.
// Any number of variations could be handled with
// modifications to this method +
// the UnmarshalStream func.
func (r1 *Rec1) Process() {
	fmt.Println(r1)
}

// define a struct for the secod format
type Rec2 struct {
	Id   string  `json:"rec2_id"`
	Val1 string  `json:"val1"`
	Val2 float64 `json:"val2"`
	Val3 int     `json:"val3"`
}

// Because we're passing structs, we have to
// attach Unmarshal to the struct as a method
func (r2 *Rec2) Unmarshal(br []byte) error {
	return json.Unmarshal(br, &r2)
}

// Handle arbitrary record processing.
// Any number of variations could be handled with
// modifications to this method +
// the UnmarshalStream func.
func (r2 *Rec2) Process() {
	fmt.Println(r2)
}

//custom error type for non-empty buffer
type BufferNotEmptyError struct {
	BufferLen int
}

func (e *BufferNotEmptyError) Error() string {
	return fmt.Sprintf("Buffer not empty: current length %v", e.BufferLen)
}

// The below process takes the input stream of []bytes, the struct
// matching the tye of the data, and two WaitGroups to avoid races.
func processStream(fstream <-chan []byte, data RecordList, outchan chan RecordList, readgroup, writegroup *sync.WaitGroup) error {
	var trackerpos int
	var bytebuffer []byte

	for {
		// receive a new chunk
		received := <-fstream
		// append to remaining bytes from last chunk parsing
		bytes := append(bytebuffer, received...)
		// notify that we're done using the input []byte
		readgroup.Done()
		// clear the readbuffer
		bytebuffer = make([]byte, 0)
		// keep track of completion of a chunk
		notdone := true
		// if we recieved data (not a zero-value for []byte indicating channel closed)
		if len(received) != 0 {
			// operate on bytestring
			bstr := string(bytes)
			for notdone && len(bstr) > 0 {
				// remove leading commas or left brackets - we don't care about
				// JSON array literal characers, and the head of the buffer should
				// always be the start of a new record - we shouldn't be in
				// the middle of the record at the head of the string([]byte), except
				// for malformed JSON
				if bstr[0] == ',' || bstr[0] == '[' {
					bstr = bstr[1:]
				}
				// check up to the last found '}'
				checkbytes := []byte(bstr[:trackerpos])
				isvalid := json.Valid(checkbytes)
				// if valid, unmarshal the single record
				if isvalid {
					writegroup.Wait()
					writegroup.Add(1)
					data.Unmarshal(checkbytes)
					outchan <- data
					// in this case, the last character was a '}' and there's no more
					// data to consume
					if len(bstr) < trackerpos+1 {
						bstr = ""
					} else {
						bstr = bstr[trackerpos+1:]
					}
					// reset the position of the last-found bracket
					trackerpos = 0
				} else {
					// search for the next bracket that may terminate a record
					next_ix := strings.Index(bstr[trackerpos:], "}")
					// if we found a bracket
					if next_ix != -1 {
						trackerpos += next_ix + 1
						// we've consumed all data possible. We add the remaining to our
						// buffer, and exit the iteration
					} else {
						notdone = false
						bytebuffer = []byte(bstr)
						trackerpos = 0
					}
				}
			}
		} else {
			// remove space characters at end
			bstr := strings.TrimSpace(string(bytes))
			// we ignore square braces at the head & end of file, so we
			// don't consider having one at the end to be meaningful data
			if len(bytes) > 0 && bstr != "]" {
				return &BufferNotEmptyError{len(bytes)}
			}
			return nil
		}
	}
}

//
func UnmarshalStream(fpath string, rec RecordList, blocksize int) error {
	var err error
	buf := make([]byte, blocksize)
	fstream := make(chan []byte, 5)
	donechan := make(chan bool)
	outchan := make(chan RecordList, 5)
	var readgroup, writegroup sync.WaitGroup

	// open the  file and check for any errors
	sfile, err := os.Open(fpath)
	if err != nil {
		fmt.Println(err)
	}
	// process the records via the struct-specific Process()
	go func() {
		for {
			if v := <-outchan; v != nil {
				v.Process()
				writegroup.Done()
			} else {
				break
			}
		}
		// we've completed the processing of all records
		donechan <- true
	}()

	// process the stream of data
	go func() {
		err = processStream(fstream, rec, outchan, &readgroup, &writegroup)
		// close the output channel - once procesStream returns, we won't
		// be sending any more values over it
		close(outchan)
	}()

	// read blocks of data from the file, using a readgroup to prevent
	// overwriting the []byte sent over the chan (a data race).
	go func() {
		for {
			d, rerr := sfile.Read(buf)
			readgroup.Add(1)
			if d > 0 {
				fstream <- buf[:d]
				readgroup.Wait()
			}
			if rerr == io.EOF {
				// file is consumed. Close the file stream & exit the function
				close(fstream)
				break
			}
		}
	}()
	// wait until receipt of a complete status from the record processor
	<-donechan
	// close the file
	sfile.Close()
	return err
}

// run through each of the three test files.
func main() {
	fmt.Println("Processing File 1")
	err := UnmarshalStream("sample_data/data_format1.json", &Rec1{}, 200)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("\nProcessing File 2")
	err = UnmarshalStream("sample_data/data_format2.json", &Rec2{}, 500)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("\nProcessing File 3")
	err = UnmarshalStream("sample_data/data_format1_invalid.json", &Rec1{}, 300)
	if err != nil {
		fmt.Println(err.Error())
	}
}
