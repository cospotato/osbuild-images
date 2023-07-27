package scos

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/osbuild/images/internal/common"
	"github.com/osbuild/images/internal/oscap"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/platform"
	"github.com/osbuild/images/pkg/rpmmd"
	"github.com/osbuild/images/pkg/runner"
)

const (
	// package set names

	// main/common os image package set name
	osPkgsKey = "os"

	// container package set name
	containerPkgsKey = "container"

	// installer package set name
	installerPkgsKey = "installer"

	// blueprint package set name
	blueprintPkgsKey = "blueprint"

	//Kernel options for ami, qcow2, openstack, vhd and vmdk types
	defaultKernelOptions = "ro no_timer_check console=ttyS0,115200n8 biosdevname=0 net.ifnames=0"
)

var (
	oscapProfileAllowList = []oscap.Profile{
		oscap.Ospp,
		oscap.PciDss,
		oscap.Standard,
	}

	// Services
	scosServices = []string{
		"sshd.service",
		"containerd.service",
		"docker.service",
	}

	// ImageType

	scosRockyCommitImgType = imageType{
		name:        "ostree-commit",
		nameAliases: []string{"scos-rocky-commit"},
		filename:    "commit.tar",
		mimeType:    "application/x-tar",
		packageSets: map[string]packageSetFunc{
			osPkgsKey: rockyCommitPackageSet,
		},
		defaultImageConfig: &distro.ImageConfig{
			EnabledServices: scosServices,
		},
		rpmOstree:        true,
		image:            iotCommitImage,
		buildPipelines:   []string{"build"},
		payloadPipelines: []string{"os", "ostree-commit", "commit-archive"},
		exports:          []string{"commit-archive"},
	}

	scosRockyOCIImgType = imageType{
		name:        "ostree-container",
		nameAliases: []string{"scos-rocky-container"},
		filename:    "container.tar",
		mimeType:    "application/x-tar",
		packageSets: map[string]packageSetFunc{
			osPkgsKey: rockyCommitPackageSet,
			containerPkgsKey: func(t *imageType) rpmmd.PackageSet {
				return rpmmd.PackageSet{}
			},
		},
		defaultImageConfig: &distro.ImageConfig{
			EnabledServices: scosServices,
		},
		rpmOstree:        true,
		bootISO:          false,
		image:            iotContainerImage,
		buildPipelines:   []string{"build"},
		payloadPipelines: []string{"os", "ostree-commit", "container-tree", "container"},
		exports:          []string{"container"},
	}

	scosOECommitImgType = imageType{
		name:        "ostree-commit",
		nameAliases: []string{"scos-oe-commit"},
		filename:    "commit.tar",
		mimeType:    "application/x-tar",
		packageSets: map[string]packageSetFunc{
			osPkgsKey: rockyCommitPackageSet,
		},
		defaultImageConfig: &distro.ImageConfig{
			EnabledServices: scosServices,
		},
		rpmOstree:        true,
		image:            iotCommitImage,
		buildPipelines:   []string{"build"},
		payloadPipelines: []string{"os", "ostree-commit", "commit-archive"},
		exports:          []string{"commit-archive"},
	}

	scosOEOCIImgType = imageType{
		name:        "ostree-container",
		nameAliases: []string{"scos-oe-container"},
		filename:    "container.tar",
		mimeType:    "application/x-tar",
		packageSets: map[string]packageSetFunc{
			osPkgsKey: rockyCommitPackageSet,
			containerPkgsKey: func(t *imageType) rpmmd.PackageSet {
				return rpmmd.PackageSet{}
			},
		},
		defaultImageConfig: &distro.ImageConfig{
			EnabledServices: scosServices,
		},
		rpmOstree:        true,
		bootISO:          false,
		image:            iotContainerImage,
		buildPipelines:   []string{"build"},
		payloadPipelines: []string{"os", "ostree-commit", "container-tree", "container"},
		exports:          []string{"container"},
	}
)

type distribution struct {
	name               string
	product            string
	osVersion          string
	releaseVersion     string
	modulePlatformID   string
	ostreeRefTmpl      string
	isolabelTmpl       string
	runner             runner.Runner
	arches             map[string]distro.Arch
	defaultImageConfig *distro.ImageConfig
}

// Fedora based OS image configuration defaults
var defaultDistroImageConfig = &distro.ImageConfig{
	Timezone: common.ToPtr("UTC"),
	Locale:   common.ToPtr("en_US"),
}

func getDistro(base string, version int) distribution {
	switch base {
	default:
		fallthrough
	case "rocky":
		return distribution{
			name:               fmt.Sprintf("scos-rocky-%d", version),
			product:            "SCOS",
			osVersion:          strconv.Itoa(version),
			releaseVersion:     strconv.Itoa(version),
			modulePlatformID:   fmt.Sprintf("platform:el%d", version),
			ostreeRefTmpl:      fmt.Sprintf("scos/rocky%d/%%s/os", version),
			isolabelTmpl:       fmt.Sprintf("SCOS-Rocky%d-BaseOS-%%s", version),
			runner:             &runner.CentOS{Version: 8},
			defaultImageConfig: defaultDistroImageConfig,
		}
	case "oe":
		return distribution{
			name:               fmt.Sprintf("scos-oe-%d", version),
			product:            "SCOS",
			osVersion:          strconv.Itoa(version),
			releaseVersion:     strconv.Itoa(version),
			modulePlatformID:   fmt.Sprintf("platform:oe%d", version),
			ostreeRefTmpl:      fmt.Sprintf("scos/oe%d/%%s/os", version),
			isolabelTmpl:       fmt.Sprintf("SCOS-OpenEuler%d-BaseOS-%%s", version),
			runner:             &runner.CentOS{Version: 8},
			defaultImageConfig: defaultDistroImageConfig,
		}
	}
}

