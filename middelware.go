package main

import (
	"encoding/json"
	"strings"

	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

const (
	admissionWebhookAnnotationInjectKey = "etcd.web-hook.me/remove"
)

func Validate() gin.HandlerFunc {
	return func(c *gin.Context) {

		ar := v1beta1.AdmissionReview{}
		if err := K8SJSON.Bind(c, &ar); err != nil {
			c.AbortWithError(400, err)
			return
		}

		req := ar.Request
		var pod corev1.Pod
		if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			c.AbortWithError(400, err)
			return
		}

		glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
			req.Kind, req.Namespace, req.Name, pod.Name, req.UID, req.Operation, req.UserInfo)

		metadata := pod.ObjectMeta

		// skip special kubernete system namespaces
		for _, namespace := range ignoredNamespaces {
			if metadata.Namespace == namespace {
				glog.Infof("Skip mutation for %v for it' in special namespace:%v", metadata.Name, metadata.Namespace)
				c.AbortWithError(400, fmt.Errorf("Skip mutation for %v for it' in special namespace:%v", metadata.Name, metadata.Namespace))
				return
			}
		}

		annotations := metadata.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}

		switch strings.ToLower(annotations[admissionWebhookAnnotationInjectKey]) {
		default:
			glog.Infof("Skip mutation without special annotaion :%v", admissionWebhookAnnotationInjectKey)
			c.AbortWithError(400, fmt.Errorf("Skip mutation without special annotaion :%v", admissionWebhookAnnotationInjectKey))
			return
		case "y", "yes", "true", "on":
			glog.Infof("Mutation policy for %v/%v ", metadata.Namespace, metadata.Name)
			c.Set("web-hook-pod", &pod)
			c.Set("req-id", req.UID)
			c.Next()
		}

	}
}
