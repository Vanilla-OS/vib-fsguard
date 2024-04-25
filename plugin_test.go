package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type testCases struct {
	module   interface{}
	expected string
}

var test = []testCases{
	// No custom, location /fsguard, generate key, keypath /fsguard/keys, filelists /bin
	{
		FsGuardModule{Name: "testCase1", CustomFsGuard: false, FsGuardLocation: "/fsguard", GenerateKey: true, KeyPath: "/fsguard/keys", FilelistPaths: []string{"/bin"}},
		"mkdir /FsGuard && chmod +x /sources/testCase1/genfilelist.py && minisign -WG -s ./minisign.key && python3 /sources/testCase1/genfilelist.py /bin /FsGuard/filelist /fsguard && minisign -Sm /FsGuard/filelist -p .//minisign.pub -s .//minisign.key && touch /FsGuard/signature && echo -n \"----begin attach----\" >> /FsGuard/signature && cat /FsGuard/filelist.minisig >> /FsGuard/signature && echo -n \"----begin second attach----\" >> /FsGuard/signature && tail -n1 .//minisign.pub >> /FsGuard/signature && cat /FsGuard/signature >> /sources/FsGuard && mv /sources/FsGuard /fsguard && rm ./minisign.key ./minisign.pub",
	},

	// With custom, location /fsguard, dont generate key, keypath /fsguard/keys, filelists /bin
	{
		FsGuardModule{Name: "testCase2", CustomFsGuard: true, FsGuardLocation: "/fsguard", GenerateKey: false, KeyPath: "/fsguard/keys", FilelistPaths: []string{"/bin"}},
		"mkdir /FsGuard && chmod +x /sources/testCase2/genfilelist.py && python3 /sources/testCase2/genfilelist.py /bin /FsGuard/filelist /fsguard && minisign -Sm /FsGuard/filelist -p /fsguard/keys/minisign.pub -s /fsguard/keys/minisign.key && touch /FsGuard/signature && echo -n \"----begin attach----\" >> /FsGuard/signature && cat /FsGuard/filelist.minisig >> /FsGuard/signature && echo -n \"----begin second attach----\" >> /FsGuard/signature && tail -n1 /fsguard/keys/minisign.pub >> /FsGuard/signature && cat /FsGuard/signature >> /sources/FsGuard && mv /sources/FsGuard /fsguard",
	},
}

var recipe = "{\"Name\":\"fsguard unit test\",\"Id\":\"fsguard\",\"Stages\":[{\"id\":\"test\",\"base\":\"test:latest\",\"singlelayer\":false,\"copy\":null,\"labels\":null,\"env\":null,\"adds\":null,\"args\":null,\"runs\":null,\"expose\":null,\"cmd\":null,\"modules\":[{}],\"Entrypoint\":null}],\"Path\":\"/fakepath/recipe.yml\",\"ParentPath\":\"/fakepath\",\"DownloadsPath\":\"/tmp/fsguard/downloads\",\"SourcesPath\":\"/tmp/fsguard/sources\",\"PluginPath\":\"/plugins\",\"Containerfile\":\"/Containerfile\"}"

func TestBuildModule(t *testing.T) {
	err := os.MkdirAll("/tmp/fsguard/downloads", 0777)
	if err != nil {
		t.Errorf("%s", err)
	}
	err = os.MkdirAll("/tmp/fsguard/sources", 0777)
	if err != nil {
		t.Errorf("%s", err)
	}
	for i, testCase := range test {
		err = os.RemoveAll("/tmp/fsguard/downloads/*")
		if err != nil {
			t.Errorf("%s", err)
		}
		err = os.RemoveAll("/tmp/fsguard/sources/*")
		if err != nil {
			t.Errorf("%s", err)
		}
		output := convertToCString("")
		moduleInterface, err := json.Marshal(testCase.module)
		if err != nil {
			t.Errorf("Error in json %s", err.Error())
		}
		output = BuildModule(convertToCString(string(moduleInterface)), convertToCString(recipe))
		if convertToGoString(output) != testCase.expected {
			t.Errorf("Output %s not equivalent to expected %s", convertToGoString(output), testCase.expected)
		} else {
			fmt.Printf("-- Testcase %d succeeded --\n", i)
		}
	}
	os.RemoveAll("/tmp/fsguard")

}
