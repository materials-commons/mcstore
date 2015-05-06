package files

import (
	"fmt"
	"mime"
	"testing"

	"github.com/rakyll/magicmime"
)

func TestUnknownMediaType(t *testing.T) {
	mtype := mime.TypeByExtension(".rhit")
	fmt.Printf("mtype = '%s'\n", mtype)
	magic, err := magicmime.New(magicmime.MAGIC_MIME)
	if err != nil {
		t.Fatalf("New failed: %s", err)
	}
	ftype, err := magic.TypeByFile("/home/gtarcea/data/1_bmp.unknown")
	fmt.Println("err = ", err)
	fmt.Println("ftype = ", ftype)
}
