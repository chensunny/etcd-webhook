package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/plugin/webhook/generic"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apiserver/pkg/admission/plugin/webhook/request"

	"github.com/chensunny/etcd-webhook/e2e"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WebHookTestSuite struct {
	suite.Suite
	router      *gin.Engine
	etcdCluster etcdCluster
}

func TestWebHookTestSuite(t *testing.T) {
	suite.Run(t, new(WebHookTestSuite))
}

func (suite *WebHookTestSuite) SetupSuite() {
	suite.router = newRouter()
	var err error
	e2e.EtcdMain(binDir, certDir)
	suite.etcdCluster, err = e2e.NewEtcdProcessCluster(&e2e.ConfigNoTLS)
	if err != nil {
		log.Fatal(err)
	}

}

func (suite *WebHookTestSuite) TearDownSuite() {
	suite.etcdCluster.Stop()
}

func (suite *WebHookTestSuite) Test_Handler() {

	tests := []struct {
		name       string
		in         runtime.Object
		out        runtime.Object
		respStatus int
		allowed    bool
	}{
		{
			name: " test without annotation",
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "etcd-0",
				},
			},
			out:        &corev1.Pod{},
			respStatus: 400,
		},
		{
			name: " test  normal",
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "etcd-0",
					Annotations: map[string]string{
						admissionWebhookAnnotationInjectKey: "yes",
					},
				},
				Status: corev1.PodStatus{PodIP: "localhost"},
			},
			out:        &corev1.Pod{},
			respStatus: 200,
			allowed:    true,
		},
		{
			name: " test  duplicate remove",
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "etcd-0",
					Annotations: map[string]string{
						admissionWebhookAnnotationInjectKey: "yes",
					},
				},
				Status: corev1.PodStatus{PodIP: "localhost"},
			},
			out:        &corev1.Pod{},
			respStatus: 200,
			allowed:    false,
		},
	}

	auser := &user.DefaultInfo{}
	for _, test := range tests {
		attr := &generic.VersionedAttributes{
			Attributes:         admission.NewAttributesRecord(test.out, nil, schema.GroupVersionKind{}, "", "", schema.GroupVersionResource{}, "", admission.Operation(""), false, auser),
			VersionedOldObject: nil,
			VersionedObject:    test.in,
		}
		request := request.CreateAdmissionReview(attr)
		statsuCode, admissionReview := post("/remove", &request, suite.router)
		suite.Equal(statsuCode, test.respStatus)
		if statsuCode == 200 {
			suite.Equal(admissionReview.Response.Allowed, test.allowed)
		}
	}
}

func post(uri string, obj interface{}, router *gin.Engine) (int, *admissionv1beta1.AdmissionReview) {
	jsonByte, _ := buildBody(obj)
	req := httptest.NewRequest("POST", uri, bytes.NewReader(jsonByte))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// 调用相应的handler接口
	router.ServeHTTP(w, req)
	result := w.Result()
	defer result.Body.Close()

	body, _ := ioutil.ReadAll(result.Body)
	response := &admissionv1beta1.AdmissionReview{}
	json.Unmarshal(body, response)
	return result.StatusCode, response
}

func buildBody(obj interface{}) (data []byte, err error) {
	switch t := obj.(type) {
	case runtime.Object:
		// callers may pass typed interface pointers, therefore we must check nil with reflection
		if reflect.ValueOf(t).IsNil() {
			return data, nil
		}
		admissionScheme := runtime.NewScheme()
		admissionv1beta1.AddToScheme(admissionScheme)

		//cfg := restclient.ContentConfig{
		//	ContentType: runtime.ContentTypeJSON,
		//	NegotiatedSerializer: serializer.NegotiatedSerializerWrapper(runtime.SerializerInfo{
		//		Serializer: serializer.NewCodecFactory(admissionScheme).LegacyCodec(admissionv1beta1.SchemeGroupVersion),
		//	}),
		//	GroupVersion: &admissionv1beta1.SchemeGroupVersion,
		//}

		serializers, _ := createSerializers(defaultContentConfig())
		data, _ = runtime.Encode(serializers.Encoder, t)

	default:
		return nil, fmt.Errorf("unknown type used for body: %+v", obj)
	}
	return data, nil
}

func createSerializers(config restclient.ContentConfig) (*restclient.Serializers, error) {
	mediaTypes := config.NegotiatedSerializer.SupportedMediaTypes()
	contentType := config.ContentType
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("the content type specified in the client configuration is not recognized: %v", err)
	}
	info, ok := runtime.SerializerInfoForMediaType(mediaTypes, mediaType)
	if !ok {
		if len(contentType) != 0 || len(mediaTypes) == 0 {
			return nil, fmt.Errorf("no serializers registered for %s", contentType)
		}
		info = mediaTypes[0]
	}

	internalGV := schema.GroupVersions{
		{
			Group:   config.GroupVersion.Group,
			Version: runtime.APIVersionInternal,
		},
		// always include the legacy group as a decoding target to handle non-error `Status` return types
		{
			Group:   "",
			Version: runtime.APIVersionInternal,
		},
	}

	s := &restclient.Serializers{
		Encoder: config.NegotiatedSerializer.EncoderForVersion(info.Serializer, *config.GroupVersion),
		Decoder: config.NegotiatedSerializer.DecoderToVersion(info.Serializer, internalGV),

		RenegotiatedDecoder: func(contentType string, params map[string]string) (runtime.Decoder, error) {
			info, ok := runtime.SerializerInfoForMediaType(mediaTypes, contentType)
			if !ok {
				return nil, fmt.Errorf("serializer for %s not registered", contentType)
			}
			return config.NegotiatedSerializer.DecoderToVersion(info.Serializer, internalGV), nil
		},
	}
	if info.StreamSerializer != nil {
		s.StreamingSerializer = info.StreamSerializer.Serializer
		s.Framer = info.StreamSerializer.Framer
	}

	return s, nil
}

func defaultContentConfig() restclient.ContentConfig {
	gvCopy := corev1.SchemeGroupVersion
	return restclient.ContentConfig{
		ContentType:          "application/json",
		GroupVersion:         &gvCopy,
		NegotiatedSerializer: serializer.DirectCodecFactory{CodecFactory: scheme.Codecs},
	}
}
