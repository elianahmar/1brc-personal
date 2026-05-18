package files

import (
	"fmt"
	"os"
	"testing"
)

func test_ChunkFile(t *testing.T) {
	ChunkFile("../../testdata/test_file.txt")
}

func Test_ChunkFileImproved(t *testing.T) {
	path := "../../../1brc-go/small_measurements.txt"
	ranges := ChunkFileImproved(path)
	fmt.Println(ranges)
	file, _ := os.Open(path)
	for i, part := range ranges {
		fmt.Printf("Buffer %d =========\n", i)

		buff := make([]byte, part.End-part.Start+1)
		file.ReadAt(buff, part.Start)
		fmt.Println(string(buff))
	}
}

// func Test_ByteReading(t *testing.T) {
// 	b := []byte("Baltimore;12.0\nNew York City;-1.0")
// 	reader := bytes.NewReaderSize(b, 16)
// 	reader.ReadBytes
// }
