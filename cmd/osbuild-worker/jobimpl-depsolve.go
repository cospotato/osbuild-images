package main

import (
	"fmt"

	"github.com/osbuild/osbuild-composer/internal/rpmmd"
	"github.com/osbuild/osbuild-composer/internal/worker"
)

type DepsolveJobImpl struct {
	RPMMD rpmmd.RPMMD
}

func (impl *DepsolveJobImpl) depsolve(packageSets map[string]rpmmd.PackageSet, repos []rpmmd.RepoConfig, modulePlatformID, arch, releasever string) (map[string][]rpmmd.PackageSpec, error) {
	packageSpecs := make(map[string][]rpmmd.PackageSpec)
	for name, packageSet := range packageSets {
		packageSpec, _, err := impl.RPMMD.Depsolve(packageSet, repos, modulePlatformID, arch, releasever)
		if err != nil {
			return nil, err
		}
		packageSpecs[name] = packageSpec
	}
	return packageSpecs, nil
}

func (impl *DepsolveJobImpl) Run(job worker.Job) error {
	var args worker.DepsolveJob
	err := job.Args(&args)
	if err != nil {
		return err
	}

	var result worker.DepsolveJobResult
	result.PackageSpecs, err = impl.depsolve(args.PackageSets, args.Repos, args.ModulePlatformID, args.Arch, args.Releasever)
	if err != nil {
		result.Error = err.Error()
	}

	err = job.Update(&result)
	if err != nil {
		return fmt.Errorf("Error reporting job result: %v", err)
	}

	return nil
}