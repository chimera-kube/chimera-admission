package chimera

import (
	"io/ioutil"
	"net/url"

	"github.com/engineerd/wasm-to-oci/pkg/oci"
)

func IsWasmModuleLocal(uri string) bool {
	parsedUri, err := url.Parse(uri)
	if err != nil {
		return false
	}
	return parsedUri.Scheme == "file"
}

func FetchRemoteWasmModule(uri string) (string, error) {
	parsedUri, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	wasmModule, err := ioutil.TempFile("", "wasm-module-*")
	if err != nil {
		return "", err
	}
	if err := oci.Pull(parsedUri.String(), wasmModule.Name(), true, true); err != nil {
		return "", err
	}

	return wasmModule.Name(), nil
}
