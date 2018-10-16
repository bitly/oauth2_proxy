#!/bin/bash
set -eo pipefail

# Determine the image and cache tags
IMAGE_TAG=${BUILDKITE_BRANCH}
CACHE_TAG=${IMAGE_TAG}

# Determine the Dockerfile location
if [ -z "$DOCKERFILE" ]; then
  DOCKERFILE="Dockerfile-buildimage"
fi

eval $(aws ecr get-login --no-include-email --region us-east-1) 

# Pull the latest branch tag for caching, if it exists
IMAGE_EXISTS=1
docker pull $IMAGE_NAME:$IMAGE_TAG || IMAGE_EXISTS=0

# If the branch image didn't already exist, pull the latest
if [ $IMAGE_EXISTS -eq 0 ]; then
  docker pull $IMAGE_NAME:latest || true
  CACHE_TAG=latest
fi

EXTRA_TAGS="--tag $IMAGE_NAME:$BUILDKITE_COMMIT"

# If the branch is master, also tag with latest
if [[ "$IMAGE_TAG" == "master" ]]; then
  EXTRA_TAGS="$EXTRA_TAGS --tag $IMAGE_NAME:latest"
fi

if [ ! -d build/public ]; then
  mkdir -p build/public
fi

git log -1 > build/public/REVISION.txt

# Retrieve our artifact
buildkite-agent artifact download oauth2_proxy .

# Build the new image
docker build \
  --network host \
  --cache-from $IMAGE_NAME:$CACHE_TAG \
  --tag $IMAGE_NAME:$IMAGE_TAG \
  $EXTRA_TAGS \
  -f $DOCKERFILE \
  .

# Push to the repository
docker push $IMAGE_NAME:$IMAGE_TAG
echo "Pushing docker image to ECR: $IMAGE_NAME:$IMAGE_TAG"

docker push $IMAGE_NAME:$BUILDKITE_COMMIT
echo "Pushing docker image to ECR: $IMAGE_NAME:$BUILDKITE_COMMIT"

# If the branch is master, also push the latest tag
if [[ "$IMAGE_TAG" == "master" ]]; then
  docker push $IMAGE_NAME:latest
  echo "Pushing docker image to ECR: $IMAGE_NAME:latest"
fi
