package webhook

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/soer3n/yaho/pkg/client"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func quarantineHandler(w http.ResponseWriter, r *http.Request) {

	var body []byte
	var pod *corev1.Pod
	var err error

	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		log.Fatal("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	if r.URL.Path != "/validate" {
		log.Fatal("no validate")
		http.Error(w, "no validate", http.StatusBadRequest)
		return
	}

	handler := QuarantineHandler{
		body:     body,
		response: &v1beta1.AdmissionReview{},
		client:   client.New().TypedClient,
	}

	if pod, err = handler.getControllerPod(); err != nil {
		log.Fatal("error on getting controller pod")
		http.Error(w, "no validate", http.StatusBadRequest)
		return
	}

	if !handler.controllerShouldBeRescheduled(pod) {
		log.Print("controller pod is on a valid node")
		return
	}

	if err := handler.parseAdmissionResponse(); err != nil {
		log.Fatal("admission validation failed")
		http.Error(w, "admission validation failed", http.StatusBadRequest)
		return
	}

}
