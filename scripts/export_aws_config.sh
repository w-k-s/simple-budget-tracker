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

export_private_subnets(){
    export PRIVATE_SUBNETS=$(aws ec2 describe-subnets --filters Name=vpc-id,Values=$VPC_ID Name=tag:Public,Values=0 --query 'Subnets[].{subnetId: SubnetId}[].subnetId' | jq -r '. | join(",")')
}

export_rds_security_group(){
    export RDS_SECURITY_GROUP=$(aws ec2 describe-security-groups --filters Name=vpc-id,Values=$VPC_ID Name=tag:ProjectName,Values=$PROJECT_NAME Name=tag:Resource,Values=RDS | jq -r '.SecurityGroups[0].GroupId')
}

update_record_set(){
    if [[ -z "${HOSTED_ZONE_NAME}" ]]; then 
        echo 'Expected an environment variable named "HOSTED_ZONE_NAME" with a value like "example.com." (Note: the dot at the end)';
        exit 1;
    fi

    if [[ -z "${RECORD_SET_DOMAIN_NAME}" ]]; then
        echo "Expected an environent variable named 'RECORD_SET_DOMAIN_NAME' with a value like 'budget.example.com'";
        exit 1;
    fi

    export HOSTED_ZONE_ID=$(aws route53 list-hosted-zones | jq -r '.HostedZones[] | select(.Name == "$HOSTED_ZONE_NAME") | .Id') | sed -e "s/^\/hostedzone\///"
    export RECORD_SET_TYPE="CNAME"
    export RECORD_SET_VALUE="$(aws cloudformation describe-stacks --stack-name $PROJECT_NAME --query "Stacks[?StackName=='$PROJECT_NAME'].Outputs[?OutputKey=='DNSName'].OutputValue" --output text )"
    
    envsubst < .change-resource-record-set.tmpl > change-resource-record-sets.json
    aws route53 change-resource-record-sets --hosted-zone-id "$HOSTED_ZONE_ID" --change-batch file://change-resource-record-sets.json
}

export_repo_name;
export_image_id;
export_vpc_id;
export_public_subnets;
export_private_subnets;
export_rds_security_group;