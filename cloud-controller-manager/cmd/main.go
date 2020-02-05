// Copyright 2019 inspur Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"k8s.io/apiserver/pkg/server/healthz"
	"math/rand"
	"net/http"
	"os"
	"time"

	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app"

	// NOTE: Importing all in-tree cloud-providers is not required when
	// implementing an out-of-tree cloud-provider.
	_ "gitserver/kubernetes/inspur-cloud-controller-manager/pkg" //pre load loadbalance incloud init
	_ "k8s.io/kubernetes/pkg/util/prometheusclientgo"            // load all the prometheus client-go plugins
	_ "k8s.io/kubernetes/pkg/version/prometheus"                 // for version metric registration
)

func init() {
	healthz.InstallHandler(http.DefaultServeMux)
}
func main() {
	rand.Seed(time.Now().UnixNano())

	command := app.NewCloudControllerManagerCommand()
	// TODO: once we switch everything over to Cobra commands, we can go back to calling
	// utilflag.InitFlags() (by removing its pflag.Parse() call). For now, we have to set the
	// normalize func and add the go flag set by hand.
	// utilflag.InitFlags()

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
