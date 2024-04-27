package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"
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

		// Walk all directories for components
		err = filepath.Walk("reference/components", func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				if info.Name() == "component.yaml" {
					err := installComponent(ctx, ns, path)
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		ctx.Export("ns", pns.Metadata.Name())
		return nil
	})
}

func installComponent(ctx *pulumi.Context, ns string, componentPath string) error {
	// Get component
	yamlFile, err := os.ReadFile(componentPath)
	if err != nil {
		return err
	}
	component := OneDevxCRD{}
	err = yaml.Unmarshal(yamlFile, &component)
	if err != nil {
		return err
	}
	switch component.Spec.ComponentType {
	case "helm":
		deployHelmComponent(ctx, ns, component)
	case "image":
		deployImageComponent(ctx, ns, component)
	default:
		return fmt.Errorf("error on component type: %s", component.Spec.ComponentType)
	}
	//ctx.Export("name", deployment.Metadata.Name())
	//ctx.Export("component", pulumi.String(component.Metadata.Name))
	return nil
}

func deployHelmComponent(ctx *pulumi.Context, ns string, component OneDevxCRD) error {
	ctx.Log.Warn("helm is unsupported", nil)
	return fmt.Errorf("helm unsupported")
}

func deployImageComponent(ctx *pulumi.Context, ns string, component OneDevxCRD) error {
	appLabels := pulumi.StringMap{
		"onedevxComponent": pulumi.String(component.Metadata.Name),
	}
	_, err := appsv1.NewDeployment(ctx, fmt.Sprintf("onedevx-%s", component.Metadata.Name), &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: pulumi.String(ns),
		},
		Spec: appsv1.DeploymentSpecArgs{
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: appLabels,
			},
			Replicas: pulumi.Int(1),
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: appLabels,
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						corev1.ContainerArgs{
							Name:  pulumi.String(component.Metadata.Name),
							Image: pulumi.String(component.Spec.ImageInfo.ImageName),
						}},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
