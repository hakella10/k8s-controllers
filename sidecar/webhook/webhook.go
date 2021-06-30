package main

import (
  "log"
  "fmt"
  "net/http"
  "io/ioutil"
  "time"
  "encoding/json"
  
  corev1 "k8s.io/api/core/v1"
  "k8s.io/api/admission/v1beta1"
  "k8s.io/apimachinery/pkg/runtime"
  "k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
  runtimeScheme = runtime.NewScheme()
  codecs        = serializer.NewCodecFactory(runtimeScheme)
  deserializer  = codecs.UniversalDeserializer()
)

func init() {
  log.Println("You are init()")
}

func validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
  if(ar.Request == nil) {
    return &v1beta1.AdmissionResponse{
      Allowed: false,
    }
  }
  
  var pod corev1.Pod
  req := ar.Request
  if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
    log.Println(err)
    return &v1beta1.AdmissionResponse{
      Allowed: false,
    }
  }
  
  return &v1beta1.AdmissionResponse{
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
  
  admissionReview  := v1beta1.AdmissionReview{}
  if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
    log.Println(err)
    http.Error(w, fmt.Sprintf("could not decode body: %v", err), http.StatusInternalServerError)
  }
  
  admissionResponse := mutate(&admissionReview)
  review := v1beta1.AdmissionReview{}
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

func mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
  if(ar.Request == nil) {
    return &v1beta1.AdmissionResponse{
      Allowed: false,
    }
  }
  
  var pod corev1.Pod
  req := ar.Request
  if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
    log.Println(err)
    return &v1beta1.AdmissionResponse{
      Allowed: false,
    }
  }
  
  /*
  patchBytes,err := json.Marshal([]byte(""))
  if(err != nil) {
    log.Println(err)
    return &v1beta1.AdmissionResponse{
      Allowed: false,
    }
  }
  
  return &v1beta1.AdmissionResponse{
    Allowed: true,
    Patch:   patchBytes,
    PatchType: func() *v1beta1.PatchType {
      pt := v1beta1.PatchTypeJSONPatch
      return &pt
    }(),
  }
  */
  
  return &v1beta1.AdmissionResponse{
    Allowed: true,
  }
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
  
  body, err := ioutil.ReadAll(r.Body)
  defer r.Body.Close()
  if err != nil {
    log.Println(err)
    http.Error(w, fmt.Sprintf("could not read body: %v", err), http.StatusInternalServerError)
  }
  
  admissionReview  := v1beta1.AdmissionReview{}
  if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
    log.Println(err)
    http.Error(w, fmt.Sprintf("could not decode body: %v", err), http.StatusInternalServerError)
  }
  
  admissionResponse := mutate(&admissionReview)
  review := v1beta1.AdmissionReview{}
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
