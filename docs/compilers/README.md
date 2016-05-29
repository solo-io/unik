**Compilers** conform to the interface:
```go
type Compiler interface {
	CompileRawImage(params types.CompileImageParams) (*types.RawImage, error)
}
```

Where `CompileImageParams` consists of the following:

```go
type CompileImageParams struct {
	SourcesDir string //path to directory containing application source code
	Args string //arguments to pass to the kernel at runtime
	MntPoints []string //mount points to expect at runtime
	NoCleanup bool //indicates to the compiler not to clean up compilation artifacts after exiting
}
```

The job of a compiler is to compile a directory source files to a raw boot disk image. The behavior of compilers is meant to be independent of providers. Compilers can pass additional information required by providers in the `RawImage` return type, such as what Storage Driver or Network Adapter to use with unikernels created by this compiler. See the [types](../../pkg/types/) package for more about `RawImage`
 
Providers must specify what compilers they are compatible with through their `GetConfig()` method. If you've added a compiler to UniK, you should add the compiler's name to the provider's `GetConfig()` method for each provider your compiler is intended to be used with.

To add compiler support to UniK, you must the compiler to the `_compilers` map in the Unik API Server constructor function `func NewUnikDaemon(config config.DaemonConfig) (*UnikDaemon, error)` in [`daemon.go`](../pkg/daemon/daemon.go)

Your change should look something like this:
```go
func NewUnikDaemon(config config.DaemonConfig) (*UnikDaemon, error) {
	_compilers := make(map[string]compilers.Compiler)
  //...
  //Add your compiler here, like so:

  myCompiler, err := mycompiler.NewCompiler()
  if err != nil {
    //handle err
  }
  _compilers["my_compiler_name"] = myCompiler

  //...
  d := &UnikDaemon{
    server:    lxmartini.QuietMartini(),
    providers: _providers,
    compilers: _compilers,
  }
  return d, nil
}
```
