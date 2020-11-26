package cmd

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

func exportedEnvVar(envVar string) (string, string, error) {
	export := strings.SplitN(envVar, "=", 2)
	if len(export) != 2 {
		return "", "", errors.Errorf("environment variable %q cannot be splitted", envVar)
	}
	return export[0], export[1], nil
}

func computeWasmEnv() ([]string, []string) {
	wasmEnv := map[string]string{}
	// Inherited envvars with CHIMERA_EXPORT_ prefix: trim and forward to the Wasm guest
	for _, env := range os.Environ() {
		exportVarName, exportVarValue, err := exportedEnvVar(env)
		if err != nil {
			continue
		}
		exportVarNameToShare := strings.TrimPrefix(exportVarName, exportedEnvVarPrefix)
		if exportVarNameToShare != exportVarName {
			wasmEnv[exportVarNameToShare] = exportVarValue
		}
	}
	// Explicitly set envvars with (--env): set on the Wasm guest directly
	for _, env := range wasmEnvVars.Value() {
		exportVarName, exportVarValue, err := exportedEnvVar(env)
		if err != nil {
			continue
		}
		wasmEnv[exportVarName] = exportVarValue
	}

	wasmEnvKeys := []string{}
	wasmEnvValues := []string{}
	for wasmEnvKey, wasmEnvValue := range wasmEnv {
		wasmEnvKeys = append(wasmEnvKeys, wasmEnvKey)
		wasmEnvValues = append(wasmEnvValues, wasmEnvValue)
	}

	return wasmEnvKeys, wasmEnvValues
}
