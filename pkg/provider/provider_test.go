// Copyright (C) 2020-2022 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package provider

import (
	"testing"

	"errors"

	corev1 "k8s.io/api/core/v1"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// All this catalogSources and installPlans are used by more than one unit test, so make sure
	// you fully understand them before changing these values.
	// They define runtime objects for 2 CSVs "op1.v1.0.1" and "op2.v2.0.2" that are installed in
	// namespaces ns1 (op1) and ns2 (op1 + op2). So there's one catalogSource + installPlan for
	// each installation. Subscriptions and CSVs are only needed by TestCreateOperators, so they're
	// defined there only.
	catalogSource1 = olmv1Alpha.CatalogSource{
		TypeMeta:   metav1.TypeMeta{Kind: "CatalogSource"},
		ObjectMeta: metav1.ObjectMeta{Name: "catalogSource1", Namespace: "ns1"},
		Spec:       olmv1Alpha.CatalogSourceSpec{Image: "catalogSource1Image"},
		Status:     olmv1Alpha.CatalogSourceStatus{},
	}

	catalogSource2 = olmv1Alpha.CatalogSource{
		TypeMeta:   metav1.TypeMeta{Kind: "CatalogSource"},
		ObjectMeta: metav1.ObjectMeta{Name: "catalogSource2", Namespace: "ns2"},
		Spec:       olmv1Alpha.CatalogSourceSpec{Image: "catalogSource2Image"},
		Status:     olmv1Alpha.CatalogSourceStatus{},
	}

	catalogSource3 = olmv1Alpha.CatalogSource{
		TypeMeta:   metav1.TypeMeta{Kind: "CatalogSource"},
		ObjectMeta: metav1.ObjectMeta{Name: "catalogSource3", Namespace: "ns2"},
		Spec:       olmv1Alpha.CatalogSourceSpec{Image: "catalogSource3Image"},
		Status:     olmv1Alpha.CatalogSourceStatus{},
	}

	ns1InstallPlan1 = olmv1Alpha.InstallPlan{
		TypeMeta: metav1.TypeMeta{Kind: "InstallPlan"}, ObjectMeta: metav1.ObjectMeta{Name: "ns1Plan1", Namespace: "ns1"},
		Spec: olmv1Alpha.InstallPlanSpec{CatalogSource: "catalogSource1", CatalogSourceNamespace: "ns1",
			ClusterServiceVersionNames: []string{"op1.v1.0.1"}, Approval: olmv1Alpha.ApprovalAutomatic, Approved: true},
		Status: olmv1Alpha.InstallPlanStatus{BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath1",
			CatalogSourceRef: &corev1.ObjectReference{Name: "catalogSource1", Namespace: "ns1"}}}},
	}

	ns2InstallPlan1 = olmv1Alpha.InstallPlan{
		TypeMeta: metav1.TypeMeta{Kind: "InstallPlan"}, ObjectMeta: metav1.ObjectMeta{Name: "ns2Plan1", Namespace: "ns2"},
		Spec: olmv1Alpha.InstallPlanSpec{CatalogSource: "catalogSource2", CatalogSourceNamespace: "ns2",
			ClusterServiceVersionNames: []string{"op1.v1.0.1"}, Approval: olmv1Alpha.ApprovalAutomatic, Approved: true},
		Status: olmv1Alpha.InstallPlanStatus{BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath2",
			CatalogSourceRef: &corev1.ObjectReference{Name: "catalogSource2", Namespace: "ns2"}}}},
	}

	ns2InstallPlan2 = olmv1Alpha.InstallPlan{
		TypeMeta: metav1.TypeMeta{Kind: "InstallPlan"}, ObjectMeta: metav1.ObjectMeta{Name: "ns2Plan2", Namespace: "ns2"},
		Spec: olmv1Alpha.InstallPlanSpec{CatalogSource: "catalogSource3", CatalogSourceNamespace: "ns2",
			ClusterServiceVersionNames: []string{"op2.v2.0.2"}, Approval: olmv1Alpha.ApprovalAutomatic, Approved: true},
		Status: olmv1Alpha.InstallPlanStatus{BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath3",
			CatalogSourceRef: &corev1.ObjectReference{Name: "catalogSource3", Namespace: "ns2"}}}},
	}
)

func TestGetUID(t *testing.T) {
	testCases := []struct {
		testCID     string
		expectedErr error
		expectedUID string
	}{
		{
			testCID:     "cid://testing",
			expectedErr: nil,
			expectedUID: "testing",
		},
		{
			testCID:     "cid://",
			expectedErr: errors.New("cannot determine container UID"),
			expectedUID: "",
		},
	}

	for _, tc := range testCases {
		c := NewContainer()
		c.Status.ContainerID = tc.testCID
		uid, err := c.GetUID()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedUID, uid)
	}
}

