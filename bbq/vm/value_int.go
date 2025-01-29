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
	"strconv"

	"github.com/onflow/atree"

	"github.com/onflow/cadence/interpreter"
)

type IntValue struct {
	SmallInt int64
}

func NewIntValue(smallInt int64) IntValue {
	return IntValue{
		SmallInt: smallInt,
	}
}

func (v IntValue) String() string {
	return strconv.FormatInt(v.SmallInt, 10)
}

var _ Value = IntValue{}

func (IntValue) isValue() {}

func (IntValue) StaticType(*Config) StaticType {
	return interpreter.PrimitiveStaticTypeInt
}

func (v IntValue) Transfer(*Config, atree.Address, bool, atree.Storable) Value {
	return v
}

func (v IntValue) Add(other IntValue) Value {
	sum := safeAdd(int(v.SmallInt), int(other.SmallInt))
	return NewIntValue(int64(sum))
}

func (v IntValue) Subtract(other IntValue) Value {
	return NewIntValue(v.SmallInt - other.SmallInt)
}

func (v IntValue) Less(other IntValue) Value {
	if v.SmallInt < other.SmallInt {
		return TrueValue
	}
	return FalseValue
}

func (v IntValue) LessOrEqual(other IntValue) Value {
	if v.SmallInt <= other.SmallInt {
		return TrueValue
	}
	return FalseValue
}

func (v IntValue) Greater(other IntValue) Value {
	if v.SmallInt > other.SmallInt {
		return TrueValue
	}
	return FalseValue
}

func (v IntValue) GreaterOrEqual(other IntValue) Value {
	if v.SmallInt >= other.SmallInt {
		return TrueValue
	}
	return FalseValue
}
