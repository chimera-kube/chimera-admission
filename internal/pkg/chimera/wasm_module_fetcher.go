package chimera

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/engineerd/wasm-to-oci/pkg/oci"
	"github.com/pkg/errors"
)

type ModuleSource int

const (
	UnknownSource  ModuleSource = iota
	FileSource     ModuleSource = iota
	HTTPSource     ModuleSource = iota
	RegistrySource ModuleSource = iota
)

func WASMModuleSource(uri string) (ModuleSource, string, error) {
	parsedUri, err := url.Parse(uri)
	if err != nil {
		return UnknownSource, "", errors.Errorf("invalid source: %q", uri)
	}
	switch parsedUri.Scheme {
	case "file":
		return FileSource, parsedUri.Path, nil
	case "http", "https":
		return HTTPSource, uri, nil
	case "registry":
		parsedUri.Scheme = ""
		return RegistrySource, strings.TrimLeft(parsedUri.String(), "/"), nil
	}
	return FileSource, "", errors.Errorf("unknown scheme %q", parsedUri.Scheme)
}

func FetchRemoteWASMModule(moduleSource ModuleSource, uri string) (string, error) {
	wasmModule, err := ioutil.TempFile("", "wasm-module-*")
	if err != nil {
		return "", err
	}
	switch moduleSource {
	case HTTPSource:
		resp, err := http.Get(uri)
		if err != nil {
			return "", errors.Errorf("could not download WASM module from %q: %v", uri, err)
		}
		defer resp.Body.Close()
		if _, err := io.Copy(wasmModule, resp.Body); err != nil {
			return "", errors.Errorf("could not download WASM module from %q: %v", uri, err)
		}
	case RegistrySource:
		if err := oci.Pull(uri, wasmModule.Name(), true, true); err != nil {
			return "", err
		}
	default:
		os.Remove(wasmModule.Name())
		return "", errors.Errorf("invalid source: %q", uri)
	}
	return wasmModule.Name(), nil
}
