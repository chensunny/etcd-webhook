package main

import (
	"github.com/chensunny/etcd-webhook/e2e"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

type etcdCluster interface {
	Stop() (err error)
}

type EtcdTestSuite struct {
	suite.Suite
	etcdCluster etcdCluster
}

func TestEtcdTestSuite(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}

func (suite *EtcdTestSuite) SetupSuite() {
	var err error
	e2e.EtcdMain(binDir, certDir)
	suite.etcdCluster, err = e2e.NewEtcdProcessCluster(&e2e.ConfigNoTLS)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *EtcdTestSuite) TearDownSuite() {
	suite.etcdCluster.Stop()
}

func (suite *EtcdTestSuite) Test_removeMembers() {
	resp, err := listEtcdMembers([]string{"http://localhost:2379"}, nil)
	suite.Nil(err)

	id, err := getCurrentEtcdMemberId(resp, "http://localhost:2379", "etcd-0")
	suite.NotZero(id)

	clientURLs := listClientURLs(resp)
	suite.Equal(3, len(clientURLs))

	err = removeEtcdMember(clientURLs, nil, id)
	suite.Nil(err)

	resp, _ = listEtcdMembers([]string{"http://localhost:2384"}, nil)
	clientURLs = listClientURLs(resp)
	suite.Equal(2, len(clientURLs))

}
