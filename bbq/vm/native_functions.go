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
	"github.com/onflow/cadence/bbq/commons"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/errors"
	"github.com/onflow/cadence/interpreter"
	"github.com/onflow/cadence/sema"
	"github.com/onflow/cadence/stdlib"
)

const (
	receiverIndex                   = 0
	typeBoundFunctionArgumentOffset = 1
)

var nativeFunctions = map[string]Value{}

// BuiltInLocation is the location of built-in constructs.
// It's always nil.
var BuiltInLocation common.Location = nil

type BuiltinGlobalsProvider func() map[string]*Variable

func NativeFunctions() map[string]*Variable {
	funcs := make(map[string]*Variable, len(nativeFunctions))
	for name, value := range nativeFunctions { //nolint:maprange
		variable := &interpreter.SimpleVariable{}
		variable.InitializeWithValue(value)
		funcs[name] = variable
	}
	return funcs
}

func RegisterFunction(functionValue *NativeFunctionValue) {
	functionName := functionValue.Name
	registerFunction(functionName, functionValue)
}

func registerFunction(functionName string, functionValue *NativeFunctionValue) {
	_, ok := nativeFunctions[functionName]
	if ok {
		panic(errors.NewUnexpectedError("function already exists: %s", functionName))
	}

	nativeFunctions[functionName] = functionValue
}

func RegisterTypeBoundFunction(typeName string, functionValue *NativeFunctionValue) {
	// Update the name of the function to be type-qualified
	qualifiedName := commons.QualifiedName(typeName, functionValue.Name)
	functionValue.Name = qualifiedName

	RegisterFunction(functionValue)
}

func RegisterTypeBoundCommonFunction(typeName string, functionValue *NativeFunctionValue) {
	// Here the function value is common for many types.
	// Hence, do not update the function name to be type-qualified.
	// Only the key in the map is type-qualified.
	qualifiedName := commons.QualifiedName(typeName, functionValue.Name)
	registerFunction(qualifiedName, functionValue)
}

