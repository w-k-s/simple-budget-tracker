#!/bin/bash

PROJECT_NAME=simple-budget-tracker

export_repo_name(){
    export ECR_REPOSITORY_URI=$(aws ecr describe-repositories --repository-names $PROJECT_NAME | jq -r '.repositories[0].repositoryUri');
}

export_image_id(){
    export IMAGE_ID=$(aws ecr describe-images --repository-name $PROJECT_NAME --query 'sort_by(imageDetails,& imagePushedAt)[-1]' | jq -r ".imageTags[0]")
}

export_vpc_id(){
    export VPC_ID=$(aws ec2 describe-vpcs --filters Name=tag:ProjectName,Values=simple-budget-tracker | jq -r '.Vpcs[0].VpcId')
}

export_public_subnets(){
    export PUBLIC_SUBNETS=$(aws ec2 describe-subnets --filters Name=vpc-id,Values=$VPC_ID Name=tag:Public,Values=1 --query 'Subnets[].{subnetId: SubnetId}[].subnetId' | jq -r '. | join(",")')
}

export_rds_security_group(){
    export RDS_SECURITY_GROUP=$(aws ec2 describe-security-groups --filters Name=vpc-id,Values=$VPC_ID Name=tag:ProjectName,Values=$PROJECT_NAME Name=tag:Resource,Values=RDS | jq -r '.SecurityGroups[0].GroupId')
}

export_hosted_zone_id(){
    export HOSTED_ZONE_ID=$(aws route53 list-hosted-zones | jq -r '.HostedZones[] | select(.Name == "$DOMAIN_NAME") | .Id') | sed -e "s/^\/hostedzone\///"
}

export_repo_name;
export_image_id;
export_vpc_id;
export_public_subnets;
export_rds_security_group;
export_hosted_zone_id;