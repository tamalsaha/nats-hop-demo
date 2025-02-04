package main

import (
	"context"
	"fmt"
	"github.com/tamalsaha/learn-nats/natsclient"
	"time"

	"github.com/tamalsaha/learn-nats/shared"
	"github.com/tamalsaha/learn-nats/transport"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"kmodules.xyz/client-go/tools/clusterid"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	_ = clientgoscheme.AddToScheme(scheme)

	nc, err := natsclient.NewConnection(shared.NATS_URL, "")
	if err != nil {
		klog.Fatalln(err)
	}
	defer nc.Close()

	ctrl.SetLogger(klogr.New())
	cfg := ctrl.GetConfigOrDie()

	uid, err := clusterid.ClusterUID(kubernetes.NewForConfigOrDie(cfg).CoreV1().Namespaces())
	if err != nil {
		panic(err)
	}

	tr, err := cfg.TransportConfig()
	if err != nil {
		panic(err)
	}
	cfg.Transport, err = transport.New(tr, nc, "proxy."+uid, 10000*time.Second)
	if err != nil {
		panic(err)
	}

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return err
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		Opts: client.WarningHandlerOptions{
			SuppressWarnings:   false,
			AllowDuplicateLogs: false,
		},
	})
	if err != nil {
		return err
	}

	var nodes core.NodeList
	err = c.List(context.TODO(), &nodes)
	if err != nil {
		panic(err)
	}
	for _, n := range nodes.Items {
		fmt.Println(n.Name)
	}
	return nil
}
