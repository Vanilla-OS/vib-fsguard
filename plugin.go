package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/vanilla-os/vib/api"
)

type FsGuardModule struct {
	Name string `json:"name"`
	Type string `json:"type"`

	FsGuardLocation string   `json:"fsguardlocation`
	GenerateKey     bool     `json:"genkey"`
	KeyPath         string   `json:"keypath"`
	FilelistPaths   []string `json:"filelistpaths"`
}

var FSGUARD_URL string = "https://github.com/linux-immutability-tools/FsGuard/releases/download/v0.1.2/FsGuard_0.1.2_linux_amd64.tar.gz"
var FSGUARD_CHECKSUM string = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

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

func writeFilelistScript(module FsGuardModule, recipe *api.Recipe) error {
	var script string
	script = `while [ $# -gt 0 ]; do
    BASEPATH="$1"
    for f in $(ls -1 $BASEPATH); do
        echo "$BASEPATH/$f $(sha1sum $BASEPATH/$f | sed 's/ .*//g') $(ls -al $BASEPATH/$f | awk 'BEGIN{FS=" "}; {print $1};' | grep s > /dev/null && echo "true" || echo "false")" >> /FsGuard/filelist
    done
    shift
done`
	prepCommands = append(prepCommands, "mkdir /FsGuard")
	prepCommands = append(prepCommands, fmt.Sprintf("chmod +x /sources/%s/gen_filelist", module.Name))
	os.MkdirAll(recipe.SourcesPath+module.Name, 0666)
	err := os.WriteFile(recipe.SourcesPath+module.Name+"gen_filelist", []byte(script), 0666)
	return err
}

func signFileList(module FsGuardModule) {
	mainCommands = append(mainCommands, fmt.Sprintf("minisign -Sm /FsGuard/filelist -p %s/minisign.pub -s %s/minisign.key", module.KeyPath, module.KeyPath))
	mainCommands = append(mainCommands, "touch /FsGuard/signature")
	mainCommands = append(mainCommands, "echo -n \"----begin attach----\" >> /FsGuard/signature")
	mainCommands = append(mainCommands, "cat /FsGuard/filelist.minisig >> /FsGuard/signature")
	mainCommands = append(mainCommands, " echo -n \"----begin second attach----\" >> /FsGuard/signature")
	mainCommands = append(mainCommands, fmt.Sprintf("tail -n1 %s/minisign.pub >> /FsGuard/signature", module.KeyPath))
	mainCommands = append(mainCommands, "cat /FsGuard/signature >> /usr/bin/FsGuard")
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
	err = writeFilelistScript(module, recipe)
	if err != nil {
		return "", err
	}

	cleanCommands = append(cleanCommands, fmt.Sprintf("mv /sources/%s/FsGuard %s", module.Name, module.FsGuardLocation))

	if module.GenerateKey {
		prepCommands = append(prepCommands, "minisign -WG -s ./minisign.sec")
		cleanCommands = append(cleanCommands, "rm ./minisign.sec ./minisign.pub")
		module.KeyPath = "./"
	} else if len(strings.TrimSpace(module.KeyPath)) == 0 {
		return "", fmt.Errorf("Keypath not specified and GenerateKey set to false. Cannot proceed")
	}

	for _, listDirectories := range module.FilelistPaths {
		mainCommands = append(mainCommands, fmt.Sprintf("/sources/%s/gen_filelist %s", module.Name, listDirectories))
	}

	signFileList(module)

	cmd := append(append(prepCommands, mainCommands...), cleanCommands...)

	return strings.Join(cmd[:], " && "), nil
}
