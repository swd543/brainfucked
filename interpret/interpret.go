package interpret

import (
	"fmt"
	"io"
)

/*
StackImpl is a generic array based stack implementation
*/
type StackImpl[T any] struct {
	data []T
	iter int
}

/*
NewStackImpl is the constructor for StackImpl
*/
func NewStackImpl[T any]() *StackImpl[T] {
	return &StackImpl[T]{data: make([]T, 0)}
}

/*
Stack is the interface for a generic Stack
*/
type Stack[T any] interface {
	Push(data T)
	Pop() T
	Peek() T
	IsEmpty() bool
}

/*
Push Data into the stack
*/
func (s *StackImpl[T]) Push(data T) {
	s.data = append(s.data, data)
	s.iter++
}

/*
Pop Data from the top of the stack, removing the element
*/
func (s *StackImpl[T]) Pop() T {
	s.iter--
	r := s.data[s.iter]
	s.data = s.data[:s.iter]
	return r
}

/*
Peek the top element of the stack without removing
*/
func (s *StackImpl[T]) Peek() T {
	r := s.data[s.iter-1]
	return r
}

/*
IsEmpty returns if the Stack is empty
*/
func (s *StackImpl[T]) IsEmpty() bool {
	return len(s.data) <= 0
}

/*
CellType hosts the types of Cells allowed
*/
type CellType interface {
	~int | ~uint | ~byte
}

/*
State is a generic struct that contains the entire state of the interpreter.
Define custom commands in the mapping variable.
*/
type State[C CellType] struct {
	// reader is the interface for the Program stream
	reader io.Reader
	// writer is the interface for the Program output
	writer io.Writer
	// inputReader is the interface for reading the inputs to the Program
	inputReader io.Reader
	// Program is a holder for the lazy-loaded Program
	Program []byte
	// Data is an array of cells of type C
	Data []C
	// Pc is Program counter, or the pointer to the instruction
	Pc uint
	// Dp is Data pointer, or the pointer to the Data cell
	Dp uint
	// mapping is the mapping between characters to functions,
	// which mutate the State
	mapping map[byte]func(state *State[C])
	// loopStack is the stack for keeping track of loops
	loopStack Stack[uint]
	// jumpMap is an optimization over the streamed Program, providing
	// fast lookup for jump instructions
	jumpMap map[uint]uint
}

/*
NewState is the constructor for State
*/
func NewState[C CellType](reader io.Reader, writer io.Writer, inputReader io.Reader) *State[C] {
	return &State[C]{
		reader:      reader,
		writer:      writer,
		inputReader: inputReader,
		Program:     make([]byte, 0, 30000),
		Data:        make([]C, 300, 30000),
		Dp:          0,
		Pc:          0,
		loopStack:   NewStackImpl[uint](),
		jumpMap:     map[uint]uint{},
		mapping: map[byte]func(state *State[C]){
			'>': func(state *State[C]) {
				state.Dp++
				state.Pc++
			},
			'<': func(state *State[C]) {
				state.Dp--
				state.Pc++
			},
			'+': func(state *State[C]) {
				state.Data[state.Dp]++
				state.Pc++
			},
			'-': func(state *State[C]) {
				state.Data[state.Dp]--
				state.Pc++
			},
			'.': func(state *State[C]) {
				_, err := writer.Write([]byte(string(rune(state.Data[state.Dp]))))
				if err != nil {
					fmt.Errorf("%e while writing output", err)
				}
				//fmt.Print(string(rune(state.Data[state.Dp])))
				state.Pc++
			},
			',': func(state *State[C]) {
				read := make([]byte, 1)
				_, err := inputReader.Read(read)
				if err != nil {
					fmt.Errorf("%e while reading input", err)
				}
				state.Data[state.Dp] = C(read[0])
			},
			'[': func(state *State[C]) {
				// push current location in stack
				state.loopStack.Push(state.Pc)
				// jump forward
				if state.Data[state.Dp] == 0 {
					// if exists in jumpMap, jump there
					if v, ok := state.jumpMap[state.loopStack.Peek()]; ok {
						state.Pc = v
					} else {
						// find the next closing brace and jump to it
						state.findNextClosingBrace()
						state.Pc = state.jumpMap[state.loopStack.Peek()]
					}
				} else {
					// continue into loop
					state.Pc++
				}
			},
			']': func(state *State[C]) {
				v := state.loopStack.Pop()
				state.jumpMap[v] = state.Pc
				// continue inside scope
				if state.Data[state.Dp] == 0 {
					state.Pc++
				} else {
					// jump to corresponding open brace
					state.Pc = v
				}
			},
		}}
}

/*
AddOrReplaceCommand enables us to add custom commands to the interpreter
*/
func (self *State[C]) AddOrReplaceCommand(symbol byte, function func(state *State[C])) {
	self.mapping[symbol] = function
}

/*
DeleteCommand deletes a command in the function mapping
*/
func (self *State[C]) DeleteCommand(symbol byte) {
	delete(self.mapping, symbol)
}

/*
GetCommand allows us to get the function for a command
*/
func (self *State[C]) GetCommand(symbol byte) func(state *State[C]) {
	return self.mapping[symbol]
}

/*
findNextClosingBrace streams the provided reader, reading (and parsing
subloops) up to the next closing brace. Invoked when the corresponding
close brace has not been encountered yet, and a jump forward is requested.
*/
func (self *State[T]) findNextClosingBrace() {
	buff := make([]byte, 1)
	var tempStack Stack[uint] = NewStackImpl[uint]()
	tempStack.Push(self.Pc)
	// loop until requested matching closing brace found
	for i := uint(len(self.Program)); ; {
		_, err := self.reader.Read(buff)
		if err != nil {
			// if error encountered, panic
			panic(err)
		}
		op := buff[0]
		if _, ok := self.mapping[op]; ok {
			// if valid command, append to Program
			self.Program = append(self.Program, op)
			if op == '[' {
				tempStack.Push(i)
			} else if op == ']' {
				v := tempStack.Pop()
				self.jumpMap[v] = i
				if tempStack.IsEmpty() {
					break
				}
			}
			i++
		}
	}
}

/*
GetNextSymbol streams the provided reader interface, either returning the next
instruction or back to the loop start if the Program requests so.
*/
func (self *State[T]) GetNextSymbol() (byte, error) {
	var err error
	if self.Pc >= uint(len(self.Program)) {
		buff := make([]byte, 1)
		for i := uint(len(self.Program)); i <= self.Pc; {
			_, err = self.reader.Read(buff)
			if err == nil {
			} else {
				return 0, err
			}
			instruction := buff[0]
			if self.mapping[instruction] != nil {
				self.Program = append(self.Program, instruction)
				i++
			}
		}
	}
	return self.Program[self.Pc], err
}
