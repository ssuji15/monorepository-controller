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
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/fluxcd/pkg/sourceignore"
	apiv1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/garethjevans/filter-controller/api/v1alpha1"
	"github.com/garethjevans/filter-controller/internal/util"
	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	"golang.org/x/mod/sumdb/dirhash"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

//+kubebuilder:rbac:groups=source.garethjevans.org,resources=filteredrepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=source.garethjevans.org,resources=filteredrepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=source.garethjevans.org,resources=filteredrepositories/finalizers,verbs=update
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=helmrepositories;gitrepositories;ocirepositories,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;create;update

func NewFilteredRepositoryReconciler(c reconcilers.Config) *reconcilers.ResourceReconciler[*v1alpha1.FilteredRepository] {
	return &reconcilers.ResourceReconciler[*v1alpha1.FilteredRepository]{
		Name: "FilteredRepository",
		Reconciler: reconcilers.Sequence[*v1alpha1.FilteredRepository]{
			NewResourceValidator(c),
			NewChecksumCalculator(c),
		},
		Config: c,
	}
}

func NewResourceValidator(c reconcilers.Config) reconcilers.SubReconciler[*v1alpha1.FilteredRepository] {
	return &reconcilers.SyncReconciler[*v1alpha1.FilteredRepository]{
		Name: "ResourceValidator",
		Sync: func(ctx context.Context, resource *v1alpha1.FilteredRepository) error {
			log := util.L(ctx)
			// resolve the input
			key := resource.Spec.SourceRef.Key(resource.ObjectMeta.Namespace)

			component := GetKind(resource.Spec.SourceRef.Kind)

			err := c.Client.Get(ctx, key, component)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Info("unable to resolve", "key", key)
					resource.Status.MarkResourceMissing(key.Name, key.Name, key.Namespace)
				} else {
					log.Error(err, "error resolving", "key", key)
					resource.Status.MarkFailed(err)
				}
				return nil
			}

			// parse the status
			a, ok := component.(Artifacter)
			if !ok {
				log.Info("component does not have an artifact")
			}

			artifact := a.GetArtifact()
			if artifact != nil {
				log.Info("got", "artifact", artifact)
				resource.Status.MarkArtifactResolved(artifact.URL)

				stashArtifact(ctx, artifact)
			}

			return nil
		},
	}
}

const artifactKey reconcilers.StashKey = "artifact"

func stashArtifact(ctx context.Context, artifact *apiv1beta2.Artifact) {
	reconcilers.StashValue(ctx, artifactKey, artifact)
}

func retreiveArtifact(ctx context.Context) *apiv1beta2.Artifact {
	if components, ok := reconcilers.RetrieveValue(ctx, artifactKey).(*apiv1beta2.Artifact); ok {
		return components
	}

	return nil
}

func NewChecksumCalculator(c reconcilers.Config) reconcilers.SubReconciler[*v1alpha1.FilteredRepository] {
	return &reconcilers.SyncReconciler[*v1alpha1.FilteredRepository]{
		Name: "ChecksumCalculator",
		Sync: func(ctx context.Context, resource *v1alpha1.FilteredRepository) error {
			log := util.L(ctx)

			artifact := retreiveArtifact(ctx)
			if artifact != nil {
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
				err = DownloadFile(tarGzLocation, artifact.URL)
				if err != nil {
					return err
				}

				// extract tar.gz to temp location
				tarGzExtractedLocation := filepath.Join(tempDir, fmt.Sprintf("%s-extracted", resource.Spec.SourceRef.Name))
				err = ExtractTarGz(tarGzLocation, tarGzExtractedLocation)
				if err != nil {
					return err
				}

				files, err := ListFiles(tarGzExtractedLocation)
				if err != nil {
					return err
				}

				log.Info("Full file list", "files", files)

				filteredFiles := FilterFileList(files, resource.Spec.Include)
				log.Info("Using files for checksum calculation", "files", filteredFiles)

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

func GetKind(kind string) client.Object {
	if kind == "OCIRepository" {
		return &apiv1beta2.OCIRepository{}
	} else if kind == "HelmRepository" {
		return &apiv1beta2.HelmRepository{}
	}
	return &apiv1beta2.GitRepository{}
}

type Artifacter interface {
	GetArtifact() *apiv1beta2.Artifact
}

var _ Artifacter = (*apiv1beta2.GitRepository)(nil)
var _ Artifacter = (*apiv1beta2.OCIRepository)(nil)
var _ Artifacter = (*apiv1beta2.HelmRepository)(nil)

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func ExtractTarGz(tarGzPath string, dir string) error {
	r, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}

	uncompressedStream, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(filepath.Join(dir, header.Name), 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(dir, header.Name))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			outFile.Close()

		default:
			return err
		}
	}
	return nil
}

func FilterFileList(list []string, include string) []string {
	var domain []string
	patterns := sourceignore.ReadPatterns(strings.NewReader(include), domain)
	matcher := sourceignore.NewDefaultMatcher(patterns, domain)

	var filtered []string
	for _, file := range list {
		fileParts := strings.Split(file, string(filepath.Separator))

		if matcher.Match(fileParts, false) {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

func ListFiles(dir string) ([]string, error) {
	return dirhash.DirFiles(dir, ".")
}

func HashFiles(list []string, dir string) (string, error) {
	return dirhash.Hash1(list, func(name string) (io.ReadCloser, error) {
		return os.Open(filepath.Join(dir, name))
	})
}