//nolint:funlen
func TestConvertArrayPods(t *testing.T) {
	testCases := []struct {
		testPods     []*corev1.Pod
		expectedPods []*Pod
	}{
		{ // Test Case 1 - No containers
			testPods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testpod1",
						Namespace: "testnamespace1",
					},
				},
			},
			expectedPods: []*Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testpod1",
							Namespace: "testnamespace1",
						},
					},
				},
			},
		},
		{ // Test Case 2 - Containers
			testPods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testpod1",
						Namespace: "testnamespace1",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "testcontainer1",
							},
						},
					},
				},
			},
			expectedPods: []*Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testpod1",
							Namespace: "testnamespace1",
						},
					},
					Containers: []*Container{
						{
							Container: &corev1.Container{
								Name: "testcontainer1",
							},
							Namespace: "testnamespace1",
							Podname:   "testpod1",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		convertedArray := ConvertArrayPods(tc.testPods)
		assert.Equal(t, tc.expectedPods[0].Containers, convertedArray[0].Containers)
		assert.Equal(t, tc.expectedPods[0].Name, convertedArray[0].Name)
		assert.Equal(t, tc.expectedPods[0].Namespace, convertedArray[0].Namespace)
	}
}

func TestIsSkipHelmChart(t *testing.T) {
	testCases := []struct {
		testHelmName   string
		testList       []configuration.SkipHelmChartList
		expectedOutput bool
	}{
		{ // Test Case #1 - Helm Chart names match, skipping
			testHelmName: "test1",
			testList: []configuration.SkipHelmChartList{
				{
					Name: "test1",
				},
			},
			expectedOutput: true,
		},
		{ // Test Case #2 - Helm Chart names mismatch, not skipping
			testHelmName: "test2",
			testList: []configuration.SkipHelmChartList{
				{
					Name: "test1",
				},
			},
			expectedOutput: false,
		},
		{ // Test Case #3 - Empty list
			testHelmName:   "test3",
			testList:       []configuration.SkipHelmChartList{},
			expectedOutput: false,
		},
		{ // Test Case #4 - Empty list, helm name empty
			testHelmName:   "",
			testList:       []configuration.SkipHelmChartList{},
			expectedOutput: false,
		},
		{ // Test Case #5 - Helm Chart name missing
			testHelmName: "",
			testList: []configuration.SkipHelmChartList{
				{
					Name: "test1",
				},
			},
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, isSkipHelmChart(tc.testHelmName, tc.testList))
	}
}

func TestContainerStringFuncs(t *testing.T) {
	testCases := []struct {
		nodename    string
		namespace   string
		podname     string
		name        string
		containerID string
		runtime     string

		expectedStringOutput     string
		expectedStringLongOutput string
	}{
		{
			nodename:                 "testnode",
			namespace:                "testnamespace",
			podname:                  "testpod",
			name:                     "name1",
			containerID:              "cID1",
			runtime:                  "runtime1",
			expectedStringLongOutput: "node: testnode ns: testnamespace podName: testpod containerName: name1 containerUID: cID1 containerRuntime: runtime1",
			expectedStringOutput:     "container: name1 pod: testpod ns: testnamespace",
		},
		{
			nodename:                 "testnode",
			namespace:                "testnamespace",
			podname:                  "testpod",
			name:                     "name2",
			containerID:              "cID2",
			runtime:                  "runtime2",
			expectedStringLongOutput: "node: testnode ns: testnamespace podName: testpod containerName: name2 containerUID: cID2 containerRuntime: runtime2",
			expectedStringOutput:     "container: name2 pod: testpod ns: testnamespace",
		},
	}

	for _, tc := range testCases {
		c := &Container{
			NodeName:  tc.nodename,
			Namespace: tc.namespace,
			Podname:   tc.podname,
			Container: &corev1.Container{
				Name: tc.name,
			},
			Status: corev1.ContainerStatus{
				ContainerID: tc.containerID,
			},
			Runtime: tc.runtime,
		}
		assert.Equal(t, tc.expectedStringLongOutput, c.StringLong())
		assert.Equal(t, tc.expectedStringOutput, c.String())
	}
}

//nolint:funlen
func TestIsWorkerNode(t *testing.T) {
	testCases := []struct {
		node           *corev1.Node
		expectedResult bool
	}{
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1"}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/master": ""}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/worker": ""}},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/worker": "blahblah"}},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/worker": ""}},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/worker": ""}},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		node := Node{Data: tc.node}
		assert.Equal(t, tc.expectedResult, node.IsWorkerNode())
	}
}

//nolint:funlen
func TestIsMasterNode(t *testing.T) {
	testCases := []struct {
		node           *corev1.Node
		expectedResult bool
	}{
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1"}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/worker": ""}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/master": ""}},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/master": "blahblah"}},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/control-plane": "blablah"}},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/master": ""}},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/control-plane": ""}},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/master": "", "node-role.kubernetes.io/control-plane": ""}},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		node := Node{Data: tc.node}
		assert.Equal(t, tc.expectedResult, node.IsMasterNode())
	}
}

