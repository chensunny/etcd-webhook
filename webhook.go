package main

import (
	"errors"
	"net/http"

	"k8s.io/api/admission/v1beta1"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Remove(c *gin.Context) {
	var err error
	v, _ := c.Get("web-hook-pod")
	if pod, ok := v.(*corev1.Pod); ok {
		glog.Infof("try remove member ip: %v", pod.Status.PodIP)
		err = removeEtcdMemberByIP(pod)
	} else {
		err = errors.New("convert to pod errors ")
	}

	admissionResponse := &v1beta1.AdmissionResponse{Allowed: true}
	if err != nil {
		admissionResponse.Allowed = false
		admissionResponse.Result = &metav1.Status{
			Message: err.Error(),
		}
	}

	admissionReview := v1beta1.AdmissionReview{}
	admissionReview.Response = admissionResponse

	v, _ = c.Get("req-id")
	if uid, ok := v.(types.UID); ok {
		admissionReview.Response.UID = uid
	}

	c.JSON(http.StatusOK, admissionReview)

}
