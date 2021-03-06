package k8s

import (
	"github.com/astaxie/beego/logs"
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"cloud/util"
	"cloud/models/app"
	"encoding/json"
)

type ServicePod struct {
	// 容器端口
	ContainerPort []int32
	// 应用名称
	AppName string
	// 集群名称
	ClusterName string
	// 容器名称
	ContainerName string
	// 协议类型
	Protocol string
	// 选择器
	Selector map[string]string
	// 资源空间
	ResouceName string
}

type AppPodStatus struct {
	// 应用名称
	AppName string
	// 镜像名称
	Image string
	// 配置信息
	Cpu    string
	Memory string
	// 宿主机地址
	HostIp string
	// 容器IP
	PodIp string
	// 状态
	Status bool
	// 重启次数
	RestartCount int32
	// 资源名称
	ResourceName string
	// 集群名称
	ClusterName string
	// 容器ID
	ContainerId int
	// 创建时间
	CreateTime string
	// 容器名称
	ContainerName string
}

// 获取pods数量
func GetPodsNumber(namespace string, clientset kubernetes.Clientset) int {
	opt := metav1.ListOptions{}

	pods, err := clientset.CoreV1().Pods(namespace).List(opt)
	if err != nil {
		logs.Error("获取k8s Pods失败", err.Error())
		return 0
	}
	return len(pods.Items)
}

// 2018-01-16 12:25
// 删除某个pod后自动重建
func DeletePod(namespace string, name string, clientset kubernetes.Clientset) error {
	err := clientset.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{})
	return err
}

// 获取pod数据
//fmt.Println(p.Status.HostIP)
//{"metadata":{"name":"zhaoyun1-rc-28fp6","generateName":"zhaoyun1-rc-","namespace":"default","selfLink":"/api/v1/namespaces/default/pods/zhaoyun1-rc-28fp6","uid":"29676a19-dbbd-11e7-a7e2-0894ef37b2d2","resourceVersion":"287211","creationTimestamp":"2017-12-08T02:11:54Z","labels":{"app":"www-gg-com","max-scale":"3","min-scale":"3"},"annotations":{"kubernetes.io/created-by":"{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicationController\",\"namespace\":\"default\",\"name\":\"zhaoyun1-rc\",\"uid\":\"2966d58b-dbbd-11e7-a7e2-0894ef37b2d2\",\"apiVersion\":\"v1\",\"resourceVersion\":\"287184\"}}\n"},"ownerReferences":[{"apiVersion":"v1","kind":"ReplicationController","name":"zhaoyun1-rc","uid":"2966d58b-dbbd-11e7-a7e2-0894ef37b2d2","controller":true,"blockOwnerDeletion":true}]},"spec":{"containers":[{"name":"zhaoyun1","image":"nginx:1.11","ports":[{"containerPort":80,"protocol":"TCP"}],"resources":{"limits":{"cpu":"1","memory":"0"},"requests":{"cpu":"1","memory":"0"}},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","imagePullPolicy":"IfNotPresent"}],"restartPolicy":"Always","terminationGracePeriodSeconds":30,"dnsPolicy":"ClusterFirst","nodeName":"10.16.55.103","securityContext":{},"schedulerName":"default-scheduler"},"status":{"phase":"Running","conditions":[{"type":"Initialized","status":"True","lastProbeTime":null,"lastTransitionTime":"2017-12-08T02:10:19Z"},{"type":"Ready","status":"True","lastProbeTime":null,"lastTransitionTime":"2017-12-08T02:10:23Z"},{"type":"PodScheduled","status":"True","lastProbeTime":null,"lastTransitionTime":"2017-12-08T02:11:54Z"}],"hostIP":"10.16.55.103","podIP":"172.16.8.15","startTime":"2017-12-08T02:10:19Z","containerStatuses":[{"name":"zhaoyun1","state":{"running":{"startedAt":"2017-12-08T02:10:22Z"}},"lastState":{},"ready":true,"restartCount":0,"image":"nginx:1.11","imageID":"docker-pullable://nginx@sha256:e6693c20186f837fc393390135d8a598a96a833917917789d63766cab6c59582","containerID":"docker://c0a1cae85d6146d415996252750add084373c0f0e90c68fb129e3aa440262645"}],"qosClass":"Burstable"}}
func GetPods(namespace string, clientset kubernetes.Clientset) []v1.Pod {
	opt := metav1.ListOptions{}
	pods, err := clientset.CoreV1().Pods(namespace).List(opt)
	if err != nil {
		logs.Error("获取Pods错误", err.Error())
		return make([]v1.Pod, 0)
	}
	return pods.Items
}

