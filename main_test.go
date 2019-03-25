package main

import (
	"flag"
	"os"
	"testing"
    "runtime"
)


var (
    binDir  string
    certDir string

)
func TestMain(m *testing.M) {
    os.Setenv("ETCD_UNSUPPORTED_ARCH", runtime.GOARCH)
    os.Unsetenv("ETCDCTL_API")

    flag.StringVar(&binDir, "bin-dir", "../bin", "The directory for store etcd and etcdctl binaries.")
    flag.StringVar(&certDir, "cert-dir", "../integration/fixtures", "The directory for store certificate files.")
    flag.Parse()

	os.Exit(m.Run())
}
