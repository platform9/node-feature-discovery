/*
Copyright 2023 The Kubernetes Authors.

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

package e2e

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	"golang.org/x/exp/maps"

	corev1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	e2elog "k8s.io/kubernetes/test/e2e/framework"
)

type k8sLabels map[string]string

// eventuallyNonControlPlaneNodes is a helper for asserting node properties
func eventuallyNonControlPlaneNodes(ctx context.Context, cli clientset.Interface) AsyncAssertion {
	return Eventually(func(g Gomega, ctx context.Context) ([]corev1.Node, error) {
		return getNonControlPlaneNodes(ctx, cli)
	}).WithPolling(1 * time.Second).WithTimeout(10 * time.Second).WithContext(ctx)
}

// MatchLabels returns a specialized Gomega matcher for checking if a list of
// nodes are labeled as expected.
func MatchLabels(expectedNew map[string]k8sLabels, oldNodes []corev1.Node, ignoreUnexpected bool) gomegatypes.GomegaMatcher {
	matcher := &nodeIterablePropertyMatcher[k8sLabels]{
		propertyName:     "labels",
		ignoreUnexpected: ignoreUnexpected,
		matchFunc: func(newNode, oldNode corev1.Node, expected k8sLabels) ([]string, []string, []string) {
			expectedAll := maps.Clone(oldNode.Labels)
			maps.Copy(expectedAll, expected)
			return matchMap(newNode.Labels, expectedAll)
		},
	}

	return &nodeListPropertyMatcher[k8sLabels]{
		expected: expectedNew,
		oldNodes: oldNodes,
		matcher:  matcher,
	}
}

// nodeListPropertyMatcher is a generic Gomega matcher for asserting one property a group of nodes.
type nodeListPropertyMatcher[T any] struct {
	expected map[string]T
	oldNodes []corev1.Node

	matcher nodePropertyMatcher[T]
}

// nodePropertyMatcher is a generic helper type for matching one node.
type nodePropertyMatcher[T any] interface {
	match(newNode, oldNode corev1.Node, expected T) bool
	message() string
	negatedMessage() string
}

// Match method of the GomegaMatcher interface.
func (m *nodeListPropertyMatcher[T]) Match(actual interface{}) (bool, error) {
	nodes, ok := actual.([]corev1.Node)
	if !ok {
		return false, fmt.Errorf("expected []corev1.Node, got: %T", actual)
	}

	for _, node := range nodes {
		expected, ok := m.expected[node.Name]
		if !ok {
			if defaultExpected, ok := m.expected["*"]; ok {
				expected = defaultExpected
			} else {
				e2elog.Logf("Skipping node %q as no expected was specified", node.Name)
				continue
			}
		}

		oldNode := getNode(m.oldNodes, node.Name)
		if matched := m.matcher.match(node, oldNode, expected); !matched {
			return false, nil
		}
	}
	return true, nil
}

// FailureMessage method of the GomegaMatcher interface.
func (m *nodeListPropertyMatcher[T]) FailureMessage(actual interface{}) string {
	return m.matcher.message()
}

// NegatedFailureMessage method of the GomegaMatcher interface.
func (m *nodeListPropertyMatcher[T]) NegatedFailureMessage(actual interface{}) string {
	return m.matcher.negatedMessage()
}

// nodeIterablePropertyMatcher is a nodePropertyMatcher for matching iterable
// elements such as maps or lists.
type nodeIterablePropertyMatcher[T any] struct {
	propertyName     string
	ignoreUnexpected bool
	matchFunc        func(newNode, oldNode corev1.Node, expected T) ([]string, []string, []string)

	// TODO remove nolint when golangci-lint is able to cope with generics
	node         *corev1.Node //nolint:unused
	missing      []string     //nolint:unused
	invalidValue []string     //nolint:unused
	unexpected   []string     //nolint:unused

}

// TODO remove nolint when golangci-lint is able to cope with generics
//
//nolint:unused
func (m *nodeIterablePropertyMatcher[T]) match(newNode, oldNode corev1.Node, expected T) bool {
	m.node = &newNode
	m.missing, m.invalidValue, m.unexpected = m.matchFunc(newNode, oldNode, expected)

	if m.ignoreUnexpected {
		m.unexpected = nil
	}
	return len(m.missing) == 0 && len(m.invalidValue) == 0 && len(m.unexpected) == 0
}

// TODO remove nolint when golangci-lint is able to cope with generics
//
//nolint:unused
func (m *nodeIterablePropertyMatcher[T]) message() string {
	msg := fmt.Sprintf("Node %q %s did not match:", m.node.Name, m.propertyName)
	if len(m.missing) > 0 {
		msg += fmt.Sprintf("\n  missing:\n    %s", strings.Join(m.missing, "\n    "))
	}
	if len(m.invalidValue) > 0 {
		msg += fmt.Sprintf("\n  invalid value:\n    %s", strings.Join(m.invalidValue, "\n    "))
	}
	if len(m.unexpected) > 0 {
		msg += fmt.Sprintf("\n  unexpected:\n    %s", strings.Join(m.unexpected, "\n    "))
	}
	return msg
}

// TODO remove nolint when golangci-lint is able to cope with generics
//
//nolint:unused
func (m *nodeIterablePropertyMatcher[T]) negatedMessage() string {
	return fmt.Sprintf("Node %q matched unexpectedly", m.node.Name)
}

// matchMap is a helper for matching map types
func matchMap[M ~map[K]V, K comparable, V comparable](actual, expected M) (missing, invalid, unexpected []string) {
	for k, ve := range expected {
		va, ok := actual[k]
		if !ok {
			missing = append(missing, fmt.Sprintf("%v", k))
		} else if va != ve {
			invalid = append(invalid, fmt.Sprintf("%v=%v, expected value %v", k, va, ve))
		}
	}
	for k, v := range actual {
		if _, ok := expected[k]; !ok {
			unexpected = append(unexpected, fmt.Sprintf("%v=%v", k, v))
		}
	}
	return missing, invalid, unexpected
}
