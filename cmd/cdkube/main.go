package main

import (
	"context"
	"runtime"

	stub "github.com/inge4pres/cdkube/pkg/stub"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/sirupsen/logrus"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()
	sdk.Watch("delivery.inge.4pr.es/v1alpha1", "Pipeline", "default", 5)
	sdk.Handle(stub.NewHandler())
	sdk.Run(context.TODO())
}