func (d *distribution) Name() string {
	return d.name
}

func (d *distribution) Releasever() string {
	return d.releaseVersion
}

func (d *distribution) ModulePlatformID() string {
	return d.modulePlatformID
}

func (d *distribution) OSTreeRef() string {
	return d.ostreeRefTmpl
}

func (d *distribution) ListArches() []string {
	archNames := make([]string, 0, len(d.arches))
	for name := range d.arches {
		archNames = append(archNames, name)
	}
	sort.Strings(archNames)
	return archNames
}

func (d *distribution) GetArch(name string) (distro.Arch, error) {
	arch, exists := d.arches[name]
	if !exists {
		return nil, errors.New("invalid architecture: " + name)
	}
	return arch, nil
}

func (d *distribution) addArches(arches ...architecture) {
	if d.arches == nil {
		d.arches = map[string]distro.Arch{}
	}

	// Do not make copies of architectures, as opposed to image types,
	// because architecture definitions are not used by more than a single
	// distro definition.
	for idx := range arches {
		d.arches[arches[idx].name] = &arches[idx]
	}
}

func (d *distribution) getDefaultImageConfig() *distro.ImageConfig {
	return d.defaultImageConfig
}

// --- Architecture ---

type architecture struct {
	distro           *distribution
	name             string
	imageTypes       map[string]distro.ImageType
	imageTypeAliases map[string]string
}

func (a *architecture) Name() string {
	return a.name
}

func (a *architecture) ListImageTypes() []string {
	itNames := make([]string, 0, len(a.imageTypes))
	for name := range a.imageTypes {
		itNames = append(itNames, name)
	}
	sort.Strings(itNames)
	return itNames
}

func (a *architecture) GetImageType(name string) (distro.ImageType, error) {
	t, exists := a.imageTypes[name]
	if !exists {
		aliasForName, exists := a.imageTypeAliases[name]
		if !exists {
			return nil, errors.New("invalid image type: " + name)
		}
		t, exists = a.imageTypes[aliasForName]
		if !exists {
			panic(fmt.Sprintf("image type '%s' is an alias to a non-existing image type '%s'", name, aliasForName))
		}
	}
	return t, nil
}

func (a *architecture) addImageTypes(platform platform.Platform, imageTypes ...imageType) {
	if a.imageTypes == nil {
		a.imageTypes = map[string]distro.ImageType{}
	}
	for idx := range imageTypes {
		it := imageTypes[idx]
		it.arch = a
		it.platform = platform
		a.imageTypes[it.name] = &it
		for _, alias := range it.nameAliases {
			if a.imageTypeAliases == nil {
				a.imageTypeAliases = map[string]string{}
			}
			if existingAliasFor, exists := a.imageTypeAliases[alias]; exists {
				panic(fmt.Sprintf("image type alias '%s' for '%s' is already defined for another image type '%s'", alias, it.name, existingAliasFor))
			}
			a.imageTypeAliases[alias] = it.name
		}
	}
}

func (a *architecture) Distro() distro.Distro {
	return a.distro
}

func NewRocky8() distro.Distro {
	return newRockyDistro("rocky", 8)
}

func NewOE1() distro.Distro {
	return newOEDistro("oe", 1)
}

func newRockyDistro(base string, version int) distro.Distro {
	rd := getDistro(base, version)

	// Architecture definitions
	x86_64 := architecture{
		name:   platform.ARCH_X86_64.String(),
		distro: &rd,
	}

	aarch64 := architecture{
		name:   platform.ARCH_AARCH64.String(),
		distro: &rd,
	}

	x86_64.addImageTypes(
		&platform.X86{
			BIOS:       true,
			UEFIVendor: "smartx",
		},
		scosRockyCommitImgType,
		scosRockyOCIImgType,
	)

	aarch64.addImageTypes(
		&platform.Aarch64{
			UEFIVendor: "smartx",
		},
		scosRockyCommitImgType,
		scosRockyOCIImgType,
	)

	rd.addArches(x86_64, aarch64)

	return &rd
}

func newOEDistro(base string, version int) distro.Distro {
	rd := getDistro(base, version)

	// Architecture definitions
	x86_64 := architecture{
		name:   platform.ARCH_X86_64.String(),
		distro: &rd,
	}

	aarch64 := architecture{
		name:   platform.ARCH_AARCH64.String(),
		distro: &rd,
	}

	x86_64.addImageTypes(
		&platform.X86{
			BIOS:       true,
			UEFIVendor: "smartx",
		},
		scosOECommitImgType,
		scosOEOCIImgType,
	)

	aarch64.addImageTypes(
		&platform.Aarch64{
			UEFIVendor: "smartx",
		},
		scosOECommitImgType,
		scosOEOCIImgType,
	)

	rd.addArches(x86_64, aarch64)

	return &rd
}
