/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodefeaturerule

import (
	"testing"

	"github.com/stretchr/testify/assert"

	nfdv1alpha1 "sigs.k8s.io/node-feature-discovery/pkg/apis/nfd/v1alpha1"
)

type BoolAssertionFunc func(assert.TestingT, bool, ...interface{}) bool

type ValueAssertionFunc func(assert.TestingT, interface{}, ...interface{}) bool

func TestEvaluateMatchExpression(t *testing.T) {
	type V = nfdv1alpha1.MatchValue
	type TC struct {
		name   string
		op     nfdv1alpha1.MatchOp
		values V
		input  interface{}
		valid  bool
		result BoolAssertionFunc
	}

	tcs := []TC{
		{name: "MatchAny-1", op: nfdv1alpha1.MatchAny, result: assert.True},
		{name: "MatchAny-2", op: nfdv1alpha1.MatchAny, input: "2", valid: false, result: assert.True},

		{name: "MatchIn-1", op: nfdv1alpha1.MatchIn, values: V{"1"}, input: "2", valid: false, result: assert.False},
		{name: "MatchIn-2", op: nfdv1alpha1.MatchIn, values: V{"1"}, input: "2", valid: true, result: assert.False},
		{name: "MatchIn-3", op: nfdv1alpha1.MatchIn, values: V{"1", "2", "3"}, input: "2", valid: false, result: assert.False},
		{name: "MatchIn-4", op: nfdv1alpha1.MatchIn, values: V{"1", "2", "3"}, input: "2", valid: true, result: assert.True},

		{name: "MatchNotIn-1", op: nfdv1alpha1.MatchNotIn, values: V{"2"}, input: 2, valid: false, result: assert.False},
		{name: "MatchNotIn-2", op: nfdv1alpha1.MatchNotIn, values: V{"1"}, input: 2, valid: true, result: assert.True},
		{name: "MatchNotIn-3", op: nfdv1alpha1.MatchNotIn, values: V{"1", "2", "3"}, input: "2", valid: false, result: assert.False},
		{name: "MatchNotIn-4", op: nfdv1alpha1.MatchNotIn, values: V{"1", "2", "3"}, input: "2", valid: true, result: assert.False},

		{name: "MatchInRegexp-1", op: nfdv1alpha1.MatchInRegexp, values: V{"val-[0-9]$"}, input: "val-1", valid: false, result: assert.False},
		{name: "MatchInRegexp-2", op: nfdv1alpha1.MatchInRegexp, values: V{"val-[0-9]$"}, input: "val-1", valid: true, result: assert.True},
		{name: "MatchInRegexp-3", op: nfdv1alpha1.MatchInRegexp, values: V{"val-[0-9]$"}, input: "val-12", valid: true, result: assert.False},
		{name: "MatchInRegexp-4", op: nfdv1alpha1.MatchInRegexp, values: V{"val-[0-9]$", "al-[1-9]"}, input: "val-12", valid: true, result: assert.True},

		{name: "MatchExists-1", op: nfdv1alpha1.MatchExists, input: nil, valid: false, result: assert.False},
		{name: "MatchExists-2", op: nfdv1alpha1.MatchExists, input: nil, valid: true, result: assert.True},

		{name: "MatchDoesNotExist-1", op: nfdv1alpha1.MatchDoesNotExist, input: false, valid: false, result: assert.True},
		{name: "MatchDoesNotExist-2", op: nfdv1alpha1.MatchDoesNotExist, input: false, valid: true, result: assert.False},

		{name: "MatchGt-1", op: nfdv1alpha1.MatchGt, values: V{"2"}, input: 3, valid: false, result: assert.False},
		{name: "MatchGt-2", op: nfdv1alpha1.MatchGt, values: V{"2"}, input: 2, valid: true, result: assert.False},
		{name: "MatchGt-3", op: nfdv1alpha1.MatchGt, values: V{"2"}, input: 3, valid: true, result: assert.True},
		{name: "MatchGt-4", op: nfdv1alpha1.MatchGt, values: V{"-10"}, input: -3, valid: true, result: assert.True},

		{name: "MatchLt-1", op: nfdv1alpha1.MatchLt, values: V{"2"}, input: "1", valid: false, result: assert.False},
		{name: "MatchLt-2", op: nfdv1alpha1.MatchLt, values: V{"2"}, input: "2", valid: true, result: assert.False},
		{name: "MatchLt-3", op: nfdv1alpha1.MatchLt, values: V{"-10"}, input: -3, valid: true, result: assert.False},
		{name: "MatchLt-4", op: nfdv1alpha1.MatchLt, values: V{"2"}, input: "1", valid: true, result: assert.True},

		{name: "MatchGtLt-1", op: nfdv1alpha1.MatchGtLt, values: V{"1", "10"}, input: "1", valid: false, result: assert.False},
		{name: "MatchGtLt-2", op: nfdv1alpha1.MatchGtLt, values: V{"1", "10"}, input: "1", valid: true, result: assert.False},
		{name: "MatchGtLt-3", op: nfdv1alpha1.MatchGtLt, values: V{"1", "10"}, input: "10", valid: true, result: assert.False},
		{name: "MatchGtLt-4", op: nfdv1alpha1.MatchGtLt, values: V{"1", "10"}, input: "2", valid: true, result: assert.True},

		{name: "MatchIsTrue-1", op: nfdv1alpha1.MatchIsTrue, input: true, valid: false, result: assert.False},
		{name: "MatchIsTrue-2", op: nfdv1alpha1.MatchIsTrue, input: true, valid: true, result: assert.True},
		{name: "MatchIsTrue-3", op: nfdv1alpha1.MatchIsTrue, input: false, valid: true, result: assert.False},

		{name: "MatchIsFalse-1", op: nfdv1alpha1.MatchIsFalse, input: "false", valid: false, result: assert.False},
		{name: "MatchIsFalse-2", op: nfdv1alpha1.MatchIsFalse, input: "false", valid: true, result: assert.True},
		{name: "MatchIsFalse-3", op: nfdv1alpha1.MatchIsFalse, input: "true", valid: true, result: assert.False},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			me := &nfdv1alpha1.MatchExpression{Op: tc.op, Value: tc.values}
			res, err := evaluateMatchExpression(me, tc.valid, tc.input)
			tc.result(t, res)
			assert.Nil(t, err)
		})
	}

	// Error cases
	tcs = []TC{
		{name: "MatchAny-err-1", op: nfdv1alpha1.MatchAny, values: V{"1"}, input: "val"},

		{name: "MatchIn-err-1", op: nfdv1alpha1.MatchIn, input: "val"},

		{name: "MatchNotIn-err-1", op: nfdv1alpha1.MatchNotIn, input: "val"},

		{name: "MatchInRegexp-err-1", op: nfdv1alpha1.MatchInRegexp, input: "val"},
		{name: "MatchInRegexp-err-2", op: nfdv1alpha1.MatchInRegexp, values: V{"("}, input: "val"},

		{name: "MatchExists-err-1", op: nfdv1alpha1.MatchExists, values: V{"1"}},

		{name: "MatchDoesNotExist-err-1", op: nfdv1alpha1.MatchDoesNotExist, values: V{"1"}},

		{name: "MatchGt-err-1", op: nfdv1alpha1.MatchGt, input: "1"},
		{name: "MatchGt-err-2", op: nfdv1alpha1.MatchGt, values: V{"1", "2"}, input: "1"},
		{name: "MatchGt-err-3", op: nfdv1alpha1.MatchGt, values: V{""}, input: "1"},
		{name: "MatchGt-err-4", op: nfdv1alpha1.MatchGt, values: V{"2"}, input: "3a"},

		{name: "MatchLt-err-1", op: nfdv1alpha1.MatchLt, input: "1"},
		{name: "MatchLt-err-2", op: nfdv1alpha1.MatchLt, values: V{"1", "2", "3"}, input: "1"},
		{name: "MatchLt-err-3", op: nfdv1alpha1.MatchLt, values: V{"a"}, input: "1"},
		{name: "MatchLt-err-4", op: nfdv1alpha1.MatchLt, values: V{"2"}, input: "1.0"},

		{name: "MatchGtLt-err-1", op: nfdv1alpha1.MatchGtLt, input: "1"},
		{name: "MatchGtLt-err-2", op: nfdv1alpha1.MatchGtLt, values: V{"1"}, input: "1"},
		{name: "MatchGtLt-err-3", op: nfdv1alpha1.MatchGtLt, values: V{"2", "1"}, input: "1"},
		{name: "MatchGtLt-err-4", op: nfdv1alpha1.MatchGtLt, values: V{"1", "2", "3"}, input: "1"},
		{name: "MatchGtLt-err-5", op: nfdv1alpha1.MatchGtLt, values: V{"a", "2"}, input: "1"},
		{name: "MatchGtLt-err-6", op: nfdv1alpha1.MatchGtLt, values: V{"1", "10"}, input: "1.0"},

		{name: "MatchIsTrue-err-1", op: nfdv1alpha1.MatchIsTrue, values: V{"1"}, input: "true"},

		{name: "MatchIsFalse-err-1", op: nfdv1alpha1.MatchIsFalse, values: V{"1", "2"}, input: "false"},

		{name: "invalid-op-err", op: "non-existent-op", values: V{"1"}, input: 1},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			me := &nfdv1alpha1.MatchExpression{Op: tc.op, Value: tc.values}
			res, err := evaluateMatchExpression(me, true, tc.input)
			assert.False(t, res)
			assert.NotNil(t, err)
		})
	}
}

