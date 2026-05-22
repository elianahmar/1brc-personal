package files

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/throwea/1brc-go/pkg/model"
	u "github.com/throwea/1brc-go/pkg/utils"
)

func CreateDir(dmy string) {
	// check if the directory is present
	newDir := fmt.Sprintf("documentation/%s", dmy)
	u.PanicE(os.ReadDir("documentation"))
	u.PanicE(struct{}{}, os.MkdirAll(newDir, 0o755))
}

//
// func ChunkFile(path string) []Range {
// 	file, _ := os.Open(path)
// 	defer file.Close()
//
// 	info, _ := file.Stat()
// 	size := info.Size()
//
// 	workers := 4
//
// 	chunkSize := size / int64(workers)
// 	fmt.Println("chunksize = ", strconv.FormatInt(chunkSize, 10))
//
// 	ranges := make([]Range, 0, workers)
// 	lastOffset := int64(0)
// 	for i := 0; i < workers; i++ {
// 		buffer := make([]byte, chunkSize)
// 		n := utils.PanicE(file.ReadAt(buffer, chunkSize))
// 		// Find the last index byte and set that as the end
// 		lastNewline := int64(bytes.LastIndexByte(buffer, '\n'))
// 		distFromNewLine := len(buffer) - lastNewLine
// 		ranges = append(ranges, Range{Start: lastOffset, End: int64(lastNewline)})
// 		lastOffset = lastNewline + 1
// 	}
// 	return ranges
// }

func ChunkFile(path string) {
	file, _ := os.Open(path)
	defer file.Close()

	chunkSize := 16 // bytes
	info, err := file.Stat()
	if err != nil {
		panic("can't file.Stat()")
	}
	fmt.Printf("file size = %d\n", info.Size())

	buffer := make([]byte, chunkSize)
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Printf("Error reading the file: %v\n", err)
		}
		if bytesRead == 0 {
			break
		}
		fmt.Println(string(buffer))
	}
}

// IDEA: what if I read the chunks and pushed each chunk to a buffered channel.
// Then, I would need to do all of this upfront
func ChunkFileImproved(path string) []model.Range {
	file, _ := os.Open(path)

	chunkSize := 4 * 1024 * 1024 // 4mb
	info, err := file.Stat()
	if err != nil {
		panic("no stat")
	}
	buffer := make([]byte, chunkSize)
	fileSize := info.Size()
	maxLen := fileSize / int64(chunkSize)
	remainder := fileSize%int64(chunkSize) > 0
	if remainder {
		maxLen++
	}
	ranges := make([]model.Range, 0, maxLen)
	lastOffset := int64(0)
	newline := byte('\n')
	for {
		bytesRead, err := file.ReadAt(buffer, lastOffset)
		if err != nil && err != io.EOF {
			panic("Error reading the file")
		}
		if bytesRead == 0 {
			break
		}
		lastNewline := int64(bytes.LastIndexByte(buffer, newline))
		ending := int64(lastOffset + lastNewline)
		ranges = append(ranges, model.Range{Start: lastOffset, End: ending})
		lastOffset = ending + 1
	}
	return ranges
}

func ChunkFileAsync(path string, rangeChan chan model.Range) []model.Range {
	file, _ := os.Open(path)

	chunkSize := 4 * 1024 * 1024 // 4mb
	info, err := file.Stat()
	if err != nil {
		panic("no stat")
	}
	buffer := make([]byte, chunkSize)
	fileSize := info.Size()
	maxLen := fileSize / int64(chunkSize)
	remainder := fileSize%int64(chunkSize) > 0
	if remainder {
		maxLen++
	}
	ranges := make([]model.Range, 0, maxLen)
	lastOffset := int64(0)
	newline := byte('\n')
	for i := 0; ; i++ {
		bytesRead, err := file.ReadAt(buffer, lastOffset)
		if err != nil && err != io.EOF {
			panic("Error reading the file")
		}
		if bytesRead == 0 {
			break
		}
		lastNewline := int64(bytes.LastIndexByte(buffer, newline))
		ending := int64(lastOffset + lastNewline)
		rangeChan <- model.Range{Start: lastOffset, End: ending, Index: i}
		lastOffset = ending + 1
	}
	close(rangeChan)
	return ranges
}
