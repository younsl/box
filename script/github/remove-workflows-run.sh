#!/bin/bash

# for reference path to your code repository:
# https://github.com/$OWNER/$REPOSITORY
export OWNER=<johndoe>
export REPOSITORY=<johndoe-nodejs-demo>

# repeat command 30 times, if there are 30 pages of workflow history 
for i in {1..30}; do
  gh api -X GET /repos/$OWNER/$REPOSITORY/actions/runs \
    | jq '.workflow_runs[] | .id' \
    | xargs -t -I{} gh api --silent -X DELETE /repos/$OWNER/$REPOSITORY/actions/runs/{};
done
