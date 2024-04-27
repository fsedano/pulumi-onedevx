package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"
)

func walkWorkspecs(ctx *pulumi.Context, ns string, path string) error {
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if info.Name() == "workspec.yaml" {
				err := installWorkspec(ctx, ns, path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}

func installWorkspec(ctx *pulumi.Context, ns string, path string) error {

	ws := OneDevxWorkspecCRD{}
	ctx.Log.Info(fmt.Sprintf("Processing ws: %s", path), nil)
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &ws)
	if err != nil {
		return err
	}

	for _, comp := range ws.Spec.ComponentList {
		ctx.Log.Info(fmt.Sprintf("Processing component %s", comp.Path), nil)
		switch comp.Type {
		case "directory":
			err = walkComponents(ctx, ns, comp.Path, ws.Metadata.Name, ws.Metadata.Name)
			if err != nil {
				return err
			}
		default:
			ctx.Log.Warn(fmt.Sprintf("component type not supported: %s", comp.Type), nil)
		}
	}
	return nil
}
