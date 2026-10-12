package admin

import (
	"io/fs"
	"os"
)

var DistFS fs.FS = os.DirFS("dist")
