# Contributing Guide

The VMware Event Broker Appliance team welcomes contributions from the
community.

To help you get started making contributions to VMware Event Broker Appliance,
we have collected some helpful best practices in the [Contributing
guidelines](https://vmweventbroker.io/community#guidelines).


## Working with examples

- Create a topic branch in your fork repository where you want to base your work
- Make the changes you need and commit them
- Make sure your commit messages are in the proper format (see below) and signed
- Push your changes to a topic branch in your fork of the repository
- Submit a pull request

> Before submitting a pull request, please make sure that your change satisfies the requirements specified [here](https://vmweventbroker.io/community#pull-requests)

## Publish a new version of an example

These examples are hosted in [`ghcr.io`](https://github.com/features/packages) and we will explain how to deploy the examples to this registry but you are free to use any registry of your like.

To publish an image manually you will need a [personal github token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) (Classic) and save it in a environment variable like `$GH_CLASSIC_TOKEN`.

After you have your token in your environment you can login using

```bash
echo $GH_CLASSIC_TOKEN | ko login ghcr.io -u <your-username> --password-stdin
```

The rest of the steps depends on the tools use to build and push each example.

## Go Examples

Go-based examples use [`ko`](https://ko.build) to build and push images of the working directory of the example.

For `ko`, you need to set the base url for the repository these images are going to be part of using the `KO_DOCKER_REPO`. Assuming you have this repository forked on an account you have access to `ghcr.io/<github-account>/vcenter-event-broker-appliance`.

```bash
export KO_DOCKER_REPO=ghcr.io/<github-account>/vcenter-event-broker-appliance
```

You also need to set two environment variables for `ko`: `KO_TAG`, `KO_DOCKER_REPO`, and `KO_COMMIT`

```bash
export KO_COMMIT=$(git rev-parse --short=8 HEAD)
export KO_TAG=<new-version>
export KO_DOCKER_REPO=ghcr.io/<docker-username>/<repo-name>
```

Now you can build and push your new version of the example using `ko build`

```bash
ko build . --image-label org.opencontainers.image.source="<github-domain>/<github-username>/<repo-name>"  --tags <new-version> --tags latest -B
```

### Notes
- On GitHub, in order to attach an image to a repository we need to set the label `org.opencontainers.image.source` of the repository the package will be associated with.
- we use the flag `-B` in order to use the base path without MD5 hash after `KO_DOCKER_REPO`

## Powercli Powershell

For the Powercli examples we need to build the template that is on [knative/templates](./knative/templates/)

These templates use Dockerfiles and can build them using `docker build`. After you built the images we need to tag the images

```bash
docker image tag <template-image-name> ghcr.io/<github-username>/<repo-name>/<template-image>:latest
docker image tag <template-image-name> ghcr.io/<github-username>/<repo-name>/<template-image>:<new-version>
```

To push these images we need to run

 ```bash
 docker image push --all-tags <template-image-name>
 ```

> Notes: the PowerCLI template depends on the PowerPS image. You need to build the base before the CLI version

After you build the templates, you can now build any example on `knative/powershell` or `knative/powercli` using the same steps of the templates

## Python

[Buildpacks](https://buildpacks.io) are used to create the container image. We can build any example by using:

```bash
IMAGE=ghcr.io/<docker-username>/<repo-name>/<image-name>:<version>
pack build -B gcr.io/buildpacks/builder:v1 ${IMAGE}
```
If you are using pack@0.30.1 or newer version you can label images like 

```bash
pack buildpack package $IMAGE -l org.opencontainers.image.source="<github-domain>/<github-username>/<repo-name>"
```

If you are using pack with an older version, we need to use `gcr.io/paketo-buildpacks/image-labels:4.5.2` buildpack

```bash
pack build -B gcr.io/buildpacks/builder:v1 -b gcr.io/paketo-buildpacks/image-labels:4.5.2 $IMAGE -e BP_IMAGE_LABELS=org.opencontainers.image.source="<github-domain>/<github-username>/<repo-name>"
```

and then we can push the image

```bash
docker push $IMAGE
```

# Testing your image

Now you build and push the example with new changes you can test it by replacing the value of the image in the `function.yaml` file and following the steps of the example of testing its usage.
