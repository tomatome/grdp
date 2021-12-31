// cliprdr_test.go
package cliprdr_test

import (
	"fmt"
	"testing"

	"github.com/tomatome/grdp/plugin/cliprdr"
)

func TestClip(t *testing.T) {
	//t1, _ := cliprdr.ReadAll()
	//fmt.Printf("%s\n", t1)
	ok := cliprdr.OpenClipboard(0)
	fmt.Println(ok)
	if ok {
		name := cliprdr.GetClipboardData(13)
		fmt.Printf("name=%s\n", name)

		cliprdr.CloseClipboard()
	}
}
