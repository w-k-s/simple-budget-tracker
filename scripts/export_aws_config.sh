#!/bin/bash

export_repo_name(){
    export ECR_REPOSITORY_URI=$(aws ecr describe-repositories --repository-names simple-budget-tracker | jq -r '.repositories[0].repositoryUri');
}

export_repo_name;