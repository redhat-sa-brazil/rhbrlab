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

    echo "> get the user id on SSO"
    sso_user_id=$($KCADM_HOME/kcadm.sh get users -r cloud -q username="${CUSTOMER}-user-${idx}" --config $KCADM_HOME/.keycloak/kcadm.config | grep "id" | head -1 | cut -d ':' -f 2 | sed -n "s/\s*\"\(\S*\)\"\,/\1/p")    
    echo -e "\t user ${CUSTOMER}-user-${idx} has the id ${sso_user_id} on sso..."

    echo "> deleting users/$sso_user_id on RH-SSO's cloud realm"
    $KCADM_HOME/kcadm.sh delete users/$sso_user_id \
     -r cloud \
     --config $KCADM_HOME/.keycloak/kcadm.config 
    echo -e "\t user ${CUSTOMER}-user-${idx} with id ${sso_user_id} removed from sso..."

    echo -e "\t deleting user $ocp_user_sso_id identity"
    ocp_user_sso_id=$(oc get user ${CUSTOMER}-user-${idx} | grep -v NAME| awk '{ print $NF}')
    #sso_user_id=$(echo $ocp_user_sso_id | cut -d ':' -f2)
    oc delete --ignore-not-found identity "${ocp_user_sso_id}"
    oc delete --ignore-not-found user ${CUSTOMER}-user-${idx}

    echo -e "\t deleting ${CUSTOMER}-user-${idx}'s project"
    oc delete --ignore-not-found project ${CUSTOMER}-user-${idx}
    echo -e "---\n"
done