func init() {
	RegisterFunction(
		NewNativeFunctionValue(
			commons.LogFunctionName,
			stdlib.LogFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {
				value := arguments[0]
				return stdlib.Log(
					context,
					context,
					value,
					EmptyLocationRange,
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			commons.AssertFunctionName,
			stdlib.AssertFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {
				result, ok := arguments[0].(interpreter.BoolValue)
				if !ok {
					panic(errors.NewUnreachableError())
				}

				var message string
				if len(arguments) > 1 {
					messageValue, ok := arguments[1].(*interpreter.StringValue)
					if !ok {
						panic(errors.NewUnreachableError())
					}
					message = messageValue.Str
				}

				return stdlib.Assert(
					result,
					message,
					EmptyLocationRange,
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			commons.PanicFunctionName,
			stdlib.PanicFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {
				message := arguments[0]
				return stdlib.PanicWithError(
					message,
					EmptyLocationRange,
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			commons.GetAccountFunctionName,
			stdlib.GetAccountFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {
				address := arguments[0].(interpreter.AddressValue)
				return NewAccountReferenceValue(
					context,
					context.AccountHandler,
					common.Address(address),
				)
			},
		),
	)

	// Type constructors

	RegisterFunction(
		NewNativeFunctionValue(
			sema.MetaTypeName,
			sema.MetaTypeFunctionType,
			func(context *Context, typeArguments []bbq.StaticType, arguments ...Value) Value {
				return interpreter.NewTypeValue(
					context.MemoryGauge,
					typeArguments[0],
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.OptionalTypeFunctionName,
			sema.OptionalTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				typeValue := arguments[0].(interpreter.TypeValue)

				return interpreter.ConstructOptionalTypeValue(context, typeValue)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.VariableSizedArrayTypeFunctionName,
			sema.VariableSizedArrayTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				typeValue := arguments[0].(interpreter.TypeValue)

				return interpreter.ConstructVariableSizedArrayTypeValue(
					context,
					typeValue,
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.ConstantSizedArrayTypeFunctionName,
			sema.ConstantSizedArrayTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				typeValue := arguments[0].(interpreter.TypeValue)
				sizeValue := arguments[1].(interpreter.IntValue)

				return interpreter.ConstructConstantSizedArrayTypeValue(
					context,
					EmptyLocationRange,
					typeValue,
					sizeValue,
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.DictionaryTypeFunctionName,
			sema.DictionaryTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				keyTypeValue := arguments[0].(interpreter.TypeValue)
				valueTypeValue := arguments[1].(interpreter.TypeValue)

				return interpreter.ConstructDictionaryTypeValue(
					context,
					keyTypeValue,
					valueTypeValue,
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.CompositeTypeFunctionName,
			sema.CompositeTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				typeIDValue := arguments[0].(*interpreter.StringValue)

				return interpreter.ConstructCompositeTypeValue(context, typeIDValue)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.FunctionTypeFunctionName,
			sema.FunctionTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				parameterTypeValues := arguments[0].(*interpreter.ArrayValue)
				returnTypeValue := arguments[1].(interpreter.TypeValue)

				return interpreter.ConstructFunctionTypeValue(
					context,
					EmptyLocationRange,
					parameterTypeValues,
					returnTypeValue,
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.ReferenceTypeFunctionName,
			sema.ReferenceTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				entitlementValues := arguments[0].(*interpreter.ArrayValue)
				typeValue := arguments[1].(interpreter.TypeValue)

				return interpreter.ConstructReferenceTypeValue(
					context,
					EmptyLocationRange,
					entitlementValues,
					typeValue,
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.IntersectionTypeFunctionName,
			sema.IntersectionTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				intersectionIDs := arguments[0].(*interpreter.ArrayValue)

				return interpreter.ConstructIntersectionTypeValue(
					context,
					EmptyLocationRange,
					intersectionIDs,
				)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.CapabilityTypeFunctionName,
			sema.CapabilityTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				typeValue := arguments[0].(interpreter.TypeValue)

				return interpreter.ConstructCapabilityTypeValue(context, typeValue)
			},
		),
	)

	RegisterFunction(
		NewNativeFunctionValue(
			sema.InclusiveRangeTypeFunctionName,
			sema.InclusiveRangeTypeFunctionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

				typeValue := arguments[0].(interpreter.TypeValue)

				return interpreter.ConstructInclusiveRangeTypeValue(context, typeValue)
			},
		),
	)

	// Value conversion functions
	for _, declaration := range interpreter.ConverterDeclarations {
		// NOTE: declare in loop, as captured in closure below
		convert := declaration.Convert

		functionType := sema.BaseValueActivation.Find(declaration.Name).Type.(*sema.FunctionType)

		function := NewNativeFunctionValue(
			declaration.Name,
			functionType,
			func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {
				return convert(
					context.MemoryGauge,
					arguments[0],
					EmptyLocationRange,
				)
			},
		)
		RegisterFunction(function)

		addMember := func(name string, value interpreter.Value) {
			if function.fields == nil {
				function.fields = make(map[string]interpreter.Value)
			}
			if _, exists := function.fields[name]; exists {
				panic(errors.NewUnexpectedError("member already exists: %s", name))
			}
			function.fields[name] = value
		}

		if declaration.Min != nil {
			addMember(sema.NumberTypeMinFieldName, declaration.Min)
		}

		if declaration.Max != nil {
			addMember(sema.NumberTypeMaxFieldName, declaration.Max)
		}

		if stringValueParser, ok := interpreter.StringValueParsers[declaration.Name]; ok {
			RegisterTypeBoundFunction(
				commons.TypeQualifier(stringValueParser.ReceiverType),
				newFromStringFunction(stringValueParser),
			)
		}

		if bigEndianBytesConverter, ok := interpreter.BigEndianBytesConverters[declaration.Name]; ok {
			RegisterTypeBoundFunction(
				commons.TypeQualifier(bigEndianBytesConverter.ReceiverType),
				newFromBigEndianBytesFunction(bigEndianBytesConverter),
			)
		}
	}

	// Value constructors

	RegisterFunction(
		NewNativeFunctionValue(
			sema.StringType.String(),
			sema.StringFunctionType,
			func(_ *Context, _ []bbq.StaticType, _ ...Value) Value {
				return interpreter.EmptyString
			},
		),
	)

	// Register type-bound functions that are common to many types.
	registerCommonBuiltinTypeBoundFunctions()

	registerAllSaturatingArithmeticFunctions()
}

func registerCommonBuiltinTypeBoundFunctions() {
	for _, builtinType := range commons.BuiltinTypes {
		typeQualifier := commons.TypeQualifier(builtinType)
		registerBuiltinTypeBoundFunctions(typeQualifier)
	}

	for _, function := range commonBuiltinTypeBoundFunctions {
		IndexedCommonBuiltinTypeBoundFunctions[function.Name] = function
	}
}

func registerBuiltinTypeBoundFunctions(
	typeQualifier string,
) {
	for _, boundFunction := range commonBuiltinTypeBoundFunctions {
		RegisterTypeBoundCommonFunction(
			typeQualifier,
			boundFunction,
		)
	}
}

// Built-in functions that are common to all the types.
var commonBuiltinTypeBoundFunctions = []*NativeFunctionValue{

	// `isInstance` function
	NewNativeFunctionValue(
		sema.IsInstanceFunctionName,
		sema.IsInstanceFunctionType,
		func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {
			value := arguments[receiverIndex]

			typeValue, ok := arguments[typeBoundFunctionArgumentOffset].(interpreter.TypeValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			return interpreter.IsInstance(context, value, typeValue)
		},
	),

	// `getType` function
	NewNativeFunctionValue(
		sema.GetTypeFunctionName,
		sema.GetTypeFunctionType,
		func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {
			value := arguments[receiverIndex]
			return interpreter.ValueGetType(context, value)
		},
	),

	// TODO: add remaining functions
}

var IndexedCommonBuiltinTypeBoundFunctions = map[string]*NativeFunctionValue{}

func registerAllSaturatingArithmeticFunctions() {
	for _, ty := range common.Concat(
		sema.AllUnsignedFixedPointTypes,
		sema.AllSignedFixedPointTypes,
		sema.AllUnsignedIntegerTypes,
		sema.AllSignedIntegerTypes,
	) {
		registerSaturatingArithmeticFunctions(ty.(sema.SaturatingArithmeticType))
	}
}

func registerSaturatingArithmeticFunctions(t sema.SaturatingArithmeticType) {

	register := func(
		functionName string,
		op func(context *Context, v, other interpreter.NumberValue) interpreter.NumberValue,
	) {
		RegisterTypeBoundFunction(
			commons.TypeQualifier(t),
			NewNativeFunctionValue(
				functionName,
				sema.SaturatingArithmeticTypeFunctionTypes[t],
				func(context *Context, _ []bbq.StaticType, args ...Value) Value {
					v, ok := args[receiverIndex].(interpreter.NumberValue)
					if !ok {
						panic(errors.NewUnreachableError())
					}

					other, ok := args[typeBoundFunctionArgumentOffset].(interpreter.NumberValue)
					if !ok {
						panic(errors.NewUnreachableError())
					}

					return op(context, v, other)
				},
			),
		)
	}

	if t.SupportsSaturatingAdd() {
		register(
			sema.NumericTypeSaturatingAddFunctionName,
			func(context *Context, v, other interpreter.NumberValue) interpreter.NumberValue {
				return v.SaturatingPlus(
					context,
					other,
					EmptyLocationRange,
				)
			},
		)
	}

	if t.SupportsSaturatingSubtract() {
		register(
			sema.NumericTypeSaturatingSubtractFunctionName,
			func(context *Context, v, other interpreter.NumberValue) interpreter.NumberValue {
				return v.SaturatingMinus(
					context,
					other,
					EmptyLocationRange,
				)
			},
		)
	}

	if t.SupportsSaturatingMultiply() {
		register(
			sema.NumericTypeSaturatingMultiplyFunctionName,
			func(context *Context, v, other interpreter.NumberValue) interpreter.NumberValue {
				return v.SaturatingMul(
					context,
					other,
					EmptyLocationRange,
				)
			},
		)
	}

	if t.SupportsSaturatingDivide() {
		register(
			sema.NumericTypeSaturatingDivideFunctionName,
			func(context *Context, v, other interpreter.NumberValue) interpreter.NumberValue {
				return v.SaturatingDiv(
					context,
					other,
					EmptyLocationRange,
				)
			},
		)
	}
}

func newFromStringFunction(typedParser interpreter.TypedStringValueParser) *NativeFunctionValue {
	functionType := sema.FromStringFunctionType(typedParser.ReceiverType)
	parser := typedParser.Parser

	return NewNativeFunctionValue(
		sema.FromStringFunctionName,
		functionType,
		func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {
			argument, ok := arguments[0].(*interpreter.StringValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}
			return parser(context, argument.Str)
		},
	)
}

func newFromBigEndianBytesFunction(typedConverter interpreter.TypedBigEndianBytesConverter) *NativeFunctionValue {
	functionType := sema.FromBigEndianBytesFunctionType(typedConverter.ReceiverType)
	byteLength := typedConverter.ByteLength
	converter := typedConverter.Converter

	return NewNativeFunctionValue(
		sema.FromBigEndianBytesFunctionName,
		functionType,
		func(context *Context, _ []bbq.StaticType, arguments ...Value) Value {

			argument, ok := arguments[0].(*interpreter.ArrayValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			bytes, err := interpreter.ByteArrayValueToByteSlice(context, argument, EmptyLocationRange)
			if err != nil {
				return interpreter.Nil
			}

			// overflow
			if byteLength != 0 && uint(len(bytes)) > byteLength {
				return interpreter.Nil
			}

			return interpreter.NewSomeValueNonCopying(context, converter(context, bytes))
		},
	)
}
