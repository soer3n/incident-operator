package webhook

import (
	"io/ioutil"
	"log"
	"net/http"

	"k8s.io/api/admission/v1beta1"
)

func quarantineHandler(w http.ResponseWriter, r *http.Request) {

	var body []byte

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
	}

	if err := handler.parseAdmissionResponse(); err != nil {
		log.Fatal("admission failed")
		http.Error(w, "admission failed", http.StatusBadRequest)
		return
	}

}
