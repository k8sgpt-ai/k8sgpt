/*
Copyright 2026 The K8sGPT Authors.
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

package keda

import (
	"context"
	"errors"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

type mockHelmClient struct {
	helmclient.Client
	addOrUpdateChartRepoFunc  func(entry repo.Entry) error
	installOrUpgradeChartFunc func(ctx context.Context, spec *helmclient.ChartSpec, opts *helmclient.GenericHelmOptions) (*release.Release, error)
	listDeployedReleasesFunc  func() ([]*release.Release, error)
	uninstallReleaseFunc      func(spec *helmclient.ChartSpec) error
}

func (m *mockHelmClient) AddOrUpdateChartRepo(entry repo.Entry) error {
	if m.addOrUpdateChartRepoFunc != nil {
		return m.addOrUpdateChartRepoFunc(entry)
	}
	return nil
}

func (m *mockHelmClient) InstallOrUpgradeChart(ctx context.Context, spec *helmclient.ChartSpec, opts *helmclient.GenericHelmOptions) (*release.Release, error) {
	if m.installOrUpgradeChartFunc != nil {
		return m.installOrUpgradeChartFunc(ctx, spec, opts)
	}
	return nil, nil
}

func (m *mockHelmClient) ListDeployedReleases() ([]*release.Release, error) {
	if m.listDeployedReleasesFunc != nil {
		return m.listDeployedReleasesFunc()
	}
	return nil, nil
}

func (m *mockHelmClient) UninstallRelease(spec *helmclient.ChartSpec) error {
	if m.uninstallReleaseFunc != nil {
		return m.uninstallReleaseFunc(spec)
	}
	return nil
}

func TestNewKeda(t *testing.T) {
	k := NewKeda()
	require.NotNil(t, k)
}

func TestKeda_Deploy_ChartRepoErrorReturnsErrorWithoutPanic(t *testing.T) {
	mockErr := errors.New("network unreachable: failed to fetch chart repo")
	mock := &mockHelmClient{
		addOrUpdateChartRepoFunc: func(entry repo.Entry) error {
			assert.Equal(t, RepoShortName, entry.Name)
			assert.Equal(t, Repo, entry.URL)
			return mockErr
		},
	}

	k := (&Keda{}).WithHelmClient(mock)
	err := k.Deploy("keda-test")
	require.Error(t, err)
	assert.ErrorIs(t, err, mockErr)
}

func TestKeda_Deploy_InstallChartErrorReturnsError(t *testing.T) {
	mockErr := errors.New("failed to install helm chart")
	mock := &mockHelmClient{
		addOrUpdateChartRepoFunc: func(entry repo.Entry) error {
			return nil
		},
		installOrUpgradeChartFunc: func(ctx context.Context, spec *helmclient.ChartSpec, opts *helmclient.GenericHelmOptions) (*release.Release, error) {
			assert.Equal(t, "keda-test", spec.Namespace)
			assert.Equal(t, ReleaseName, spec.ReleaseName)
			return nil, mockErr
		},
	}

	k := (&Keda{}).WithHelmClient(mock)
	err := k.Deploy("keda-test")
	require.Error(t, err)
	assert.ErrorIs(t, err, mockErr)
}

func TestKeda_Deploy_Success(t *testing.T) {
	chartRepoCalled := false
	installCalled := false

	mock := &mockHelmClient{
		addOrUpdateChartRepoFunc: func(entry repo.Entry) error {
			chartRepoCalled = true
			return nil
		},
		installOrUpgradeChartFunc: func(ctx context.Context, spec *helmclient.ChartSpec, opts *helmclient.GenericHelmOptions) (*release.Release, error) {
			installCalled = true
			return &release.Release{Name: ReleaseName, Namespace: spec.Namespace}, nil
		},
	}

	k := (&Keda{}).WithHelmClient(mock)
	err := k.Deploy("keda-ns")
	require.NoError(t, err)
	assert.True(t, chartRepoCalled)
	assert.True(t, installCalled)
}

func TestKeda_GetNamespace_Success(t *testing.T) {
	mock := &mockHelmClient{
		listDeployedReleasesFunc: func() ([]*release.Release, error) {
			return []*release.Release{
				{Name: "other-release", Namespace: "default"},
				{Name: ReleaseName, Namespace: "custom-keda-ns"},
			}, nil
		},
	}

	k := (&Keda{}).WithHelmClient(mock)
	ns, err := k.GetNamespace()
	require.NoError(t, err)
	assert.Equal(t, "custom-keda-ns", ns)
}

func TestKeda_GetNamespace_NotFound(t *testing.T) {
	mock := &mockHelmClient{
		listDeployedReleasesFunc: func() ([]*release.Release, error) {
			return []*release.Release{
				{Name: "other-release", Namespace: "default"},
			}, nil
		},
	}

	k := (&Keda{}).WithHelmClient(mock)
	ns, err := k.GetNamespace()
	require.Error(t, err)
	assert.Empty(t, ns)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestKeda_GetNamespace_NilHelm(t *testing.T) {
	k := &Keda{}
	ns, err := k.GetNamespace()
	require.Error(t, err)
	assert.Empty(t, ns)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestKeda_GetAnalyzerName(t *testing.T) {
	k := &Keda{}
	names := k.GetAnalyzerName()
	assert.Equal(t, []string{"ScaledObject"}, names)
}

func TestKeda_OwnsAnalyzer(t *testing.T) {
	k := &Keda{}
	assert.True(t, k.OwnsAnalyzer("ScaledObject"))
	assert.False(t, k.OwnsAnalyzer("Pod"))
}

func TestKeda_AddAnalyzer(t *testing.T) {
	k := &Keda{}
	mergedMap := make(map[string]common.IAnalyzer)
	k.AddAnalyzer(&mergedMap)
	assert.Contains(t, mergedMap, "ScaledObject")
}

func TestKeda_InitHelmClient_AlreadySet(t *testing.T) {
	mock := &mockHelmClient{}
	k := (&Keda{}).WithHelmClient(mock)
	err := k.initHelmClient()
	require.NoError(t, err)
	assert.Same(t, mock, k.helm)
}

func TestKeda_GetNamespace_ListReleasesError(t *testing.T) {
	mockErr := errors.New("list releases error")
	mock := &mockHelmClient{
		listDeployedReleasesFunc: func() ([]*release.Release, error) {
			return nil, mockErr
		},
	}
	k := (&Keda{}).WithHelmClient(mock)
	_, err := k.GetNamespace()
	require.Error(t, err)
	assert.ErrorIs(t, err, mockErr)
}

