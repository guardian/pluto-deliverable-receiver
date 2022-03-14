#!/bin/bash

## This simple shell script outputs a build-info YAML file that is used by pluto versions manager
## to understand where this build came from
echo ci_commit_branch: "${GITHUB_REF_NAME}" >> build-info.yaml
echo ci_commit_ref_name: "${GITHUB_REF_NAME}" >> build-info.yaml
echo ci_commit_sha: "${GITHUB_SHA}" >> build-info.yaml
echo ci_commit_timestamp: "${CI_COMMIT_TIMESTAMP}" >> build-info.yaml
echo ci_commit_title: "${CI_COMMIT_TITLE}" >> build-info.yaml
echo ci_job_url: "$GITHUB_SERVER_URL/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID" >> build-info.yaml
echo ci_project_name: "${GITHUB_REPOSITORY}" >> build-info.yaml
echo ci_merge_request_project_url: "${CI_MERGE_REQUEST_PROJECT_URL}" >> build-info.yaml
echo ci_merge_request_title: "${CI_MERGE_REQUEST_TITLE}" >> build-info.yaml
echo ci_pipeline_iid: "${CI_PIPELINE_IID}" >> build-info.yaml
echo built_image: guardianmultimedia/deliverable-receiver:$CI_PIPELINE_IID >> build-info.yaml