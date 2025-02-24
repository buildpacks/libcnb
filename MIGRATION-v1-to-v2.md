# Migration Guide

This guide highlights the major differences between libcnb v1 and v2 with a focus on what you as an author need to change to upgrade your buildpacks from v1 to v2.

## Buildpack API Support

With libcnb v1, you get support for buildpack API 0.5 to 0.8. With v2, you get support for 0.8 to 0.10. This should provide a seamless transition as you can continue to use buildpack API 0.8 with 

## Removal of LayerContributor

The LayerContributor has been removed. Previously, libcnb would take a list of LayerContributors and execute them to retrieve the list of Layers to process. Now it just directly takes the list of Layers to process.

The path forward is to either update your buildpacks to return layers or to implement your own LayerContributor interface as a means to ease the transition to libcnb v2.

## Replace Builder and Detector with Functions

The Builder and Detector interfaces have been removed and replaced with functions, specifically BuildFunc and DetectFunc. They serve the same purpose, but simplify your implementation because you do not need to implement the single method interface, you can only need pass in a function that will be called back.

The path forward is to remove your Builder and Detector structs and pass the method directly into libcnb.

## Rename `poet` to `log`

We have renamed the `poet` module to be called `log`. This is a simple find & replace change in your code. We have also simplified the method names, so instead of needing to say `poet.NewLogger`, you can say `log.New` or `log.NewWithOptions`. 

We have also introduced a new `Logger` interface which can be overridden to control how libcnb logs. A basic implementation has been provided, called `PlainLogger`. If you want to supply a custom implementation, it can be passed in through main/build/detect functions.

Lastly, libcnb has been modified such that it will only log at the `Debug` level. Anything problems that arise will instead return an error, which you as the buildpack author should handle.

## Remove Deprecated Pre-Buildpack API 0.7 BOM

The pre-buildpack API 0.7 BOM API was marked for deprecation in libcnb v1. It has been removed in v2. If you are still using it, you will need to migrate to use the new SBOM functionality provided by the [Buildpacks API](https://github.com/buildpacks/rfcs/blob/main/text/0095-sbom.md).

## Path to Source Code Changed

In libcnb v1, `BuildContext.Application.Path` points to the application source code. This was shortened to `BuildContext.ApplicationPath`. The same change was made for `DetectContext.ApplicationPath`.

## Remove Deprecated CNB Binding Support

The CNB Binding specification has long been replaced by the Service Binding Specification for Kubernetes. Support for the CNB Binding Support had remained, but it is removed in v2. This is unlikely to impact anyone.

## Remove Shell-Specific Logic & Overridable Process Arguments

To comply with [RFC #168](https://github.com/buildpacks/rfcs/pull/168), we remove the `Direct` field from the `Process` struct. We also change `Command` from a `string` to `string[]` on the `Process` struct, to support overridable process arguments. Both of these require at leats API 0.9.

In conjunction with this, we have also removed Profile & `profile.d` support. These should be replaced with the [Profile Buildpack](https://github.com/buildpacks/profile) and `exec.d`.
