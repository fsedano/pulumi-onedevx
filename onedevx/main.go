package main

import (
	"fmt"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// Create and set NS
		stack := ctx.Stack()
		ns := fmt.Sprintf("onedevx-%s", stack)
		pns, err := corev1.NewNamespace(ctx, ns, &corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(ns),
			},
		})
		if err != nil {
			return err
		}
		err = walkWorkspecs(ctx, ns, "reference/workspecs")
		if err != nil {
			return err
		}
		ctx.Export("ns", pns.Metadata.Name())
		return nil
	})
}
