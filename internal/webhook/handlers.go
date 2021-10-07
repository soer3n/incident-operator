package webhook

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/golang/glog"

	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/internal/cli"
	"github.com/soer3n/yaho/pkg/client"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func (h *QuarantineHTTPHandler) quarantineHandler(w http.ResponseWriter, r *http.Request) {

	h.mu.Lock()
	defer h.mu.Unlock()

	var body, res []byte
	var pod *corev1.Pod
	var ar *v1beta1.AdmissionReview
	var q *v1alpha1.Quarantine
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

	if ar, err = handler.getAdmissionRequestSpec(body, w); err != nil {
		log.Error("error deserializing admission request spec")
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if q, err = handler.getQuarantineSpec(ar, w); err != nil {
		log.Error("error deserializing quarantine spec")
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if pod, err = cli.GetControllerPod(client.New().TypedClient); err != nil {
		log.Error("error on getting controller pod")
		log.Fatal(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !handler.controllerShouldBeRescheduled(pod, q) {
		log.Info("controller pod is on a valid node")
	}

	if res, err = json.Marshal(handler.response); err != nil {
		log.Error("failed to parse admission response")
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Write(res)
}
