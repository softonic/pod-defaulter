package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/softonic/pod-defaulter/pkg/admission"
	h "github.com/softonic/pod-defaulter/pkg/http"
	"github.com/softonic/pod-defaulter/pkg/version"
	"k8s.io/klog"
	"net/http"
	"os"
)

type params struct {
	version     bool
	certificate string
	privateKey  string
	LogLevel    int
}

const DEFAULT_BIND_ADDRESS = ":8443"

var handler *h.HttpHandler

func init() {
	// Read ConfigMap
	cm := make(map[string]interface{})
	handler = getHttpHandler(cm)
	klog.InitFlags(nil)
}

func main() {
	var params params

	if len(os.Args) < 2 {
		klog.Fatalf("Minimum arguments are 2")
		os.Exit(1)
	}

	flag.StringVar(&params.certificate, "tls-cert", "bar", "a string var")
	flag.StringVar(&params.privateKey, "tls-key", "bar", "a string var")

	flag.Parse()

	if params.version {
		fmt.Println("Version:", version.Version)
	} else {
		run(&params)
	}

}

func run(params *params) {
	mux := http.NewServeMux()

	_, err := tls.LoadX509KeyPair(params.certificate, params.privateKey)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
	}

	mux.HandleFunc("/mutate", func(w http.ResponseWriter, r *http.Request) {
		handler.MutationHandler(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler.HealthCheckHandler(w, r)
	})
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	address := os.Getenv("BIND_ADDRESS")
	if address == "" {
		address = DEFAULT_BIND_ADDRESS
	}
	klog.V(2).Infof("Starting server, bound at %v", address)
	klog.Infof("Listening to address %v", address)
	srv := &http.Server{
		Addr:         address,
		Handler:      mux,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	klog.Fatalf("Could not start server: %v", srv.ListenAndServeTLS(params.certificate, params.privateKey))
}

func getHttpHandler(cm map[string]interface{}) *h.HttpHandler {
	return h.NewHttpHanlder(getPodDefaultValuesAdmissionReviewer(cm))
}

func getPodDefaultValuesAdmissionReviewer(cm map[string]interface{}) *admission.AdmissionReviewer {
	return admission.NewPodDefaultValuesAdmissionReviewer(cm)
}
