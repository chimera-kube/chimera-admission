package chimera

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime"
	"sync"

	"github.com/pkg/errors"
)

type WasmWorker struct {
	stdinWasmPath  string
	stdoutWasmPath string
	stack          *WasmStack
	mutex          sync.Mutex
}

func finalizer(w *WasmWorker) {
	os.Remove(w.stdinWasmPath)
	os.Remove(w.stdoutWasmPath)
}

func NewWasmWorker(pathToWasmModule string, envKeys, envValues []string) (*WasmWorker, error) {
	stdinWasm, err := ioutil.TempFile("", "wasm-stdin-*")
	if err != nil {
		return nil, err
	}

	stdoutWasm, err := ioutil.TempFile("", "wasm-stdout-*")
	if err != nil {
		os.Remove(stdinWasm.Name())
		return nil, err
	}

	worker := &WasmWorker{
		stdinWasmPath:  stdinWasm.Name(),
		stdoutWasmPath: stdoutWasm.Name(),
	}
	runtime.SetFinalizer(worker, finalizer)

	stack, err := NewWasmStack(
		pathToWasmModule,
		worker.stdinWasmPath,
		worker.stdoutWasmPath,
		envKeys,
		envValues)
	if err != nil {
		os.Remove(stdinWasm.Name())
		os.Remove(stdoutWasm.Name())

		return nil, errors.Wrap(err, "Cannot initialize WASM stack")
	}
	worker.stack = stack

	return worker, nil
}

func (w *WasmWorker) ProcessRequest(request []byte) (ValidationResponse, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	response := ValidationResponse{}

	// setup WASM stdin
	if err := os.Truncate(w.stdinWasmPath, 0); err != nil {
		return response, errors.Wrap(err, "Cannot truncate WASM stdin file")
	}
	if err := ioutil.WriteFile(w.stdinWasmPath, request, 0400); err != nil {
		return response, errors.Wrap(err, "Cannot populate WASM stdin file")
	}

	// setup WASM stdout
	if err := os.Truncate(w.stdoutWasmPath, 0); err != nil {
		return response, errors.Wrap(err, "Cannot truncate WASM stdin file")
	}

	if err := w.stack.Run(); err != nil {
		return response, errors.Wrap(err, "Cannot run the WASM code")
	}

	stdout, err := ioutil.ReadFile(w.stdoutWasmPath)
	if err != nil {
		return response, errors.Wrap(err, "Cannot read WASM stdout")
	}

	if err := json.Unmarshal(stdout, &response); err != nil {
		return response, errors.Wrap(err, "Cannot decode WASM code response")
	}

	return response, nil
}
