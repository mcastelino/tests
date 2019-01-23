package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	mutatingwh "github.com/slok/kubewebhook/pkg/webhook/mutating"
)

func annotatePodMutator(_ context.Context, obj metav1.Object) (bool, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		// If not a pod just continue the mutation chain(if there is one) and don't do nothing.
		return false, nil
	}

	//We cannot really support --net=host in Kata
	if pod.Spec.HostNetwork {
		fmt.Println("hostnetwork: ", pod.GetNamespace(), pod.GetName())
		return false, nil
	}

	switch pod.GetNamespace() {
	case "rook-ceph-system", "rook-ceph":
		fmt.Println("rookie: ", pod.GetNamespace(), pod.GetName())
		return false, nil
	default:
		break
	}

	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].SecurityContext != nil {
			if *pod.Spec.Containers[i].SecurityContext.Privileged {
				fmt.Println("Privileged container: ", pod.GetNamespace(), pod.GetName())
				return false, nil
			}
		}
	}

	fmt.Println("katait: ", pod.GetNamespace(), pod.GetName())

	// Mutate our object with the required annotations.
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations["mutated"] = "true"
	pod.Annotations["mutator"] = "pod-annotate"
	pod.Annotations["io.kubernetes.cri-o.TrustedSandbox"] = "false"
	pod.Annotations["io.kubernetes.cri.untrusted-workload"] = "true"

	return false, nil
}

type config struct {
	certFile string
	keyFile  string
}

func initFlags() *config {
	cfg := &config{}

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")

	fl.Parse(os.Args[1:])
	return cfg
}

func main() {
	logger := &log.Std{Debug: true}

	cfg := initFlags()

	// Create our mutator
	mt := mutatingwh.MutatorFunc(annotatePodMutator)

	mcfg := mutatingwh.WebhookConfig{
		Name: "podAnnotate",
		Obj:  &corev1.Pod{},
	}
	wh, err := mutatingwh.NewWebhook(mcfg, mt, nil, nil, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook: %s", err)
		os.Exit(1)
	}

	// Get the handler for our webhook.
	whHandler, err := whhttp.HandlerFor(wh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook handler: %s", err)
		os.Exit(1)
	}
	logger.Infof("Listening on :8080")
	err = http.ListenAndServeTLS(":8080", cfg.certFile, cfg.keyFile, whHandler)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
