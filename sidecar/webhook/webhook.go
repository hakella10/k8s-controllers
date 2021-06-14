package main

import (
  "fmt"
  "html"
  "io/ioutil"
  "log"
  "net/http"
  "time"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}

func handleMutate(w http.ResponseWriter, r *http.Request) {

    body, err := ioutil.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintf(w, "%s", err)
    }

    log.Printf("Body is %b",body)

    w.WriteHeader(http.StatusOK)
    w.Write(body)
}

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", handleRoot)
  mux.HandleFunc("/mutate", handleMutate)

  srvr := &http.Server{
    Addr:           ":8443",
    Handler:        mux,
    ReadTimeout:    10 * time.Second,
    WriteTimeout:   10 * time.Second,
    MaxHeaderBytes: 1 << 20,
  }

  log.Println("Starting server on port 8443")
  log.Fatal(srvr.ListenAndServeTLS("/opt/crt/webhook.crt", "/opt/key/webhook.key"))
}
