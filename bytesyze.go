package bytesyze

import (
	"errors"
	"fmt"
	"io"
)

type byteSyze struct {
	DR byte // data register
	AR byte // address register
	IR byte // instruction register
	SR byte // switch register

	Memory [256]byte

	Stdio io.ReadWriter
}

func (bs *byteSyze) Next() {
	switch bs.Memory[bs.IR] {
	case '<': // load
		bs.DR = bs.Memory[bs.AR]
	case '>': // store
		bs.Memory[bs.AR] = bs.DR
	case '*': // point
		bs.DR, bs.AR = bs.AR, bs.DR
	case '!': // jump
		bs.AR, bs.IR = bs.IR, bs.AR
	case '\\': // switch
		bs.DR, bs.SR = bs.SR, bs.DR
	case '+': // add
		bs.DR = bs.Memory[bs.AR] + bs.DR
	case '-': // subtract
		bs.DR = bs.Memory[bs.AR] - bs.DR
	case '(': // input
		b := make([]byte, 1)
		if n, err := bs.Stdio.Read(b); err != nil || n != 1 {
			bs.DR = 0
		} else {
			bs.DR = b[0]
		}
	case ')': // output
		bs.Stdio.Write([]byte{bs.DR})
	case '?': // conditional
		if bs.DR == 0 {
			bs.IR += 1
		}
	default:
		// intentional NOP
	}

	bs.IR += 1 // progress forward
}

func New(memory [256]byte, stdio io.ReadWriter) *byteSyze {
	return &byteSyze{
		DR: 0,
		AR: 0,
		IR: 0,
		SR: 0,

		Memory: memory,

		Stdio: stdio,
	}
}

// Eval runs a Byte Syze program. It takes a [256]byte array, which is the
// program source code, and returns any errors - check rw's writes for stdout.
func Eval(code interface{}, rw io.ReadWriter) error {
	memory, ok := code.([256]byte)
	if !ok {
		return errors.New("can't cast code input to [256]byte")
	}
	bs := New(memory, rw)
	for bs.IR < 255 {
		bs.Next()
	}
	return nil
}
