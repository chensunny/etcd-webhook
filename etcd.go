package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"

	"github.com/coreos/etcd/clientv3"
	corev1 "k8s.io/api/core/v1"
		"strings"
	"time"
)

const DefaultRequestTimeout = 3 * time.Second
const DefaultDialTimeout = 2 * time.Second

/*

 这里需要考虑
   * case1 webhook 响应超时，但是真实member 已经被删除的场景
   * case2 进程出现异常，不响应的场景

*/

func removeEtcdMemberByIP(pod *corev1.Pod) error {
	/*
		      *  如果出现 case1, 在 webhook 会重新 callback，但是显然，节点已经下线，所以应该返回dail 异常，这时候正常来讲应该返回正常的
			  *  如果出现 case2,应该是下线失败。

			但是当前是没办法区分是case1 还是 case2,而且没有memberlist 的列表（状态 不再webhook 维护，所以这块为什么etcd-operator 需要维护cluste的信息），
			也就没法通过通过其他节点让该member 下线

			所以这块直接返回err，让运维的同学做接入
	*/
	clientURL := "http://" + pod.Status.PodIP + ":2379"
	resp, err := listEtcdMembers([]string{clientURL}, nil)
	if err != nil {
		return err
	}
	/*

		这块没有的话也是直接返回err
	*/
	id, err := getCurrentEtcdMemberId(resp, clientURL, pod.Name)
	if err != nil {
		return err
	}
	/*
	   //TODO
	   这里考虑该 member 被隔离的场景,如果已经是一个孤立的节点，则下线的时候，需要通知主集群
	*/
	clientURLs := listClientURLs(resp)
	removeEtcdMember(clientURLs, nil, id)
	return nil
}

/*

 http://localhost:2380
 http://etcd-1.etcd:2380

*/
func getCurrentEtcdMemberId(mlresp *clientv3.MemberListResponse, podIP string, podName string) (uint64, error) {
	for _, member := range mlresp.Members {
		// 这块需要hostname or 域名的场景
		if InStringArr(member.ClientURLs, podIP) {
			return member.ID, nil
		}
        if InStringArr(member.ClientURLs, podName) {
            return member.ID, nil
        }
	}
	return 0, errors.New("not find")
}

func listClientURLs(mlresp *clientv3.MemberListResponse) (ClientURLs []string) {
	for _, member := range mlresp.Members {
		ClientURLs = append(ClientURLs, member.ClientURLs...)
	}
	return
}

func listEtcdMembers(clientURLs []string, tc *tls.Config) (*clientv3.MemberListResponse, error) {
	cfg := clientv3.Config{
		Endpoints:   clientURLs,
		DialTimeout: DefaultDialTimeout,
		TLS:         tc,
	}
	etcdcli, err := clientv3.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("list members failed: creating etcd client failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	resp, err := etcdcli.MemberList(ctx)
	cancel()
	etcdcli.Close()
	return resp, err
}

func removeEtcdMember(clientURLs []string, tc *tls.Config, id uint64) error {
	cfg := clientv3.Config{
		Endpoints:   clientURLs,
		DialTimeout: DefaultDialTimeout,
		TLS:         tc,
	}
	etcdcli, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	defer etcdcli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	_, err = etcdcli.Cluster.MemberRemove(ctx, id)
	cancel()
	return err
}

func InStringArr(arr []string, target string) bool {
	for _, val := range arr {
		if strings.Contains(val,target) {
			return true
		}
	}
	return false
}
