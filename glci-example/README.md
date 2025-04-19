
# OCI image build component <!-- omit in toc -->

<!-- markdownlint-configure-file { "MD024": false, "MD036": false} -->
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Intro](#intro)
- [Build](#build)
  - [Usage](#usage)
    - [Build time args and labels](#build-time-args-and-labels)
    - [Component inputs](#component-inputs)
      - [Standard options](#standard-options)
      - [Kaniko/Image specific args](#kanikoimage-specific-args)
      - [Options for multiple inclusion](#options-for-multiple-inclusion)
    - [Environment variables](#environment-variables)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Intro

This components helps project to build OCI images.

Default output image will be of the form: `<ECR-URL>/dist/images/$CI_SERVER_HOST/<project-path>`, variants name will be appended in case of multi-variants.

## Build

Pipeline will trigger on web initiated pipelines, `develop` branch and semver tags `v1.2.3` (`v` is optional, release is supported but not build due to tag constraints).
By default, semver tags `1.2.3` will be also declined as `1.2` and `1`, with release if provided. (see options to disable those behaviors)

### Usage

In your `.gitlab-ci.yaml` add the following lines:

```yaml
include:
  - component: $CI_SERVER_HOST/ci/components/oci-image/build@v1
```

All default values are sensible for the platform but can be overridden. It requires an image with `kaniko`, `jq`, `envsubst` and `ecr-repo-creator`.

#### Build time args and labels

Some build related environment variables are available from inside the `Dockerfile` during the build process:

- `MAIN_TAG`: the main tag of the image
- `ALL_TAGS`: all the tags of the image
- `VARIANT`: variant of the image if exists
- `CI_COMMIT_TAG`: commit tag if present
- `CI_COMMIT_REF_NAME`: branch or tag name for which project is built
- `CI_JOB_URL`: job details URL
- `CI_JOB_STARTED_AT`: date and time when a job started
- `CI_PROJECT_URL`: URL of the project

Those values are also saved as `label`, they are supplemented with some [standard annotations](https://github.com/opencontainers/image-spec/blob/main/annotations.md):

- `org.opencontainers.image.source`: URL to get source code for building the image (`CI_PROJECT_URL`)
- `org.opencontainers.image.created`: date and time on which the image was built, conforming to RFC 3339 (`CI_JOB_STARTED_AT`)
- `org.opencontainers.image.version`: version of the packaged software (`CI_COMMIT_REF_NAME`)
- `org.opencontainers.image.revision`: Source control revision identifier for the packaged software (`CI_COMMIT_SHA`)

#### Component inputs

##### Standard options

- `runner_tags`: runner tags, default: `shared-services`
- `stage`: stage to register the job in, default: `build`
- `variants`: build multiple variants (ie search for `Dockerfile` in `docker/build/$VARIANT/` and `$VARIANT/`), default builds root `Dockerfile`
- `dependencies`: job needs, default none (`[]`)
- `rules`: trigger rules, default to semver tag, develop branch and web pipelines

If you need `needs` keyword, use the `build-w-needs` variant adds the following input:

- `needs`: job needs, default none (`[]`)

Advanced matrix parallelization:

- `matrix_key`: specify the matrix variable name for advanced parallelization, greatly alter the VARIANT behavior, this variable will be injected in the build process as build arg and label, see implementation for details.

##### Kaniko/Image specific args

- `tag_template`: template to derive tag from when not a semver commit tag (default `${CI_COMMIT_REF_SLUG}-${CI_COMMIT_SHORT_SHA}-${CI_PIPELINE_ID}`)
- `tag_suffix`: suffix to be added to all generated tags (semver declination and `tag_template`), will be prefixed with `-` (ie: semver tag `1.2.3` with `tag_suffix` set to `debug` will end up as `1.2.3-debug` , default: `""`)
- `cache`: enable or disable the cache (default `false`)
- `cache_copy_layers`: when cache is enabled, cache `COPY` layers, default `true`
- `cache_run_layers`: when cache is enabled, cache `RUN` layers, default `true`
- `use_new_run`: use new `RUN` implementation default: `true` [see doc for more details](https://github.com/GoogleContainerTools/kaniko?tab=readme-ov-file#flag---use-new-run)
- `build_args`: build arg to pass, defined as an array of string, no need to quote more than YAML requires, ex ["MY_BUILD_ARG=foo bar"], default: `[]`
- `build_args_w_subst`: build arg to pass to kaniko, through [a8m's `envsubst`](https://github.com/a8m/envsubst).
    Like `build_args` but with substitution, evaluated just before kaniko's call in the job's context, ex ["CI_JOB_STARTED_AT=$CI_JOB_STARTED_AT"], default: `[]` \
    Can be useful to break cache chain on purpose at a specific operation during build by using a changing variable as `ARG` in the middle of the `Dockerfile` like `ARG CI_JOB_STARTED_AT` (all subsequent calls to `COPY` / `RUN` won't rely on cache, [see kaniko doc about cache](https://github.com/GoogleContainerTools/kaniko?tab=readme-ov-file#caching))

##### Options for multiple inclusion

- `job_suffix`: suffix for job name, default: `""`
- `env_var_prefix`: prefix for environment variables to disambiguate when component is included multiple times, default to `OCI_BUILD`

#### Environment variables

> [!NOTE]
> all environment variables will be searched as prefixed by `env_var_prefix` component input param (default: `OCI_BUILD`). \
> With `env_var_prefix` set to `X`, `OCI_BUILD_FORCE_TAG` become `X_FORCE_TAG`.

- `OCI_BUILD_FORCE_TAG`: Force image tag instead of the generated one
- `OCI_BUILD_ADDITIONNAL_TAGS`: Other tags to tag the image with (space separated)
- `OCI_BUILD_SEMVER_MULTITAG`: If set to `true` and tag is semver, decline tag version into multiple tags
- `OCI_BUILD_KANIKO_VERBOSITY`: change kaniko verbosity (panic|fatal|error|warn|info|debug|trace) default: `info`
- `OCI_BUILD_KANIKO_ENABLE_CACHE`: override component input's kaniko cache flag (default|true|false), default: `default`
- `OCI_BUILD_KANIKO_SET_CACHE_COPY_LAYERS`: override component input's kaniko `cache-copy-layers` flag (default|true|false), default: `default`
- `OCI_BUILD_KANIKO_SET_CACHE_RUN_LAYERS`: override component input's kaniko `cache-run-layers` flag (default|true|false), default: `default`
- `OCI_BUILD_KANIKO_SET_USE_NEW_RUN`: override component input's kaniko `use-new-run` flag (default|true|false), default: `default`
- `OCI_BUILD_KANIKO_SHOW_FLAGS`: display kaniko flags for debug (true|false), default: `false`
