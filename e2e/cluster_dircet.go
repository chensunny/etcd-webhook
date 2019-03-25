package e2e

func newEtcdProcess(cfg *etcdServerProcessConfig) (etcdProcess, error) {
	return newEtcdServerProcess(cfg)
}
