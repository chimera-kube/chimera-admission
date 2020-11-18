package chimera

import (
	"github.com/bytecodealliance/wasmtime-go"
)

type WasmStack struct {
	engine     *wasmtime.Engine
	wasmModule *wasmtime.Module
	stdin      string
	stdout     string
	envKeys    []string
	envValues  []string
}

func NewWasmStack(pathToWasmModule string, stdin, stdout string, envKeys, envValues []string) (*WasmStack, error) {
	engine := wasmtime.NewEngine()
	wasmModule, err := wasmtime.NewModuleFromFile(engine, pathToWasmModule)
	if err != nil {
		return nil, err
	}

	stack := &WasmStack{
		engine:     engine,
		wasmModule: wasmModule,
		stdin:      stdin,
		stdout:     stdout,
		envKeys:    envKeys,
		envValues:  envValues,
	}

	return stack, nil
}

func (stack *WasmStack) Run() error {
	store := wasmtime.NewStore(stack.engine)
	linker := wasmtime.NewLinker(store)

	wasiCfg := wasmtime.NewWasiConfig()
	wasiCfg.InheritArgv()
	wasiCfg.InheritStderr()
	wasiCfg.SetEnv(stack.envKeys, stack.envValues)
	wasiCfg.SetStdinFile(stack.stdin)
	wasiCfg.SetStdoutFile(stack.stdout)

	//TODO: do not link against this hard coded value
	// inspect the object and find the right wasi snapshot
	wasiInstance, err := wasmtime.NewWasiInstance(
		store,
		wasiCfg,
		"wasi_snapshot_preview1")
	if err != nil {
		return err
	}

	if err := linker.DefineWasi(wasiInstance); err != nil {
		return err
	}

	instance, err := linker.Instantiate(stack.wasmModule)
	if err != nil {
		return err
	}

	_, err = instance.GetExport("_start").Func().Call()
	return err
}
