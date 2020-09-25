package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/softonic/pod-defaulter/pkg/admission"
	h "github.com/softonic/pod-defaulter/pkg/http"
	"github.com/softonic/pod-defaulter/pkg/version"

	//This supports JSON tags
	"github.com/ghodss/yaml"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	// Read ConfigMap
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespace := os.Getenv("POD_NAMESPACE")
	cmName := os.Getenv("CONFIGMAP_NAME")

	//@TODO: extract this to its own class
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{})
	if err != nil {
		klog.Fatalf("Invalid config %s/%s : %v", namespace, cmName, err)
	}
	configPodTemplate := &v1.PodTemplate{}
	err = yaml.Unmarshal([]byte(cm.Data["config"]), configPodTemplate)

	if err != nil {
		klog.Fatalf("Invalid config %v", err)
	}
	klog.Infof("Unserialized config: %v", configPodTemplate)
	handler = getHttpHandler(configPodTemplate)

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

func getHttpHandler(cm *v1.PodTemplate) *h.HttpHandler {
	return h.NewHttpHanlder(getPodDefaultValuesAdmissionReviewer(cm))
}

func getPodDefaultValuesAdmissionReviewer(cm *v1.PodTemplate) *admission.AdmissionReviewer {
	return admission.NewPodDefaultValuesAdmissionReviewer(cm)
}
