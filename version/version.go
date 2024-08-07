package version

// Package is the overall, canonical project import path under which the
// package was built.
var Package = "github.com/syaiful6/payung"

// Version indicates which version of the binary is running. This is set to
// the latest release tag by hand, always suffixed by "+unknown". During
// build, it will be replaced by the actual version. The value here will be
// used if the registry is run after a go get based install.
var Version = "v0.2.6"

// Revision is filled with the VCS (e.g. git) revision being used to build
// the program at linking time.
var Revision = ""
