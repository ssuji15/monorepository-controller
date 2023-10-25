# monorepository-controller

a proof of concept controller used to determine if "interesting" changes have been to a flux artifact.  We currently support
`GitRepository`.

The controller works by calculating a checksum for the exploded archive based on an "include" list of files.  If this checksum changes
then the `.status.artifact.url` changes.

## Installation

```shell
# this should be a location in dev.registry that you have write access to
export REGISTRY_PROJECT=project/repo
make package install-from-package
```

Validate that the install is working:

```shell
kubectl get pkgi -n tap-install monorepository
```

## Example

Then create a `MonoRepository` that wraps a `GitRepository` e.g.

```yaml
apiVersion: source.garethjevans.org/v1alpha1
kind: MonoRepository
metadata:
  name: where-for-dinner-availability
  namespace: default
spec:
  gitRepository:
    interval: 5m
    url: https://github.com/garethjevans/where-for-dinner
    ref:
      branch: main
  include: |
    /pom.xml
    /where-for-dinner-availability
    !.*
    !**/src/test/**
```

When this resource reconciles we can see the files it used in its calculation and the checksum:

```
status:
  artifact:
    checksum: h1:BARdBbUvae7sFv+t0UdjbEcZBgVvVJimNxi8kSCGURg=
    digest: sha256:5510011685bd931bdfe6387b942366748c4428b4da49c1685c84371da9763f89
    lastUpdateTime: "2023-05-05T10:17:23Z"
    path: gitrepository/default/where-for-dinner/68d842cd330410cf0672f862d9a799af4dcdc1d7.tar.gz
    revision: main@sha1:68d842cd330410cf0672f862d9a799af4dcdc1d7
    size: 358702
    url: http://source-controller.default.svc.cluster.local./gitrepository/default/where-for-dinner/68d842cd330410cf0672f862d9a799af4dcdc1d7.tar.gz
  conditions:
  - lastTransitionTime: "2023-05-05T11:12:43Z"
    message: resolved artifact from url http://source-controller.default.svc.cluster.local./gitrepository/default/where-for-dinner/68d842cd330410cf0672f862d9a799af4dcdc1d7.tar.gz
    reason: Resolved
    status: "True"
    type: MonoRepositoryArtifactResolved
  - lastTransitionTime: "2023-05-05T10:39:50Z"
    message: Repository has been successfully filtered for change
    reason: Succeeded
    status: "True"
    type: Ready
  observedFileList: |-
    pom.xml
    where-for-dinner-availability/Tiltfile
    where-for-dinner-availability/config/workload.yaml
    where-for-dinner-availability/pom.xml
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/WhereForDinnerAvailabilityApplication.java
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/config/OAuth2BindingsPropertiesProcessor.java
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/config/WebSecurityConfig.java
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/entity/Availability.java
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/entity/AvailabilityWindow.java
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/function/AvailabilitySink.java
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/model/Availability.java
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/repository/AvailabilityRepository.java
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/repository/AvailabilityWindowRepository.java
    where-for-dinner-availability/src/main/java/com/java/example/tanzu/wherefordinner/resources/AvailabilityResource.java
    where-for-dinner-availability/src/main/resources/META-INF/spring.factories
    where-for-dinner-availability/src/main/resources/application.yaml
    where-for-dinner-availability/src/main/resources/schema-h2.sql
    where-for-dinner-availability/src/main/resources/schema-mysql.sql
    where-for-dinner-availability/src/main/resources/schema-postgresql.sql
  observedGeneration: 2
  observedInclude: |
    /pom.xml
    /where-for-dinner-availability
    !.*
    !**/src/test/**
  url: http://source-controller.default.svc.cluster.local./gitrepository/default/where-for-dinner/68d842cd330410cf0672f862d9a799af4dcdc1d7.tar.gz
```
