package main

import (
  "log"
  "fmt"
  "net/http"
  "io/ioutil"
  "time"
  "encoding/json"

  corev1 "k8s.io/api/core/v1"
  admission "k8s.io/api/admission/v1beta1"
  "k8s.io/apimachinery/pkg/runtime"
  "k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
  runtimeScheme = runtime.NewScheme()
  codecs        = serializer.NewCodecFactory(runtimeScheme)
  deserializer  = codecs.UniversalDeserializer()
)

type patchOperation struct {
  Op    string `json:"op"`
  Path  string `json:"path"`
  Value interface{} `json:"value,omitempty"`
}

func init() {
  log.Println("You are init()")
}

func validate(ar *admission.AdmissionReview) *admission.AdmissionResponse {
  log.Println(ar)

  if(ar.Request == nil) {
    return &admission.AdmissionResponse{
      Allowed: false,
    }
  }

  var pod corev1.Pod
  req := ar.Request
  if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
    log.Println(err)
    return &admission.AdmissionResponse{
      Allowed: false,
    }
  }

  return &admission.AdmissionResponse{
    Allowed: true,
  }
}

func handleValidate(w http.ResponseWriter, r *http.Request) {

  body, err := ioutil.ReadAll(r.Body)
  defer r.Body.Close()
  if err != nil {
    log.Println(err)
    http.Error(w, fmt.Sprintf("could not read body: %v", err), http.StatusInternalServerError)
  }

  admissionReview  := admission.AdmissionReview{}
  if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
    log.Println(err)
    http.Error(w, fmt.Sprintf("could not decode body: %v", err), http.StatusInternalServerError)
  }

  admissionResponse := mutate(&admissionReview)
  review := admission.AdmissionReview{}
  if(admissionResponse != nil){
    review.Response = admissionResponse
    if(admissionReview.Request != nil){
        review.Response.UID = admissionReview.Request.UID
    }
  }

  body, err = json.Marshal(review)
  if err != nil {
    log.Println(err)
    http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
  }else {
    w.WriteHeader(http.StatusOK)
  }

  w.Write(body)
}

func mutate(ar *admission.AdmissionReview) *admission.AdmissionResponse {
  log.Println(ar)

  if(ar.Request == nil) {
    return &admission.AdmissionResponse{
      Allowed: false,
    }
  }

  var pod corev1.Pod
  req := ar.Request
  if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
    log.Println(err)
    return &admission.AdmissionResponse{
      Allowed: false,
    }
  }

  patches := []patchOperation{}
  patches  = append(patches,patchOperation{Op: "add", Path: "/metadata/annotations", Value: map[string]string{"sidecar-injector":"enabled"}, })
  patchBytes,err := json.Marshal(patches)
  if(err != nil) {
    log.Println(err)
    return &admission.AdmissionResponse{
      Allowed: false,
    }
  }

  return &admission.AdmissionResponse{
    Allowed: true,
    Patch:   patchBytes,
    PatchType: func() *admission.PatchType {
      pt := admission.PatchTypeJSONPatch
      return &pt
    }(),
  }

}

func handleMutate(w http.ResponseWriter, r *http.Request) {

  body, err := ioutil.ReadAll(r.Body)
  defer r.Body.Close()
  if err != nil {
    log.Println(err)
    http.Error(w, fmt.Sprintf("could not read body: %v", err), http.StatusInternalServerError)
  }

  admissionReview  := admission.AdmissionReview{}
  if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
    log.Println(err)
    http.Error(w, fmt.Sprintf("could not decode body: %v", err), http.StatusInternalServerError)
  }

  admissionResponse := mutate(&admissionReview)
  review := admission.AdmissionReview{}
  if(admissionResponse != nil){
    review.Response = admissionResponse
    if(admissionReview.Request != nil){
        review.Response.UID = admissionReview.Request.UID
    }
  }

  body, err = json.Marshal(review)
  if err != nil {
    log.Println(err)
    http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
  }else {
    w.WriteHeader(http.StatusOK)
  }

  w.Write(body)
}

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/validate", handleValidate)
  mux.HandleFunc("/mutate", handleMutate)

  srvr := &http.Server{
    Addr:           ":8443",
    Handler:        mux,
    ReadTimeout:    10 * time.Second,
    WriteTimeout:   10 * time.Second,
    MaxHeaderBytes: 1 << 20,
  }

  log.Fatal(srvr.ListenAndServeTLS("./ssl/webhook.crt", "./ssl/webhook.key"))
}
