#!/usr/bin/env bash
set -euo pipefail

BUCKET_NAME="${BUCKET_NAME:?BUCKET_NAME is not set}"
gsutil mb -l US-WEST2 "gs://$BUCKET_NAME" >&2
gsutil versioning set on "gs://$BUCKET_NAME" >&2
echo $BUCKET_NAME
