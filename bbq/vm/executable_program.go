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

package vm

import (
	"github.com/onflow/cadence/bbq"
	"github.com/onflow/cadence/bbq/opcode"
	"github.com/onflow/cadence/common"
)

// ExecutableProgram is the 'executable' version of a `bbq.Program`.
// It holds information that are accessible to a given program,
// such as constants, static-types, and global variables.
// These info are accessed by the opcodes of the program.
// i.e: indexes used in opcodes refer to the indexes of its ExecutableProgram.
type ExecutableProgram struct {
	Location    common.Location
	Program     *bbq.Program[opcode.Instruction, bbq.StaticType]
	Globals     []Value
	Constants   []Value
	StaticTypes []bbq.StaticType
}

func NewExecutableProgram(
	location common.Location,
	program *bbq.Program[opcode.Instruction, bbq.StaticType],
	globals []Value,
) *ExecutableProgram {
	return &ExecutableProgram{
		Location:    location,
		Program:     program,
		Globals:     globals,
		Constants:   make([]Value, len(program.Constants)),
		StaticTypes: program.Types,
	}
}