// 获取某个服务的pods
// @param namespace
// @param serviceName
// 2018-01-18 9:53
func GetPodsService(namespace string, serviceName string, clientset kubernetes.Clientset) []v1.Pod  {
	opt := metav1.ListOptions{}
	opt.LabelSelector = "name="+ serviceName
	pods, err := clientset.CoreV1().Pods(namespace).List(opt)
	if err != nil {
		logs.Error("获取Pods错误", err.Error())
		return make([]v1.Pod, 0)
	}
	return pods.Items
}

// 获取某个服务器的pod数量
func GetIpPodNumber(pods []v1.Pod, ip string) int {
	var count int
	for _, item := range pods {
		if item.Status.HostIP == ip && !strings.Contains(item.Name, "kubernetes") {
			count += 1
		}
	}
	return count
}

// 获取某个namespace下面的服务
func GetPodStatus(namespace string, clientset kubernetes.Clientset) []AppPodStatus {
	data := GetPods(namespace, clientset)
	datas := make([]AppPodStatus, 0)
	for _, d := range data {
		app := AppPodStatus{}
		app.HostIp = d.Status.HostIP
		app.PodIp = d.Status.PodIP
		obj := d.Status.ContainerStatuses[0]
		app.RestartCount = obj.RestartCount
		app.Image = obj.Image
		app.AppName = obj.Name
		app.Status = obj.Ready
		app.ContainerName = d.Name
		app.CreateTime = util.GetMinTime(util.ReplaceTime(d.CreationTimestamp.String()))
		datas = append(datas, app)
	}
	return datas
}


// 获取容器挂载的目录信息
// 2018-01-16 11:18
func getMountPath(d v1.Pod) string {
	volumn := d.Spec.Containers[0].VolumeMounts
	result := []StorageData{}
	// mount的名称和路径
	mounts := util.Lock{}
	for _, v := range d.Spec.Volumes {
		if v.HostPath != nil {
			mounts.Put(v.Name, v.HostPath.Path)
		}
	}

	for _, v := range volumn {
		data := StorageData{}
		data.ContainerPath = v.MountPath
		data.Volume = v.Name
		m, ok := mounts.Get(v.Name)
		if ok {
			data.HostPath = m.(string)
		}
		result = append(result, data)
	}
	t, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(t)
}

// 2018-02-09 08:34
// pod状态数据
func podStatus(app app.CloudContainer, obj v1.ContainerStatus) app.CloudContainer {
	app.WaitingMessages = "0"
	app.WaitingReason = "0"
	app.TerminatedReason = "0"
	app.TerminatedMessages = "0"
	if obj.State.Waiting != nil {
		app.WaitingMessages = obj.State.Waiting.Message
		app.WaitingReason = obj.State.Waiting.Reason
	}

	if 	obj.State.Terminated != nil{
		app.TerminatedMessages = obj.State.Terminated.Message
		app.TerminatedReason = obj.State.Terminated.Reason
	}
	app.Image = obj.Image
	app.AppName = obj.Name
	app.ServiceName = obj.Name
	return app
}

// {"name":"auto-service","image":"nginx:1.10","ports":[{"containerPort":80,"protocol":"TCP"}],"resources":{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"1","memory":"2Gi"}},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","imagePullPolicy":"IfNotPresent"}
// 获取某个namespace下面的服务
func GetContainerStatus(namespace string, clientset kubernetes.Clientset) []app.CloudContainer {
	data := GetPods(namespace, clientset)
	datas := make([]app.CloudContainer, 0)
	for _, d := range data {
		app := app.CloudContainer{}
		app.ServerAddress = d.Status.HostIP
		app.ContainerIp = d.Status.PodIP
		if len(d.Status.ContainerStatuses) == 0 {
			continue
		}

		obj := d.Status.ContainerStatuses[0]

		app = podStatus(app, obj)
		if app.WaitingMessages == "" {
			for _, v := range d.Status.Conditions {
				if v.Status == v1.ConditionFalse {
					app.WaitingMessages = v.Message
					app.WaitingReason = v.Reason
				}
			}
		}

		limit := d.Spec.Containers[0].Resources.Limits
		app.Cpu = limit.Cpu().Value()
		app.Memory = limit.Memory().Value() / 1024 / 1024

		envs := make([]string, 0)
		for _, v := range d.Spec.Containers[0].Env {
			envs = append(envs, v.Name+"="+v.Value+"\n")
		}
		app.StorageData = getMountPath(d)
		app.Env = strings.Join(envs, " ")
		app.Status = strings.Replace(util.ObjToString(d.Status.Phase), "\"", "", -1)
		app.ContainerName = d.Name

		app.CreateTime = util.ReplaceTime(d.CreationTimestamp.String())
		datas = append(datas, app)
	}
	return datas
}
