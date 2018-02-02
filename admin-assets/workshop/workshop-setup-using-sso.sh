#!/bin/bash

usage="Uso:       $0 customer_name number_of_attenddes  "

if [ $# -lt 2 ] ; then
    clear
    echo -e "\n\n\t $usage \n\n"
    exit 1
fi

OCP_API_SERVER_URL='https://console.paas.rhbrlab.com'
KEYCLOAK_SERVER_URL='https://login.apps.paas.rhbrlab.com/auth'
#WORKSHOP_TYPE=${1:-ocp}
CUSTOMER=${1:-workshop}
DEFAULT_TEMP_PWD='abc@123'
NUMBER_OF_ATTENDDES=${2:-1}
KCADM_HOME=/home/cloud-user/admin-assets/workshop/rh-sso-client

echo "Login on OCP Cluster"
#oc login -u system:admin > /dev/null
oc login -u admin --server $OCP_API_SERVER_URL

echo "Login on RH-SSO"
$KCADM_HOME/kcadm.sh config credentials \
 --server $KEYCLOAK_SERVER_URL \
 --realm master \
 --user admin \
 --config $KCADM_HOME/.keycloak/kcadm.config

for idx in $( seq $NUMBER_OF_ATTENDDES ); do

    echo -e "\ncreating user ${CUSTOMER}-user-${idx} on RH-SSO' cloud realm"
    $KCADM_HOME/kcadm.sh create users -r cloud \
     -s username=${CUSTOMER}-user-${idx} \
     -s firstName=user-${idx} \
     -s lastName=${CUSTOMER} \
     -s enabled=true \
     -s emailVerified=true \
     --config $KCADM_HOME/.keycloak/kcadm.config 

    $KCADM_HOME/kcadm.sh set-password -r cloud \
     --username ${CUSTOMER}-user-${idx} -p ${DEFAULT_TEMP_PWD} \
     --temporary \
     --config $KCADM_HOME/.keycloak/kcadm.config
 
    echo "creating ${CUSTOMER}-user-${idx}'s project"
    oc adm new-project ${CUSTOMER}-user-${idx} \
      --admin=${CUSTOMER}-user-${idx} \
      --display-name="[WORKSHOP Project] ${CUSTOMER}-user-${idx}" \
      --description="[WORKSHOP] This project will be removed at the end of our workshop!"

    echo -e "\t apply the 'rhbrlab-type=workshop' label"
    oc patch namespace ${CUSTOMER}-user-${idx} -p '{"metadata":{"labels":{"rhbrlab-type":"workshop"}}}'
    
    echo -e "\tapplying limits and quotas for ${CUSTOMER}-user-${idx}..."
    oc create -f workshop-limitrange.yaml -n ${CUSTOMER}-user-${idx}
    oc create -f workshop-quota.yaml -n ${CUSTOMER}-user-${idx}
    
done
