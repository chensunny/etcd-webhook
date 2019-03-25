

###  背景介绍：

> 有状态应用在删除时一般需要做清理操作，比如 etcd 集群在删除一个节点时需要先执行 remove member 才能安全删除节点 

[相关资料](https://coreos.com/etcd/docs/latest/op-guide/runtime-configuration.html#remove-a-member) 



> 利用 k8s admission validation webhook 机制实现 webhook 用以解决 etcd 集群安全下线问题。当将 statefulset 的 replicas 减小 1 时，能在 pod 删除前执行 remove member 操作，保证 3 节点的 etcd 集群在 replicas 减小到 1 时，仍能正常工作。
[其中 etcd 集群部署可参考](https://github.com/kubernetes/kubernetes/blob/master/test/e2e/testing-manifests/statefulset/etcd/statefulset.yaml)


**注**

> validate admission 本身作为一个 资源变更的一个 validate 的方式，从职责上不太适合做这个动作

### 相关设计


![](http://meitu-test.oss-cn-beijing.aliyuncs.com/%E4%BC%81%E4%B8%9A%E5%BE%AE%E4%BF%A1%E6%88%AA%E5%9B%BE_075beed3-f1ec-4775-a9b9-e5547f804c45.png)


* 相关边界
	* pod 销毁
	* node 挂掉，导致 pod 没有执行 preStop
	* statefulset replicas 减少

#### 实现方式1
 
 * 描述：  
 call webhook in statefulset 的 update 事件
 
 * 问题：
 	 *  webhook 不是controller or operator 本身是没有状态的，在变动的时候不知道是 replicas 增加还是减少，
 	 *  同时出现减少的时候，没办法同步的知道是哪个实例被删除
 	 *  如果 node 挂掉实例出现异常，需要启动新的实例自己更新状态
 
 
#### 实现方式2
 
  * 描述：  
 call webhook in pod 的 delete 事件
 
  * 问题：
  	 * 没办法区分是 pod 主动下掉，还是异常场景的迁移
  	 	 * 所以需要新启动的节点做一下初始化的工作做
 	 * 如果是 node 上出现荡机，比较难通知到其他 member （因为 pod 上没有其他member 的信息），需要多次查询 apiserver。
 	   
  
 
### 关于 Admission Webhook

[k8s 社区介绍]( https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#write-an-admission-webhook-server) 




