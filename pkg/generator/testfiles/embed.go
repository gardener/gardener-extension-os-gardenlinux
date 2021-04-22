package testfiles

import "embed"

// Files contains the contents of the testfiles directory
//go:embed cloud-* containerd-* docker-*
var Files embed.FS
