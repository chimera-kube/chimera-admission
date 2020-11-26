package chimera

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
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

func WasmModuleSource(uri string) (ModuleSource, string, error) {
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

func FetchRemoteWasmModule(moduleSource ModuleSource, uri string, insecure, nonTLS bool, caPath string) (string, error) {
	wasmModule, err := ioutil.TempFile("", "wasm-module-*")
	if err != nil {
		return "", err
	}
	switch moduleSource {
	case HTTPSource:
		tlsConfig, err := newTLSConfig(insecure, caPath)
		if err != nil {
			os.Remove(wasmModule.Name())
			return "", err
		}
		tr := &http.Transport{TLSClientConfig: tlsConfig}
		client := http.Client{Transport: tr}
		resp, err := client.Get(uri)
		if err != nil {
			os.Remove(wasmModule.Name())
			return "", errors.Errorf("could not download Wasm module from %q: %v", uri, err)
		}
		defer resp.Body.Close()
		if _, err := io.Copy(wasmModule, resp.Body); err != nil {
			os.Remove(wasmModule.Name())
			return "", errors.Errorf("could not download Wasm module from %q: %v", uri, err)
		}
	case RegistrySource:
		if caPath != "" {
			log.Printf("WARNING: currently we don't support a custom CA when pulling from an OCI registry, switching to 'insecure' mode")
			insecure = true
		}
		if err := oci.Pull(uri, wasmModule.Name(), insecure, nonTLS); err != nil {
			os.Remove(wasmModule.Name())
			return "", err
		}
	default:
		os.Remove(wasmModule.Name())
		return "", errors.Errorf("invalid source: %q", uri)
	}
	return wasmModule.Name(), nil
}

func newTLSConfig(insecure bool, caPath string) (*tls.Config, error) {
	config := tls.Config{
		InsecureSkipVerify: insecure,
	}

	if caPath == "" {
		return &config, nil
	}

	caPEM, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	roots, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	ok := roots.AppendCertsFromPEM(caPEM)
	if !ok {
		return nil, errors.New("failed to parse CA certificate")
	}
	config.RootCAs = roots

	return &config, nil
}
