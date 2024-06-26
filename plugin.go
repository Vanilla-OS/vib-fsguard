// Copyright 2024, axtlos <axtlos@disroot.org>
// SPDX-License-Identifier: GPL-3.0-ONLY

package main

import (
	"C"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vanilla-os/vib/api"
)

type FsGuardModule struct {
	Name string `json:"name"`
	Type string `json:"type"`

	CustomFsGuard   bool     `json:"CustomFsGuard"`
	FsGuardLocation string   `json:"FsGuardLocation"`
	GenerateKey     bool     `json:"GenerateKey"`
	KeyPath         string   `json:"KeyPath"`
	FilelistPaths   []string `json:"FilelistPaths"`
}

var (
	FSGUARD_URL      string = "https://github.com/linux-immutability-tools/FsGuard/releases/download/v0.1.2-2/FsGuard_0.1.2-2_linux_amd64.tar.gz"
	FSGUARD_CHECKSUM string = "b4aa058e4c4828ac57335e8cabd6b3baeff660ff524aa71069c3f56fd0445335"
)

var (
	GENFILELIST_URL      string = "https://raw.githubusercontent.com/Vanilla-OS/vib-fsguard/3323f7c3c3f8459a64b97ad408d805edc5520c8d/genfilelist.py"
	GENFILELIST_CHECKSUM string = "22658b7246d7a38c71d0c0fa64fd073ea7867da08344365242873f003abff8c5"
)

var (
	prepCommands  []string
	mainCommands  []string
	cleanCommands []string
)

// Helper functions for tests
// TODO: move to api
func convertToCString(s string) *C.char {
	return C.CString(s)
}

func convertToGoString(s *C.char) string {
	return C.GoString(s)
}

func fetchFsGuard(module *FsGuardModule, recipe *api.Recipe) error {
	source := api.Source{URL: FSGUARD_URL, Type: "tar", Checksum: FSGUARD_CHECKSUM}
	err := api.DownloadSource(recipe.DownloadsPath, source, module.Name)
	if err != nil {
		return err
	}
	err = api.MoveSource(recipe.DownloadsPath, recipe.SourcesPath, source, module.Name)
	return err
}

func fetchFileListScript(module *FsGuardModule, recipe *api.Recipe) error {
	source := api.Source{URL: GENFILELIST_URL, Type: "single", Checksum: GENFILELIST_CHECKSUM}
	api.DownloadTarSource(recipe.DownloadsPath, source, module.Name)
	err := os.MkdirAll(filepath.Join(recipe.SourcesPath, module.Name), 0o777)
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(recipe.DownloadsPath, module.Name+".tar"), filepath.Join(recipe.SourcesPath, module.Name, "genfilelist.py"))
	if err != nil {
		return err
	}
	prepCommands = append(prepCommands, "mkdir /FsGuard")
	prepCommands = append(prepCommands, fmt.Sprintf("chmod +x /sources/%s/genfilelist.py", module.Name))
	return nil
}

func signFileList(module *FsGuardModule) {
	fmt.Println("In signFileList")
	mainCommands = append(mainCommands, fmt.Sprintf("minisign -Sm /FsGuard/filelist -p %s/minisign.pub -s %s/minisign.key", module.KeyPath, module.KeyPath))
	mainCommands = append(mainCommands, "touch /FsGuard/signature")
	mainCommands = append(mainCommands, "echo -n \"----begin attach----\" >> /FsGuard/signature")
	mainCommands = append(mainCommands, "cat /FsGuard/filelist.minisig >> /FsGuard/signature")
	mainCommands = append(mainCommands, "echo -n \"----begin second attach----\" >> /FsGuard/signature")
	mainCommands = append(mainCommands, fmt.Sprintf("tail -n1 %s/minisign.pub >> /FsGuard/signature", module.KeyPath))
	mainCommands = append(mainCommands, "cat /FsGuard/signature >> /sources/FsGuard")
}

//export BuildModule
func BuildModule(moduleInterface *C.char, recipeInterface *C.char) *C.char {
	var module *FsGuardModule
	var recipe *api.Recipe

	// Clean command arrays for testcases
	prepCommands = []string{}
	mainCommands = []string{}
	cleanCommands = []string{}

	err := json.Unmarshal([]byte(C.GoString(moduleInterface)), &module)
	if err != nil {
		return C.CString(fmt.Sprintf("ERROR: %s", err.Error()))
	}

	err = json.Unmarshal([]byte(C.GoString(recipeInterface)), &recipe)
	if err != nil {
		return C.CString(fmt.Sprintf("ERROR: %s", err.Error()))
	}

	err = fetchFsGuard(module, recipe)
	if err != nil {
		return C.CString(fmt.Sprintf("ERROR: %s", err.Error()))
	}
	err = fetchFileListScript(module, recipe)
	if err != nil {
		return C.CString(fmt.Sprintf("ERROR: %s", err.Error()))
	}

	cleanCommands = append(cleanCommands, fmt.Sprintf("mv /sources/FsGuard %s", module.FsGuardLocation))

	if module.GenerateKey {
		prepCommands = append(prepCommands, "minisign -WG -s ./minisign.key")
		cleanCommands = append(cleanCommands, "rm ./minisign.key ./minisign.pub")
		module.KeyPath = "./"
	} else if len(strings.TrimSpace(module.KeyPath)) == 0 {
		return C.CString("ERROR: Keypath not specified and GenerateKey set to false. Cannot proceed")
	}

	for _, listDirectories := range module.FilelistPaths {
		mainCommands = append(mainCommands, fmt.Sprintf("python3 /sources/%s/genfilelist.py %s /FsGuard/filelist %s", module.Name, listDirectories, module.FsGuardLocation))
	}

	signFileList(module)

	cmd := append(append(prepCommands, mainCommands...), cleanCommands...)

	return C.CString(strings.Join(cmd, " && "))
}

func main() {}
