package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
	K8SJSON       = k8sJsonBinding{}
)

type k8sJsonBinding struct{}

func (k8sJsonBinding) Name() string {
	return "json"
}

func (k8sJsonBinding) Bind(c *gin.Context, obj runtime.Object) error {
	//var obj runtime.Object
	//obj, ok = v.(runtime.Object)
	//if !ok {
	//    err := errors.New("invalid type ")
	//	c.AbortWithError(400, err).SetType(gin.ErrorTypeBind)
	//	return err
	//}
	var body []byte
	req := c.Request
	if req.Body != nil {
		if data, err := ioutil.ReadAll(req.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		return errors.New("empty body")
	}
	// verify the content type is accurate
	contentType := req.Header.Get("Content-Type")
	if contentType != binding.MIMEJSON {
		return errors.New("invalid Content-Type, expect `application/json`")
	}
	if _, _, err := deserializer.Decode(body, nil, obj); err != nil {
		c.AbortWithError(400, err).SetType(gin.ErrorTypeBind)
		return err
	}
	return nil
}