func TestEvaluateMatchExpressionKeys(t *testing.T) {
	type V = nfdv1alpha1.MatchValue
	type I = map[string]nfdv1alpha1.Nil
	type TC struct {
		name   string
		key    string
		op     nfdv1alpha1.MatchOp
		values V
		input  I
		result BoolAssertionFunc
		err    ValueAssertionFunc
	}

	tcs := []TC{
		{name: "1", op: nfdv1alpha1.MatchAny, result: assert.True, err: assert.Nil},

		{name: "2", op: nfdv1alpha1.MatchExists, key: "foo", input: nil, result: assert.False, err: assert.Nil},
		{name: "3", op: nfdv1alpha1.MatchExists, key: "foo", input: I{"bar": {}}, result: assert.False, err: assert.Nil},
		{name: "4", op: nfdv1alpha1.MatchExists, key: "foo", input: I{"bar": {}, "foo": {}}, result: assert.True, err: assert.Nil},

		{name: "5", op: nfdv1alpha1.MatchDoesNotExist, key: "foo", input: nil, result: assert.True, err: assert.Nil},
		{name: "6", op: nfdv1alpha1.MatchDoesNotExist, key: "foo", input: I{}, result: assert.True, err: assert.Nil},
		{name: "7", op: nfdv1alpha1.MatchDoesNotExist, key: "foo", input: I{"bar": {}}, result: assert.True, err: assert.Nil},
		{name: "8", op: nfdv1alpha1.MatchDoesNotExist, key: "foo", input: I{"bar": {}, "foo": {}}, result: assert.False, err: assert.Nil},

		// All other ops should return an error
		{name: "9", op: nfdv1alpha1.MatchIn, values: V{"foo"}, key: "foo", result: assert.False, err: assert.NotNil},
		{name: "10", op: nfdv1alpha1.MatchNotIn, values: V{"foo"}, key: "foo", result: assert.False, err: assert.NotNil},
		{name: "11", op: nfdv1alpha1.MatchInRegexp, values: V{"foo"}, key: "foo", result: assert.False, err: assert.NotNil},
		{name: "12", op: nfdv1alpha1.MatchGt, values: V{"1"}, key: "foo", result: assert.False, err: assert.NotNil},
		{name: "13", op: nfdv1alpha1.MatchLt, values: V{"1"}, key: "foo", result: assert.False, err: assert.NotNil},
		{name: "14", op: nfdv1alpha1.MatchGtLt, values: V{"1", "10"}, key: "foo", result: assert.False, err: assert.NotNil},
		{name: "15", op: nfdv1alpha1.MatchIsTrue, key: "foo", result: assert.False, err: assert.NotNil},
		{name: "16", op: nfdv1alpha1.MatchIsFalse, key: "foo", result: assert.False, err: assert.NotNil},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			me := &nfdv1alpha1.MatchExpression{Op: tc.op, Value: tc.values}
			res, err := evaluateMatchExpressionKeys(me, tc.key, tc.input)
			tc.result(t, res)
			tc.err(t, err)
		})
	}
}