func TestGetNodeCount(t *testing.T) {
	generateEnv := func(isMaster bool) *TestEnvironment {
		key := "node-role.kubernetes.io/worker"
		if isMaster {
			key = "node-role.kubernetes.io/master"
		}

		return &TestEnvironment{
			Nodes: map[string]Node{
				"node1": {
					Data: &corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node1",
							Labels: map[string]string{key: ""},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testIsMaster bool
	}{
		{
			testIsMaster: true,
		},
		{
			testIsMaster: false,
		},
	}

	for _, tc := range testCases {
		tEnv := generateEnv(tc.testIsMaster)

		if tc.testIsMaster {
			assert.Equal(t, 1, tEnv.GetMasterCount())
			assert.Equal(t, 0, tEnv.GetWorkerCount())
		} else {
			assert.Equal(t, 1, tEnv.GetWorkerCount())
			assert.Equal(t, 0, tEnv.GetMasterCount())
		}
	}
}

func TestIsRTKernel(t *testing.T) {
	generateNode := func(kernel string) *Node {
		return &Node{
			Data: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						KernelVersion: kernel,
					},
				},
			},
		}
	}

	testCases := []struct {
		testKernel     string
		expectedOutput bool
	}{
		{ // Test Case #1 - Kernel is RT
			testKernel:     "3.10.0-1127.10.1.rt56.1106.el7",
			expectedOutput: true,
		},
		{ // Test Case #2 - Kernel is standard
			testKernel:     "3.10.0-1127.10.1.1106.el7",
			expectedOutput: false,
		},
		{ // Test Case #3 - Kernel string empty
			testKernel:     "",
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		n := generateNode(tc.testKernel)
		assert.Equal(t, n.IsRTKernel(), tc.expectedOutput)
	}
}

func TestIsRHCOS(t *testing.T) {
	testCases := []struct {
		testImageName  string
		expectedOutput bool
	}{
		{
			testImageName:  "Red Hat Enterprise Linux CoreOS 410.84.202205031645-0 (Ootpa)",
			expectedOutput: true,
		},
		{
			testImageName:  "Ubuntu 20.04",
			expectedOutput: false,
		},
		{
			testImageName:  "Ubuntu 21.10",
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		node := Node{
			Data: &corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage: tc.testImageName,
					},
				},
			},
		}
		assert.Equal(t, tc.expectedOutput, node.IsRHCOS())
	}
}

func TestIsRHEL(t *testing.T) {
	testCases := []struct {
		testImageName  string
		expectedOutput bool
	}{
		{
			testImageName:  "Red Hat Enterprise Linux 8.5 (Ootpa)",
			expectedOutput: true,
		},
		{
			testImageName:  "Ubuntu 20.04",
			expectedOutput: false,
		},
		{
			testImageName:  "Ubuntu 21.10",
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		node := Node{
			Data: &corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage: tc.testImageName,
					},
				},
			},
		}
		assert.Equal(t, tc.expectedOutput, node.IsRHEL())
	}
}

func TestGetRHCOSVersion(t *testing.T) {
	testCases := []struct {
		testImageName  string
		expectedOutput string
		expectedErr    error
	}{
		{
			testImageName:  "Red Hat Enterprise Linux CoreOS 410.84.202205031645-0 (Ootpa)",
			expectedOutput: "4.10.14",
			expectedErr:    nil,
		},
		{
			testImageName:  "Ubuntu 20.04",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 20.04"),
		},
		{
			testImageName:  "Ubuntu 21.10",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 21.10"),
		},
	}

	for _, tc := range testCases {
		node := Node{
			Data: &corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage: tc.testImageName,
					},
				},
			},
		}

		origValue := rhcosRelativePath
		rhcosRelativePath = "%s/../../cnf-certification-test/platform/operatingsystem/files/rhcos_version_map" // for testing only
		result, err := node.GetRHCOSVersion()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedOutput, result)
		rhcosRelativePath = origValue
	}
}

func TestGetRHELVersion(t *testing.T) {
	testCases := []struct {
		testImageName  string
		expectedOutput string
		expectedErr    error
	}{
		{
			testImageName:  "Red Hat Enterprise Linux 8.5 (Ootpa)",
			expectedOutput: "8.5",
			expectedErr:    nil,
		},
		{
			testImageName:  "Ubuntu 20.04",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 20.04"),
		},
		{
			testImageName:  "Ubuntu 21.10",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 21.10"),
		},
	}

	for _, tc := range testCases {
		node := Node{
			Data: &corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage: tc.testImageName,
					},
				},
			},
		}
		result, err := node.GetRHELVersion()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedOutput, result)
	}
}