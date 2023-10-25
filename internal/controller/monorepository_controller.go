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
	apiv1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/garethjevans/monorepository-controller/api/v1alpha1"
	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:rbac:groups=source.garethjevans.org,resources=monorepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=source.garethjevans.org,resources=monorepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=source.garethjevans.org,resources=monorepositories/finalizers,verbs=update
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=gitrepositories,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;create;update

func NewMonoRepositoryReconciler(c reconcilers.Config) *reconcilers.ResourceReconciler[*v1alpha1.MonoRepository] {
	return &reconcilers.ResourceReconciler[*v1alpha1.MonoRepository]{
		Name: "MonoRepository",
		Reconciler: reconcilers.Sequence[*v1alpha1.MonoRepository]{
			NewResourceValidator(c),
			//NewChecksumCalculator(c),
		},
		Config: c,
	}
}

func NewResourceValidator(c reconcilers.Config) reconcilers.SubReconciler[*v1alpha1.MonoRepository] {
	return &reconcilers.ChildReconciler[*v1alpha1.MonoRepository, *apiv1beta2.GitRepository, *apiv1beta2.GitRepositoryList]{
		Name: "GitRepository",
		DesiredChild: func(ctx context.Context, parent *v1alpha1.MonoRepository) (*apiv1beta2.GitRepository, error) {
			//log := util.L(ctx)

			child := &apiv1beta2.GitRepository{
				ObjectMeta: v1.ObjectMeta{
					Labels:      FilterLabelsOrAnnotations(reconcilers.MergeMaps(parent.Labels)),
					Annotations: FilterLabelsOrAnnotations(reconcilers.MergeMaps(parent.Annotations)),

					GenerateName: fmt.Sprintf("%s-mr-", parent.Name),
					Namespace:    parent.Namespace,
				},
				Spec: parent.Spec.GitRepository,
			}

			return child, nil
		},
		MergeBeforeUpdate: func(actual, desired *apiv1beta2.GitRepository) {
			actual.Labels = desired.Labels
			actual.Spec = desired.Spec
		},
		ReflectChildStatusOnParent: func(ctx context.Context, parent *v1alpha1.MonoRepository, child *apiv1beta2.GitRepository, err error) {
			//log := util.L(ctx)

			if child == nil {
				// parent.Status.MarkCustomRunFailed("Failed", "Failed to resolve")
			} else {
				//for _, rm := range child.Spec.ResultMappings {
				//	if !HasResultWithName(parent.Status.Results, rm.Name) {
				//		if hasResultWithName(child.Status.Results, rm.Name) {
				//			value := findResultWithName(child.Status.Results, rm.Name).Value
				//			log.Info("Adding result", "name", rm.Name, "value", value)
				//
				//			parent.Status.Results = append(parent.Status.Results, tektonv1beta1.CustomRunResult{
				//				Name:  rm.Name,
				//				Value: value,
				//			})
				//		}
				//	}
				//}
				//
				//healthy := child.Status.GetCondition("Healthy")
				//if healthy == nil || healthy.Status == "Unknown" {
				//	parent.Status.MarkCustomRunRunning("Running", "")
				//} else if healthy.Status == "True" {
				//	parent.Status.MarkCustomRunSucceeded("Succeeded", healthy.Message)
				//} else {
				//	parent.Status.MarkCustomRunFailed("Failed", healthy.Message)
				//}
			}
		},
		Sanitize: func(child *apiv1beta2.GitRepository) interface{} {
			return child.Spec
		},
	}

	//return &reconcilers.SyncReconciler[*v1alpha1.MonoRepository]{
	//	Name: "ResourceValidator",
	//	Sync: func(ctx context.Context, resource *v1alpha1.MonoRepository) error {
	//		log := util.L(ctx)
	//		// resolve the input
	//		key := resource.Spec.SourceRef.Key(resource.ObjectMeta.Namespace)
	//
	//		component := GetKind(resource.Spec.SourceRef.APIVersion, resource.Spec.SourceRef.Kind)
	//
	//		err := c.Client.Get(ctx, key, component)
	//		if err != nil {
	//			if errors.IsNotFound(err) {
	//				log.Info("unable to resolve", "name", key.Name, "namespace", key.Namespace)
	//				resource.Status.MarkResourceMissing(key.Name, key.Name, key.Namespace)
	//			} else {
	//				log.Error(err, "error resolving", "name", key.Name, "namespace", key.Namespace)
	//				resource.Status.MarkFailed(err)
	//			}
	//			return nil
	//		}
	//
	//		// parse the status
	//		artifact, err := GetArtifact(component)
	//		if err != nil {
	//			log.Error(err, "error finding artifact", "name", key.Name, "namespace", key.Namespace)
	//			resource.Status.MarkFailed(err)
	//		}
	//
	//		log.Info("got", "artifact", artifact)
	//		resource.Status.MarkArtifactResolved(artifact.URL)
	//
	//		stashArtifact(ctx, artifact)
	//
	//		return nil
	//	},
	//}
}

