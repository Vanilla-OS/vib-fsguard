// Copyright 2024, axtlos <axtlos@disroot.org>
// SPDX-License-Identifier: GPL-3.0-ONLY

package main

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/vanilla-os/vib/api"
	"os"
	"path/filepath"
	"strings"
)

type FsGuardModule struct {
	Name string `json:"name"`
	Type string `json:"type"`

	CustomFsGuard   bool     `json:"customfsguard"`
	FsGuardLocation string   `json:"fsguardlocation`
	GenerateKey     bool     `json:"genkey"`
	KeyPath         string   `json:"keypath"`
	FilelistPaths   []string `json:"filelistpaths"`
}

var FSGUARD_URL string = "https://github.com/linux-immutability-tools/FsGuard/releases/download/v0.1.2-1/FsGuard_0.1.2-1_linux_amd64.tar.gz"
var FSGUARD_CHECKSUM string = "dbd71388f8591fe8dfdbdc57a004e4df02a8f495caa4081e959d6d66cd494f1e"

var GENFILELIST_URL string = "https://raw.githubusercontent.com/Vanilla-OS/vib-fsguard/main/genfilelist.py"
var GENFILELIST_CHECKSUM string = "55d575f65613a2de43344f9502734e001a89a036670f6e44e9292a6d33beeb64"

var prepCommands []string
var mainCommands []string
var cleanCommands []string

func fetchFsGuard(module FsGuardModule, recipe *api.Recipe) error {
	var source api.Source
	source = api.Source{URL: FSGUARD_URL, Type: "tar", Checksum: FSGUARD_CHECKSUM}
	err := api.DownloadSource(recipe.DownloadsPath, source, module.Name)
	if err != nil {
		return err
	}
	err = api.MoveSource(recipe.DownloadsPath, recipe.SourcesPath, source, module.Name)
	return err
}

func fetchFileListScript(module FsGuardModule, recipe *api.Recipe) error {
	var source api.Source
	source = api.Source{URL: GENFILELIST_URL, Type: "single", Checksum: GENFILELIST_CHECKSUM}
	api.DownloadTarSource(recipe.DownloadsPath, source, module.Name)
	err := os.Rename(filepath.Join(recipe.DownloadsPath, module.Name), filepath.Join(recipe.SourcesPath, module.Name))
	if err != nil {
		return err
	}
	prepCommands = append(prepCommands, "mkdir /FsGuard")
	prepCommands = append(prepCommands, fmt.Sprintf("chmod +x /sources/%s/genfilelist.py", module.Name))
	return nil
}

func signFileList(module FsGuardModule) {
	mainCommands = append(mainCommands, fmt.Sprintf("minisign -Sm /FsGuard/filelist -p %s/minisign.pub -s %s/minisign.key", module.KeyPath, module.KeyPath))
	mainCommands = append(mainCommands, "touch /FsGuard/signature")
	mainCommands = append(mainCommands, "echo -n \"----begin attach----\" >> /FsGuard/signature")
	mainCommands = append(mainCommands, "cat /FsGuard/filelist.minisig >> /FsGuard/signature")
	mainCommands = append(mainCommands, "echo -n \"----begin second attach----\" >> /FsGuard/signature")
	mainCommands = append(mainCommands, fmt.Sprintf("tail -n1 %s/minisign.pub >> /FsGuard/signature", module.KeyPath))
	mainCommands = append(mainCommands, "cat /FsGuard/signature >> /sources/FsGuard")
}

func BuildModule(moduleInterface interface{}, recipe *api.Recipe) (string, error) {
	var module FsGuardModule
	err := mapstructure.Decode(moduleInterface, &module)
	if err != nil {
		return "", err
	}
	err = fetchFsGuard(module, recipe)
	if err != nil {
		return "", err
	}
	err = fetchFileListScript(module, recipe)
	if err != nil {
		return "", err
	}

	cleanCommands = append(cleanCommands, fmt.Sprintf("mv /sources/FsGuard %s", module.FsGuardLocation))

	if module.GenerateKey {
		prepCommands = append(prepCommands, "minisign -WG -s ./minisign.key")
		cleanCommands = append(cleanCommands, "rm ./minisign.key ./minisign.pub")
		module.KeyPath = "./"
	} else if len(strings.TrimSpace(module.KeyPath)) == 0 {
		return "", fmt.Errorf("Keypath not specified and GenerateKey set to false. Cannot proceed")
	}

	for _, listDirectories := range module.FilelistPaths {
		mainCommands = append(mainCommands, fmt.Sprintf("python3 /sources/%s/genfilelist.py %s /FsGuard/filelist %s", module.Name, listDirectories, module.FsGuardLocation))
	}

	signFileList(module)

	cmd := append(append(prepCommands, mainCommands...), cleanCommands...)

	return strings.Join(cmd, " && "), nil
}
