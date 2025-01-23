/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence/bbq"
	"github.com/onflow/cadence/bbq/constantkind"
	"github.com/onflow/cadence/bbq/opcode"
	. "github.com/onflow/cadence/test_utils/sema_utils"
)

func TestCompileRecursionFib(t *testing.T) {

	t.Parallel()

	checker, err := ParseAndCheck(t, `
      fun fib(_ n: Int): Int {
          if n < 2 {
             return n
          }
          return fib(n - 1) + fib(n - 2)
      }
  `)
	require.NoError(t, err)

	compiler := NewBytecodeCompiler(checker)
	program := compiler.Compile()

	require.Len(t, program.Functions, 1)
	require.Equal(t,
		[]byte{
			// if n < 2
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.GetConstant), 0, 0,
			byte(opcode.IntLess),
			byte(opcode.JumpIfFalse), 0, 14,
			// then return n
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.ReturnValue),
			// fib(n - 1)
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.GetConstant), 0, 1,
			byte(opcode.IntSubtract),
			byte(opcode.Transfer), 0, 0,
			byte(opcode.GetGlobal), 0, 0,
			byte(opcode.Invoke), 0, 0,
			// fib(n - 2)
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.GetConstant), 0, 0,
			byte(opcode.IntSubtract),
			byte(opcode.Transfer), 0, 0,
			byte(opcode.GetGlobal), 0, 0,
			byte(opcode.Invoke), 0, 0,
			// return sum
			byte(opcode.IntAdd),
			byte(opcode.ReturnValue),
		},
		compiler.ExportFunctions()[0].Code,
	)

	require.Equal(t,
		[]*bbq.Constant{
			{
				Data: []byte{0x2},
				Kind: constantkind.Int,
			},
			{
				Data: []byte{0x1},
				Kind: constantkind.Int,
			},
		},
		program.Constants,
	)
}

func TestCompileImperativeFib(t *testing.T) {

	t.Parallel()

	checker, err := ParseAndCheck(t, `
      fun fib(_ n: Int): Int {
          var fib1 = 1
          var fib2 = 1
          var fibonacci = fib1
          var i = 2
          while i < n {
              fibonacci = fib1 + fib2
              fib1 = fib2
              fib2 = fibonacci
              i = i + 1
          }
          return fibonacci
      }
  `)
	require.NoError(t, err)

	compiler := NewBytecodeCompiler(checker)
	program := compiler.Compile()

	require.Len(t, program.Functions, 1)
	require.Equal(t,
		[]byte{
			// var fib1 = 1
			byte(opcode.GetConstant), 0, 0,
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 1,
			// var fib2 = 1
			byte(opcode.GetConstant), 0, 0,
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 2,
			// var fibonacci = fib1
			byte(opcode.GetLocal), 0, 1,
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 3,
			// var i = 2
			byte(opcode.GetConstant), 0, 1,
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 4,
			// while i < n
			byte(opcode.GetLocal), 0, 4,
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.IntLess),
			byte(opcode.JumpIfFalse), 0, 93,
			// fibonacci = fib1 + fib2
			byte(opcode.GetLocal), 0, 1,
			byte(opcode.GetLocal), 0, 2,
			byte(opcode.IntAdd),
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 3,
			// fib1 = fib2
			byte(opcode.GetLocal), 0, 2,
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 1,
			// fib2 = fibonacci
			byte(opcode.GetLocal), 0, 3,
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 2,
			// i = i + 1
			byte(opcode.GetLocal), 0, 4,
			byte(opcode.GetConstant), 0, 0,
			byte(opcode.IntAdd),
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 4,
			// continue loop
			byte(opcode.Jump), 0, 36,
			// return fibonacci
			byte(opcode.GetLocal), 0, 3,
			byte(opcode.ReturnValue),
		},
		compiler.ExportFunctions()[0].Code,
	)

	require.Equal(t,
		[]*bbq.Constant{
			{
				Data: []byte{0x1},
				Kind: constantkind.Int,
			},
			{
				Data: []byte{0x2},
				Kind: constantkind.Int,
			},
		},
		program.Constants,
	)
}

func TestCompileBreak(t *testing.T) {

	t.Parallel()

	checker, err := ParseAndCheck(t, `
      fun test(): Int {
          var i = 0
          while true {
              if i > 3 {
                 break
              }
              i = i + 1
          }
          return i
      }
  `)
	require.NoError(t, err)

	compiler := NewBytecodeCompiler(checker)
	program := compiler.Compile()

	require.Len(t, program.Functions, 1)
	require.Equal(t,
		[]byte{
			// var i = 0
			byte(opcode.GetConstant), 0, 0,
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 0,
			// while true
			byte(opcode.True),
			byte(opcode.JumpIfFalse), 0, 42,
			// if i > 3
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.GetConstant), 0, 1,
			byte(opcode.IntGreater),
			byte(opcode.JumpIfFalse), 0, 26,
			// break
			byte(opcode.Jump), 0, 42,
			// i = i + 1
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.GetConstant), 0, 2,
			byte(opcode.IntAdd),
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 0,
			// repeat
			byte(opcode.Jump), 0, 9,
			// return i
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.ReturnValue),
		},
		compiler.ExportFunctions()[0].Code,
	)

	require.Equal(t,
		[]*bbq.Constant{
			{
				Data: []byte{0x0},
				Kind: constantkind.Int,
			},
			{
				Data: []byte{0x3},
				Kind: constantkind.Int,
			},
			{
				Data: []byte{0x1},
				Kind: constantkind.Int,
			},
		},
		program.Constants,
	)
}

func TestCompileContinue(t *testing.T) {

	t.Parallel()

	checker, err := ParseAndCheck(t, `
      fun test(): Int {
          var i = 0
          while true {
              i = i + 1
              if i < 3 {
                 continue
              }
              break
          }
          return i
      }
  `)
	require.NoError(t, err)

	compiler := NewBytecodeCompiler(checker)
	program := compiler.Compile()

	require.Len(t, program.Functions, 1)
	require.Equal(t,
		[]byte{
			// var i = 0
			byte(opcode.GetConstant), 0, 0,
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 0,
			// while true
			byte(opcode.True),
			byte(opcode.JumpIfFalse), 0, 45,
			// i = i + 1
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.GetConstant), 0, 1,
			byte(opcode.IntAdd),
			byte(opcode.Transfer), 0, 0,
			byte(opcode.SetLocal), 0, 0,
			// if i < 3
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.GetConstant), 0, 2,
			byte(opcode.IntLess),
			byte(opcode.JumpIfFalse), 0, 39,
			// continue
			byte(opcode.Jump), 0, 9,
			// break
			byte(opcode.Jump), 0, 45,
			// repeat
			byte(opcode.Jump), 0, 9,
			// return i
			byte(opcode.GetLocal), 0, 0,
			byte(opcode.ReturnValue),
		},
		compiler.ExportFunctions()[0].Code,
	)

	require.Equal(t,
		[]*bbq.Constant{
			{
				Data: []byte{0x0},
				Kind: constantkind.Int,
			},
			{
				Data: []byte{0x1},
				Kind: constantkind.Int,
			},
			{
				Data: []byte{0x3},
				Kind: constantkind.Int,
			},
		},
		program.Constants,
	)
}
