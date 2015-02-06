package file

import (
	"fmt"
	"testing"
)

var _ = fmt.Println

func TestExFileInfo(t *testing.T) {
	fi, err := ExStat("exstat.go")
	if err != nil {
		t.Fatalf("ExStat failed: %s", err)
	}

	// Cheap test - just look at the values
	fmt.Println("ctime", fi.CTime())
	fmt.Println("atime", fi.ATime())
	fmt.Println("fid", fi.FID())
	fmt.Println("path", fi.Path())
}
