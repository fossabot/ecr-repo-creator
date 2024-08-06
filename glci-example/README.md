# OCI image build component

## TL;DR

1. Adds the following in your `.gitlab.ci-yaml`
    ```yaml
    include:
      - component: $CI_SERVER_HOST/ci/components/oci-image/build@v1
    ```
2. get an oci image in `<REGISTRY_URL>/dist/images/$CI_SERVER_HOST/<project-path>`


## Intro

This components helps project to build OCI images.
Default output image will be of the form: `<ECR-URL>/dist/images/$CI_SERVER_HOST/<project-path>`, component name will be appended in case of multi-component.

Pipeline will trigger on web initiated pipelines, `develop` branch and semver tags `v1.2.3` (`v` is optional, release is supported but not build due to tag constraints).
By default, semver tags `1.2.3` will be also declined as `1.2` and `1`, with release if provided. (see options to disable those behaviors)

## Basic usage

In your `.gitlab-ci.yaml` add the following lines:
```yaml
include:
  - component: $CI_SERVER_HOST/ci/components/oci-image/build@v1
```

All default values are sensible for the platform but can be overridden. It requires an image with `jq` and `ecr-repo-creator`.

### Component inputs

#### Standard options

- `tags`: runner tags, default: `shared-services`
- `stage`: stage to register the job in, default: `build`
- `parallel_matrix`: job parallelization definition, defaults to a single job
- `needs`: job needs, default none (`[]`)
- `dependencies`: job needs, default none (`[]`)
- `rules`: trigger rules, default to semver tag, develop branch and web pipelines

#### Kaniko specific args

- `cache`: enable or disable the cache (default `true`)
- `cache_copy_layers`: when cache is enabled, cache `COPY` layers, default `true`
- `cache_run_layers`: when cache is enabled, cache `RUN` layers, default `true`
- `use_new_run`: use new `RUN` implementation default: `true` [see doc for more details](https://github.com/GoogleContainerTools/kaniko?tab=readme-ov-file#flag---use-new-run)
- `build_args`: build arg to pass, defined as an array of string, no need to quote more than YAML requires, ex ["MY_BUILD_ARG=foo bar"], default: `[]`
- `build_args_w_subst`: build arg to pass to kaniko, through [a8m's `envsubst`](https://github.com/a8m/envsubst).
    Like `build_args` but with substitution, evaluated just before kaniko's call in the job's context, ex ["CI_JOB_STARTED_AT=$CI_JOB_STARTED_AT"], default: `[]` \
    Can be useful to break cache chain on purpose at a specific operation during build by using a changing variable as `ARG` in the middle of the `Dockerfile` like `ARG CI_JOB_STARTED_AT` (all subsequent calls to `COPY` / `RUN` won't rely on cache, [see kaniko doc about cache](https://github.com/GoogleContainerTools/kaniko?tab=readme-ov-file#caching))

#### Options for multiple inclusion

- `job_name`: override job name
- `env_var_prefix`: prefix for environment variables to disambiguate when component is included times, default to `OCI_BUILD`


### Environment variables

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