func TestEvaluateMatchExpressionValues(t *testing.T) {
	type V = []string
	type I = map[string]string

	type TC struct {
		name   string
		op     nfdv1alpha1.MatchOp
		values V
		key    string
		input  I
		result BoolAssertionFunc
		err    ValueAssertionFunc
	}

	tcs := []TC{
		{name: "1", op: nfdv1alpha1.MatchAny, result: assert.True, err: assert.Nil},

		{name: "2", op: nfdv1alpha1.MatchIn, values: V{"1", "2"}, key: "foo", input: I{"bar": "2"}, result: assert.False, err: assert.Nil},
		{name: "3", op: nfdv1alpha1.MatchIn, values: V{"1", "2"}, key: "foo", input: I{"foo": "3"}, result: assert.False, err: assert.Nil},
		{name: "4", op: nfdv1alpha1.MatchIn, values: V{"1", "2"}, key: "foo", input: I{"foo": "2"}, result: assert.True, err: assert.Nil},

		{name: "5", op: nfdv1alpha1.MatchNotIn, values: V{"1", "2"}, key: "foo", input: I{"bar": "2"}, result: assert.False, err: assert.Nil},
		{name: "6", op: nfdv1alpha1.MatchNotIn, values: V{"1", "2"}, key: "foo", input: I{"foo": "3"}, result: assert.True, err: assert.Nil},
		{name: "7", op: nfdv1alpha1.MatchNotIn, values: V{"1", "2"}, key: "foo", input: I{"foo": "2"}, result: assert.False, err: assert.Nil},

		{name: "8", op: nfdv1alpha1.MatchInRegexp, values: V{"1", "2"}, key: "foo", input: I{"bar": "2"}, result: assert.False, err: assert.Nil},
		{name: "9", op: nfdv1alpha1.MatchInRegexp, values: V{"1", "[0-8]"}, key: "foo", input: I{"foo": "9"}, result: assert.False, err: assert.Nil},
		{name: "10", op: nfdv1alpha1.MatchInRegexp, values: V{"1", "[0-8]"}, key: "foo", input: I{"foo": "2"}, result: assert.True, err: assert.Nil},

		{name: "11", op: nfdv1alpha1.MatchExists, key: "foo", input: I{"bar": "1"}, result: assert.False, err: assert.Nil},
		{name: "12", op: nfdv1alpha1.MatchExists, key: "foo", input: I{"foo": "1"}, result: assert.True, err: assert.Nil},

		{name: "13", op: nfdv1alpha1.MatchDoesNotExist, key: "foo", input: nil, result: assert.True, err: assert.Nil},
		{name: "14", op: nfdv1alpha1.MatchDoesNotExist, key: "foo", input: I{"foo": "1"}, result: assert.False, err: assert.Nil},

		{name: "15", op: nfdv1alpha1.MatchGt, values: V{"2"}, key: "foo", input: I{"bar": "3"}, result: assert.False, err: assert.Nil},
		{name: "16", op: nfdv1alpha1.MatchGt, values: V{"2"}, key: "foo", input: I{"bar": "3", "foo": "2"}, result: assert.False, err: assert.Nil},
		{name: "17", op: nfdv1alpha1.MatchGt, values: V{"2"}, key: "foo", input: I{"bar": "3", "foo": "3"}, result: assert.True, err: assert.Nil},
		{name: "18", op: nfdv1alpha1.MatchGt, values: V{"2"}, key: "foo", input: I{"bar": "str", "foo": "str"}, result: assert.False, err: assert.NotNil},

		{name: "19", op: nfdv1alpha1.MatchLt, values: V{"2"}, key: "foo", input: I{"bar": "1"}, result: assert.False, err: assert.Nil},
		{name: "20", op: nfdv1alpha1.MatchLt, values: V{"2"}, key: "foo", input: I{"bar": "1", "foo": "2"}, result: assert.False, err: assert.Nil},
		{name: "21", op: nfdv1alpha1.MatchLt, values: V{"2"}, key: "foo", input: I{"bar": "1", "foo": "1"}, result: assert.True, err: assert.Nil},
		{name: "22", op: nfdv1alpha1.MatchLt, values: V{"2"}, key: "foo", input: I{"bar": "str", "foo": "str"}, result: assert.False, err: assert.NotNil},

		{name: "23", op: nfdv1alpha1.MatchGtLt, values: V{"-10", "10"}, key: "foo", input: I{"bar": "1"}, result: assert.False, err: assert.Nil},
		{name: "24", op: nfdv1alpha1.MatchGtLt, values: V{"-10", "10"}, key: "foo", input: I{"bar": "1", "foo": "11"}, result: assert.False, err: assert.Nil},
		{name: "25", op: nfdv1alpha1.MatchGtLt, values: V{"-10", "10"}, key: "foo", input: I{"bar": "1", "foo": "-11"}, result: assert.False, err: assert.Nil},
		{name: "26", op: nfdv1alpha1.MatchGtLt, values: V{"-10", "10"}, key: "foo", input: I{"bar": "1", "foo": "1"}, result: assert.True, err: assert.Nil},
		{name: "27", op: nfdv1alpha1.MatchGtLt, values: V{"-10", "10"}, key: "foo", input: I{"bar": "str", "foo": "str"}, result: assert.False, err: assert.NotNil},

		{name: "28", op: nfdv1alpha1.MatchIsTrue, key: "foo", result: assert.False, err: assert.Nil},
		{name: "29", op: nfdv1alpha1.MatchIsTrue, key: "foo", input: I{"foo": "1"}, result: assert.False, err: assert.Nil},
		{name: "30", op: nfdv1alpha1.MatchIsTrue, key: "foo", input: I{"foo": "true"}, result: assert.True, err: assert.Nil},

		{name: "31", op: nfdv1alpha1.MatchIsFalse, key: "foo", input: I{"foo": "true"}, result: assert.False, err: assert.Nil},
		{name: "32", op: nfdv1alpha1.MatchIsFalse, key: "foo", input: I{"foo": "false"}, result: assert.True, err: assert.Nil},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			me := &nfdv1alpha1.MatchExpression{Op: tc.op, Value: tc.values}
			res, err := evaluateMatchExpressionValues(me, tc.key, tc.input)
			tc.result(t, res)
			tc.err(t, err)
		})
	}
}
