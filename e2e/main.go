package e2e

var (
	binDir  string
	certDir string

	certPath       string
	privateKeyPath string
	caPath         string

	certPath2       string
	privateKeyPath2 string

	crlPath               string
	revokedCertPath       string
	revokedPrivateKeyPath string
)

func EtcdMain(binDir string, certDir string) {

	binPath = binDir + "/etcd"
	ctlBinPath = binDir + "/etcdctl"
	certPath = certDir + "/server.crt"
	privateKeyPath = certDir + "/server.key.insecure"
	caPath = certDir + "/ca.crt"
	revokedCertPath = certDir + "/server-revoked.crt"
	revokedPrivateKeyPath = certDir + "/server-revoked.key.insecure"
	crlPath = certDir + "/revoke.crl"

	certPath2 = certDir + "/server2.crt"
	privateKeyPath2 = certDir + "/server2.key.insecure"

}
