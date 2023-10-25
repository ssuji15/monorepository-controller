/*
Copyright 2023 VMware Inc.

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

package controller

import (
	"context"
	"fmt"
	apiv1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	apiv1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/garethjevans/monorepository-controller/api/v1alpha1"
	"github.com/garethjevans/monorepository-controller/internal/util"
	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	sourcev1alpha1 "github.com/vmware-tanzu/tanzu-source-controller/apis/source/v1alpha1"
	"golang.org/x/mod/sumdb/dirhash"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

//+kubebuilder:rbac:groups=source.garethjevans.org,resources=monorepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=source.garethjevans.org,resources=monorepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=source.garethjevans.org,resources=monorepositories/finalizers,verbs=update
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=helmrepositories;gitrepositories;ocirepositories,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;create;update

func NewMonoRepositoryReconciler(c reconcilers.Config) *reconcilers.ResourceReconciler[*v1alpha1.MonoRepository] {
	return &reconcilers.ResourceReconciler[*v1alpha1.MonoRepository]{
		Name: "MonoRepository",
		Reconciler: reconcilers.Sequence[*v1alpha1.MonoRepository]{
			NewResourceValidator(c),
			NewChecksumCalculator(c),
		},
		Config: c,
	}
}

func NewResourceValidator(c reconcilers.Config) reconcilers.SubReconciler[*v1alpha1.MonoRepository] {
	return &reconcilers.SyncReconciler[*v1alpha1.MonoRepository]{
		Name: "ResourceValidator",
		Sync: func(ctx context.Context, resource *v1alpha1.MonoRepository) error {
			log := util.L(ctx)
			// resolve the input
			key := resource.Spec.SourceRef.Key(resource.ObjectMeta.Namespace)

			component := GetKind(resource.Spec.SourceRef.ApiVersion, resource.Spec.SourceRef.Kind)

			err := c.Client.Get(ctx, key, component)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Info("unable to resolve", "name", key.Name, "namespace", key.Namespace)
					resource.Status.MarkResourceMissing(key.Name, key.Name, key.Namespace)
				} else {
					log.Error(err, "error resolving", "name", key.Name, "namespace", key.Namespace)
					resource.Status.MarkFailed(err)
				}
				return nil
			}

			// parse the status
			artifact, err := GetArtifact(component)
			if err != nil {
				log.Error(err, "error finding artifact", "name", key.Name, "namespace", key.Namespace)
				resource.Status.MarkFailed(err)
			}

			log.Info("got", "artifact", artifact)
			resource.Status.MarkArtifactResolved(artifact.URL)

			stashArtifact(ctx, artifact)

			return nil
		},
	}
}

const artifactKey reconcilers.StashKey = "artifact"

func stashArtifact(ctx context.Context, artifact v1alpha1.Artifact) {
	reconcilers.StashValue(ctx, artifactKey, artifact)
}

func retrieveArtifact(ctx context.Context) v1alpha1.Artifact {
	if components, ok := reconcilers.RetrieveValue(ctx, artifactKey).(v1alpha1.Artifact); ok {
		return components
	}

	return v1alpha1.Artifact{}
}

func NewChecksumCalculator(c reconcilers.Config) reconcilers.SubReconciler[*v1alpha1.MonoRepository] {
	return &reconcilers.SyncReconciler[*v1alpha1.MonoRepository]{
		Name: "ChecksumCalculator",
		Sync: func(ctx context.Context, resource *v1alpha1.MonoRepository) error {
			log := util.L(ctx)

			artifact := retrieveArtifact(ctx)
			if artifact.URL != "" {
				// create a temporary directory
				tempDir, err := os.MkdirTemp("", "tmp")
				if err != nil {
					return err
				}

				// cleanup on exit
				defer os.RemoveAll(tempDir)

				log.Info("created temp dir", "dir", tempDir)

				// download the filter and copy from/to path
				tarGzLocation := filepath.Join(tempDir, fmt.Sprintf("%s.tar.gz", resource.Spec.SourceRef.Name))
				err = util.DownloadFile(tarGzLocation, artifact.URL)
				if err != nil {
					return err
				}

				// extract tar.gz to temp location
				tarGzExtractedLocation := filepath.Join(tempDir, fmt.Sprintf("%s-extracted", resource.Spec.SourceRef.Name))
				err = util.ExtractTarGz(tarGzLocation, tarGzExtractedLocation)
				if err != nil {
					return err
				}

				files, err := util.ListFiles(tarGzExtractedLocation)
				if err != nil {
					return err
				}

				log.Info("Full file list", "files", files)

				filteredFiles := util.FilterFileList(files, resource.Spec.Include)
				log.Info("Using files for checksum calculation", "files", filteredFiles)
				resource.Status.ObservedFileList = strings.Join(filteredFiles, "\n")

				hash, err := dirhash.Hash1(filteredFiles, func(name string) (io.ReadCloser, error) {
					return os.Open(filepath.Join(tarGzExtractedLocation, name))
				})
				if err != nil {
					return err
				}

				log.Info("Calculated checksum", "checksum", hash)

				if resource.Status.Artifact != nil && resource.Status.Artifact.Checksum == hash {
					// nothing has changed, do nothing
					log.Info("Source hasn't changed, there is nothing to update",
						"name", resource.Spec.SourceRef.Name,
						"kind", resource.Spec.SourceRef.Kind,
						"apiVersion", resource.Spec.SourceRef.ApiVersion)
				} else {
					old := "<NA>"
					if resource.Status.Artifact != nil {
						old = resource.Status.Artifact.Checksum
					}

					log.Info("Source has changed! updating status with new checksum",
						"checksum", hash,
						"old", old)
					resource.Status.Artifact = &v1alpha1.Artifact{
						Path:           artifact.Path,
						URL:            artifact.URL,
						Revision:       artifact.Revision,
						Checksum:       hash,
						Digest:         artifact.Digest,
						LastUpdateTime: artifact.LastUpdateTime,
						Size:           artifact.Size,
						Metadata:       artifact.Metadata,
					}
					resource.Status.URL = artifact.URL
				}

				resource.Status.ObservedInclude = resource.Spec.Include
				resource.Status.MarkReady()
			}

			return nil
		},
	}
}

func GetKind(apiVersion string, kind string) client.Object {
	type match struct {
		kind       string
		apiVersion string
	}

	in := match{apiVersion: apiVersion, kind: kind}

	switch in {
	case match{apiVersion: "source.toolkit.fluxcd.io/v1beta2", kind: "OCIRepository"}:
		return &apiv1beta2.OCIRepository{}
	case match{apiVersion: "source.toolkit.fluxcd.io/v1beta2", kind: "HelmRepository"}:
		return &apiv1beta2.HelmRepository{}
	case match{apiVersion: "source.toolkit.fluxcd.io/v1beta2", kind: "GitRepository"}:
		return &apiv1beta2.GitRepository{}
	case match{apiVersion: "source.toolkit.fluxcd.io/v1beta1", kind: "HelmRepository"}:
		return &apiv1beta1.HelmRepository{}
	case match{apiVersion: "source.toolkit.fluxcd.io/v1beta1", kind: "GitRepository"}:
		return &apiv1beta1.GitRepository{}
	case match{apiVersion: "source.apps.tanzu.vmware.com/v1alpha1", kind: "ImageRepository"}:
		return &sourcev1alpha1.ImageRepository{}
	}

	return &apiv1beta2.GitRepository{}
}

func GetArtifact(in interface{}) (v1alpha1.Artifact, error) {
	switch v := in.(type) {
	case *apiv1beta2.OCIRepository:
		return v1alpha1.Artifact{
			URL:            v.GetArtifact().URL,
			Path:           v.GetArtifact().Path,
			Revision:       v.GetArtifact().Revision,
			Size:           v.GetArtifact().Size,
			Checksum:       v.GetArtifact().Checksum,
			Digest:         v.GetArtifact().Digest,
			LastUpdateTime: v.GetArtifact().LastUpdateTime,
		}, nil
	case *apiv1beta2.GitRepository:
		return v1alpha1.Artifact{
			URL:            v.GetArtifact().URL,
			Path:           v.GetArtifact().Path,
			Revision:       v.GetArtifact().Revision,
			Size:           v.GetArtifact().Size,
			Checksum:       v.GetArtifact().Checksum,
			Digest:         v.GetArtifact().Digest,
			LastUpdateTime: v.GetArtifact().LastUpdateTime,
		}, nil
	case *apiv1beta2.HelmRepository:
		return v1alpha1.Artifact{
			URL:            v.GetArtifact().URL,
			Path:           v.GetArtifact().Path,
			Revision:       v.GetArtifact().Revision,
			Size:           v.GetArtifact().Size,
			Checksum:       v.GetArtifact().Checksum,
			Digest:         v.GetArtifact().Digest,
			LastUpdateTime: v.GetArtifact().LastUpdateTime,
		}, nil
	case *apiv1beta1.GitRepository:
		return v1alpha1.Artifact{
			URL:            v.GetArtifact().URL,
			Path:           v.GetArtifact().Path,
			Revision:       v.GetArtifact().Revision,
			Checksum:       v.GetArtifact().Checksum,
			LastUpdateTime: v.GetArtifact().LastUpdateTime,
		}, nil
	case *apiv1beta1.HelmRepository:
		return v1alpha1.Artifact{
			URL:            v.GetArtifact().URL,
			Path:           v.GetArtifact().Path,
			Revision:       v.GetArtifact().Revision,
			Checksum:       v.GetArtifact().Checksum,
			LastUpdateTime: v.GetArtifact().LastUpdateTime,
		}, nil
	case *sourcev1alpha1.ImageRepository:
		return v1alpha1.Artifact{
			URL:            v.Status.Artifact.URL,
			Path:           v.Status.Artifact.Path,
			Revision:       v.Status.Artifact.Revision,
			Checksum:       v.Status.Artifact.Checksum,
			LastUpdateTime: v.Status.Artifact.LastUpdateTime,
		}, nil
	default:
		return v1alpha1.Artifact{}, fmt.Errorf("unknown type %s", v)
	}
}
