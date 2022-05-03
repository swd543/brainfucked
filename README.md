# brainfucked 

![Build](https://github.com/swd543/brainfucked/actions/workflows/go-coverage.yml/badge.svg)
![Coverage](https://img.shields.io/badge/Coverage-87.7%25-brightgreen)


A stack based, streaming brainfuck interpreter written in golang ;)

# Running
`cd brainfuck && go run . ../programs/helloworld.bf`

# Testing
`cd interpret && go test -v -cover`

# Adding as a package
Add the interpret package => `go get github.com/swd543/brainfucked/interpret`

## Initialize the interpreter state
```go
state := interpret.NewState[int](programReader, programOutputWriter, inputReader)
```
## Adding custom commands (for squaring)
```go
state.AddOrReplaceCommand('*', func(state *interpret.State[int]) {
  state.Data[state.Dp] *= state.Data[state.Dp]
  state.Pc++
})
```
