package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"
)

func walkComponents(ctx *pulumi.Context, ns string, path string) error {
	// Walk all directories for components
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
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
	return err
}

func installComponent(ctx *pulumi.Context, ns string, componentPath string) error {
	// Get component
	yamlFile, err := os.ReadFile(componentPath)
	if err != nil {
		return err
	}
	component := OneDevxComponentCRD{}
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
	return nil
}

func deployHelmComponent(ctx *pulumi.Context, ns string, component OneDevxComponentCRD) error {

	_, err := helmv3.NewRelease(ctx, component.Metadata.Name, &helmv3.ReleaseArgs{
		Chart:     pulumi.String(component.Spec.HelmInfo.ChartName),
		Namespace: pulumi.String(ns),

		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(component.Spec.HelmInfo.ChartRepo),
		},

		Version: pulumi.String(component.Spec.HelmInfo.ChartVersion),
	})
	if err != nil {
		return err
	}
	return nil

}

func deployImageComponent(ctx *pulumi.Context, ns string, component OneDevxComponentCRD) error {
	appLabels := pulumi.StringMap{
		"onedevxComponent": pulumi.String(component.Metadata.Name),
	}

	// Deployment
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

	// Service
	_, err = corev1.NewService(ctx, component.Metadata.Name, &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(component.Metadata.Name),
			Namespace: pulumi.String(ns),
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(component.Spec.RestSchema.Port),
					Protocol:   pulumi.String("TCP"),
				},
			},
			Selector: pulumi.StringMap{
				"onedevxComponent": pulumi.String(component.Metadata.Name),
			},
		},
	})

	if err != nil {
		return err
	}

	// Middleware
	_, err = apiextensions.NewCustomResource(ctx, component.Metadata.Name, &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("traefik.io/v1alpha1"),
		Kind:       pulumi.String("Middleware"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(component.Metadata.Name),
			Namespace: pulumi.String(ns),
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"stripPrefix": pulumi.Map{
					"prefixes": pulumi.StringArray{
						pulumi.String(fmt.Sprintf("/%s", component.Metadata.Name)),
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	// IngressRoute
	_, err = apiextensions.NewCustomResource(ctx, component.Metadata.Name, &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("traefik.io/v1alpha1"),
		Kind:       pulumi.String("IngressRoute"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(component.Metadata.Name),
			Namespace: pulumi.String(ns),
			Annotations: pulumi.StringMap{
				"traefik.ingress.kubernetes.io/router.middlewares": pulumi.String(fmt.Sprintf("%s-%s@kubernetescrd", ns, component.Metadata.Name)),
			},
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"entryPoints": pulumi.StringArray{
					pulumi.String("web"),
				},
				"routes": pulumi.MapArray{
					pulumi.Map{
						"kind":  pulumi.String("Rule"),
						"match": pulumi.String(fmt.Sprintf("Path(`/%s/ping`)", component.Metadata.Name)),
						"middlewares": pulumi.MapArray{
							pulumi.Map{
								"name":      pulumi.String(component.Metadata.Name),
								"namespace": pulumi.String(ns),
							},
						},
						"services": pulumi.MapArray{
							pulumi.Map{
								"kind": pulumi.String("Service"),
								"name": pulumi.String(component.Metadata.Name),
								"port": pulumi.Int(80),
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
