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
	"github.com/onflow/cadence/interpreter"
	"github.com/onflow/cadence/sema"

	"github.com/onflow/cadence/bbq/commons"
)

var nativeFunctions []*Global

func NativeFunctions() map[string]*Global {
	funcs := make(map[string]*Global, len(nativeFunctions))
	for _, nativeFunction := range nativeFunctions {

		// Always return a copy.
		// Because the indexes are modified my the imported program.
		funcs[nativeFunction.Name] = &Global{
			Name:     nativeFunction.Name,
			Location: nativeFunction.Location,
			Index:    nativeFunction.Index,
		}
	}
	return funcs
}

var stdlibFunctions = []string{
	commons.LogFunctionName,
	commons.AssertFunctionName,
	commons.PanicFunctionName,
	commons.GetAccountFunctionName,
}

type builtinFunction struct {
	name string
	typ  *sema.FunctionType
}

var commonBuiltinTypeBoundFunctions = []builtinFunction{
	{
		name: sema.GetTypeFunctionName,
		typ:  sema.GetTypeFunctionType,
	},
	{
		name: sema.IsInstanceFunctionName,
		typ:  sema.IsInstanceFunctionType,
	},
}

var valueConstructorFunctions = []builtinFunction{
	{
		name: sema.StringType.Name,
		typ:  sema.StringFunctionType,
	},
}

func init() {
	// Here the order isn't really important.
	// Because the native functions used by a program are also
	// added to the imports section of the compiled program.
	// Then the VM will link the imports (native functions) by the name.
	for _, typ := range commons.BuiltinTypes {
		registerBoundFunctions(typ)
	}

	for _, constructor := range valueConstructorFunctions {
		// Register the constructor. e.g: `String()`
		addNativeFunction(constructor.name)

		// Register the members of the constructor/type. e.g: `String.join()`
		registerBoundFunctions(constructor.typ)
	}

	for _, funcName := range stdlibFunctions {
		addNativeFunction(funcName)
	}

	// Type constructors
	for _, typeConstructor := range sema.RuntimeTypeConstructors {
		addNativeFunction(typeConstructor.Name)
	}

	// Value conversion functions
	for _, declaration := range interpreter.ConverterDeclarations {
		addNativeFunction(declaration.Name)
		declarationVariable := sema.BaseValueActivation.Find(declaration.Name)
		registerBoundFunctions(declarationVariable.Type)
	}
}

func registerBoundFunctions(typ sema.Type) {
	for name := range typ.GetMembers() { //nolint:maprange
		funcName := commons.TypeQualifiedName(typ, name)
		addNativeFunction(funcName)
	}

	compositeType, ok := typ.(*sema.CompositeType)
	if ok && compositeType.NestedTypes != nil {
		compositeType.NestedTypes.Foreach(func(_ string, nestedType sema.Type) {
			registerBoundFunctions(nestedType)
		})
	}
}

func addNativeFunction(name string) {
	global := &Global{
		Name: name,
	}
	nativeFunctions = append(nativeFunctions, global)
}
