package scos

import "github.com/osbuild/images/pkg/rpmmd"

func rockyCommitPackageSet(t *imageType) rpmmd.PackageSet {
	return rpmmd.PackageSet{
		Include: []string{
			"rocky-release", "basesystem", "network-scripts", "kernel",
			"glibc", "tmux", "nss-altfiles", "glibc-minimal-langpack",
			"lvm2", "cryptsetup", "dracut", "dracut-config-generic",
			"bash", "bash-completion", "crontabs", "logrotate",
			"coreutils", "which", "curl", "wget", "openssl", "jq",
			"hostname", "iproute", "iputils", "iptables",
			"openssh-clients", "openssh-server", "passwd",
			"dnsmasq", "traceroute", "tcpdump", "net-tools",
			"tar", "gzip", "xz", "man", "polkit",
			"e2fsprogs", "xfsprogs", "dosfstools",
			"sudo", "systemd", "util-linux", "vim-minimal",
			"setools-console", "kernel-tools",
			"setup", "shadow-utils", "attr", "audit",
			"policycoreutils", "selinux-policy-targeted",
			"procps-ng", "rpm", "rpm-ostree",
			"keyutils", "cracklib-dicts",
			"gnupg2", "pinentry", "cloud-init",
			"grub2", "grub2-efi-x64", "efibootmgr", "shim-x64",
			"containerd.io", "docker-ce", "docker-compose-plugin",
		},
		Exclude: []string{
			"geolite2-city",
			"geolite2-country",
			"glibc-all-langpacks",
			"mozjs78",
		},
	}
}
