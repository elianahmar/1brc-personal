package files

import (
	"fmt"
	"os"

	u "github.com/throwea/1brc-go/pkg/utils"
)

func CreateDir(dmy string) {
	// check if the directory is present
	newDir := fmt.Sprintf("documentation/%s", dmy)
	u.PanicE(os.ReadDir("documentation"))
	u.PanicE(struct{}{}, os.MkdirAll(newDir, 0o755))
}