const artifactKey reconcilers.StashKey = "artifact"

//func stashArtifact(ctx context.Context, artifact v1alpha1.Artifact) {
//	reconcilers.StashValue(ctx, artifactKey, artifact)
//}
//
//func retrieveArtifact(ctx context.Context) v1alpha1.Artifact {
//	if components, ok := reconcilers.RetrieveValue(ctx, artifactKey).(v1alpha1.Artifact); ok {
//		return components
//	}
//
//	return v1alpha1.Artifact{}
//}

//func NewChecksumCalculator(c reconcilers.Config) reconcilers.SubReconciler[*v1alpha1.MonoRepository] {
//	return &reconcilers.SyncReconciler[*v1alpha1.MonoRepository]{
//		Name: "ChecksumCalculator",
//		Sync: func(ctx context.Context, resource *v1alpha1.MonoRepository) error {
//			log := util.L(ctx)
//
//			artifact := retrieveArtifact(ctx)
//			if artifact.URL != "" {
//				// create a temporary directory
//				tempDir, err := os.MkdirTemp("", "tmp")
//				if err != nil {
//					return err
//				}
//
//				// cleanup on exit
//				defer os.RemoveAll(tempDir)
//
//				log.Info("created temp dir", "dir", tempDir)
//
//				// download the filter and copy from/to path
//				tarGzLocation := filepath.Join(tempDir, fmt.Sprintf("%s.tar.gz", resource.Spec.SourceRef.Name))
//				err = util.DownloadFile(tarGzLocation, artifact.URL)
//				if err != nil {
//					return err
//				}
//
//				// extract tar.gz to temp location
//				tarGzExtractedLocation := filepath.Join(tempDir, fmt.Sprintf("%s-extracted", resource.Spec.SourceRef.Name))
//				err = util.ExtractTarGz(tarGzLocation, tarGzExtractedLocation)
//				if err != nil {
//					return err
//				}
//
//				files, err := util.ListFiles(tarGzExtractedLocation)
//				if err != nil {
//					return err
//				}
//
//				log.Info("Full file list", "files", files)
//
//				filteredFiles := util.FilterFileList(files, resource.Spec.Include)
//				log.Info("Using files for checksum calculation", "files", filteredFiles)
//				resource.Status.ObservedFileList = strings.Join(filteredFiles, "\n")
//
//				hash, err := dirhash.Hash1(filteredFiles, func(name string) (io.ReadCloser, error) {
//					return os.Open(filepath.Join(tarGzExtractedLocation, name))
//				})
//				if err != nil {
//					return err
//				}
//
//				log.Info("Calculated checksum", "checksum", hash)
//
//				if resource.Status.Artifact != nil && resource.Status.Artifact.Checksum == hash {
//					// nothing has changed, do nothing
//					log.Info("Source hasn't changed, there is nothing to update",
//						"name", resource.Spec.SourceRef.Name,
//						"kind", resource.Spec.SourceRef.Kind,
//						"apiVersion", resource.Spec.SourceRef.APIVersion)
//				} else {
//					old := "<NA>"
//					if resource.Status.Artifact != nil {
//						old = resource.Status.Artifact.Checksum
//					}
//
//					log.Info("Source has changed! updating status with new checksum",
//						"checksum", hash,
//						"old", old)
//					resource.Status.Artifact = &v1alpha1.Artifact{
//						Path:           artifact.Path,
//						URL:            artifact.URL,
//						Revision:       artifact.Revision,
//						Checksum:       hash,
//						Digest:         artifact.Digest,
//						LastUpdateTime: artifact.LastUpdateTime,
//						Size:           artifact.Size,
//						Metadata:       artifact.Metadata,
//					}
//					resource.Status.URL = artifact.URL
//				}
//
//				resource.Status.ObservedInclude = resource.Spec.Include
//				resource.Status.MarkReady()
//			}
//
//			return nil
//		},
//	}
//}
