// Copyright (c) 2014 Aleksa Sarai (Cyphar)

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// * The above copyright notice and this permission notice shall be included in all
//   copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package brainfuck

import (
	"fmt"
	"io"
)

const StackSize int = 1 << 16

// Stores the jump address (offset in the code) of the last while section.
type jumpHeap []int

// Pushes a jump position to the heap.
func (j *jumpHeap) Push(v int) {
	*j = append(*j, v)
}

// Pops the last jump position.
func (j *jumpHeap) Pop() (int, error) {
	l := len(*j) - 1

	if l < 0 {
		return 0, fmt.Errorf("brainfuck eval: cannot pop from empty heap")
	}

	i := (*j)[l]
	*j = (*j)[:l]

	return i, nil
}

// Evaluate brainfuck code.
func Eval(rawCode interface{}, rw io.ReadWriter) error {
	var (
		tape   [StackSize]byte
		cursor int

		jumpTable = new(jumpHeap)

		code         = rawCode.(string)
		end      int = len(code)
		position int
	)

	for position = 0; position < end; position++ {
		switch code[position] {
		case '>':
			cursor += 1

			// Wrap the cursor around.
			if cursor < 0 {
				cursor = 0
			}
		case '<':
			cursor -= 1

			// Wrap the cursor around.
			if cursor < 0 {
				cursor = len(tape)
			}
		case '+':
			tape[cursor] += 1
		case '-':
			tape[cursor] -= 1
		case ',':
			// Make a one-byte buffer for the input.
			char := make([]byte, 1)

			// Read.
			if _, err := rw.Read(char); err != nil {
				return fmt.Errorf("brainfuck eval: cannot read byte from io")
			}

			// Copy from the buffer to the tape.
			copy(tape[cursor:cursor+1], char)
		case '.':
			// Make and copy to a one-byte buffer for the output.
			char := make([]byte, 1)
			copy(char, tape[cursor:cursor+1])

			// Write.
			if _, err := rw.Write(char); err != nil {
				return fmt.Errorf("brainfuck eval: cannot write byte to io")
			}
		case '[':
			// Save current position as jump table.
			jumpTable.Push(position)
		case ']':
			// Get last jump position.
			last, err := jumpTable.Pop()
			if err != nil {
				return err
			}

			// Condition is based on current tape value.
			cond := tape[cursor]

			// If the condition is "false" then don't push the jump table back.
			// Just let the position increment in the next iter.
			if cond == 0 {
				break
			}

			// Move position and push back.
			jumpTable.Push(last)
			position = last
		default:
			// Ignore invalid chars.
		}
	}

	return nil
}
