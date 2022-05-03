package interpret

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewStackImpl(t *testing.T) {
	tests := []struct {
		name   string
		want   any
		output any
	}{
		// TODO: Add test cases.
		{
			name:   "testNewStackImplInt",
			want:   &StackImpl[int]{data: make([]int, 0)},
			output: NewStackImpl[int](),
		},
		{
			name:   "testNewStackImplUintPtr",
			want:   &StackImpl[uintptr]{data: make([]uintptr, 0)},
			output: NewStackImpl[uintptr](),
		},
		{
			name:   "testNewStackImplString",
			want:   &StackImpl[string]{data: make([]string, 0)},
			output: NewStackImpl[string](),
		},
		{
			name:   "testNewStackImplPtr",
			want:   &StackImpl[*string]{data: make([]*string, 0)},
			output: NewStackImpl[*string](),
		},
		{
			name:   "testNewStackImplRune",
			want:   &StackImpl[rune]{data: make([]rune, 0)},
			output: NewStackImpl[rune](),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.output, tt.want) {
				t.Errorf("NewStackImpl() = %v, want %v", tt.output, tt.want)
			}
		})
	}
}

func TestStackImpl_Peek(t *testing.T) {
	t.Run("testPeekNoElementsThrowsException", func(t *testing.T) {
		var stack Stack[int] = NewStackImpl[int]()
		defer func() {
			err := recover()
			if err == nil {
				t.Errorf("Peek() => %v expected runtime exception, got", err)
			}
		}()
		stack.Peek()
	})
	t.Run("testPeekGivesCorrectElements", func(t *testing.T) {
		var stack Stack[int] = NewStackImpl[int]()
		for i := 0; i < 100000; i++ {
			stack.Push(i)
			output := stack.Peek()
			if output != i {
				t.Errorf("Peek() = %v, want %v", output, i)
			}
		}
	})
}

func TestStackImpl_Pop(t *testing.T) {
	t.Run("testPopNoElementsThrowsException", func(t *testing.T) {
		var stack Stack[int] = NewStackImpl[int]()
		defer func() {
			err := recover()
			if err == nil {
				t.Errorf("Pop() => %v expected runtime exception, got", err)
			}
		}()
		stack.Pop()
	})
	t.Run("testPopGivesCorrectElements", func(t *testing.T) {
		var stack Stack[int] = NewStackImpl[int]()
		for i := 0; i < 100000; i++ {
			stack.Push(i)
			output := stack.Pop()
			if output != i {
				t.Errorf("Peek() = %v, want %v", output, i)
			}
		}
	})
}

func TestBrainfuck(t *testing.T) {
	programs := []string{"bockbeer.bf", "helloworld.bf", "squares.bf", "triangle.bf", "yapi.bf"}
	for _, p := range programs {
		t.Run("testProgram_"+p, func(t *testing.T) {
			filename := filepath.Join("..", "programs", p)
			reader, _ := os.Open(filename)
			var b bytes.Buffer
			state := NewState[int](reader, &b, nil)
			for {
				if symbol, err := state.GetNextSymbol(); err == nil {
					state.GetCommand(symbol)(state)
				} else {
					if err != io.EOF {
						t.Errorf("Unexpected state for %s, %e", p, err)
					}
					break
				}
			}
			t.Log(b.String())
		})
	}
}
