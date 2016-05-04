**Compilers** conform to the interface:
```
type Compiler interface {
	CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (*types.RawImage, error)
}
```
The job of a compiler is to compile a compressed archive of source files to a raw boot disk image. The behavior of compilers is meant to be independent of providers; a provider should specify which raw disk images it can run in its `GetConfig()` method.


To add compiler support to UniK, you must add the compiler name to compatible providers' `GetConfig()` method, and add the compiler to the `_compilers` map in the Unik API Server constructor function `func NewUnikDaemon(config config.DaemonConfig) (*UnikDaemon, error)` in [`daemon.go`](../pkg/daemon/daemon.go)

Your change should look something like this:
```
func NewUnikDaemon(config config.DaemonConfig) (*UnikDaemon, error) {
	_compilers := make(map[string]compilers.Compiler)
  //...
  //Add your compiler here, like so:
  /*
  myCompiler, err := mycompiler.NewCompiler()
  if err != nil {
    //handle err
  }
  _compilers["my_compiler_name"] = myCompiler
  */
  //...
  d := &UnikDaemon{
    server:    lxmartini.QuietMartini(),
    providers: _providers,
    compilers: _compilers,
  }
  return d, nil
}
```
