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
	ranges := ChunkFileImproved("../../testdata/test_file.txt")
	fmt.Println(ranges)
	file, _ := os.Open("../../testdata/test_file.txt")
	for i, part := range ranges {
		fmt.Printf("Buffer %d =========\n", i)

		buff := make([]byte, part.End-part.Start+1)
		file.ReadAt(buff, part.Start)
		fmt.Println(string(buff))
	}
}
