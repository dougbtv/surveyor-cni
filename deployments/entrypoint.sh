#!/bin/bash

# Always exit on errors.
set -e

# Trap sigterm
function exitonsigterm() {
  echo "Trapped sigterm, exiting."
  exit 0
}
trap exitonsigterm SIGTERM

# Make a surveyor.d directory (for our kubeconfig)
mkdir -p /host/etc/cni/net.d/surveyor.d
SURVEYOR_KUBECONFIG=/host/etc/cni/net.d/surveyor.d/surveyor.kubeconfig
SURVEYOR_TEMP_KUBECONFIG=/host/etc/cni/net.d/surveyor.d/tmp.surveyor.kubeconfig

# ------------------------------- Generate a "kube-config"
# Inspired by: https://tinyurl.com/y7r2knme
SERVICE_ACCOUNT_PATH=/var/run/secrets/kubernetes.io/serviceaccount
SERVICE_ACCOUNT_TOKEN_PATH=$SERVICE_ACCOUNT_PATH/token
KUBE_CA_FILE=${KUBE_CA_FILE:-$SERVICE_ACCOUNT_PATH/ca.crt}

LAST_SERVICEACCOUNT_MD5SUM=""
LAST_KUBE_CA_FILE_MD5SUM=""

function generateKubeConfig {

  # Check if we're running as a k8s pod.
  if [ -f "$SERVICE_ACCOUNT_TOKEN_PATH" ]; then
    # We're running as a k8d pod - expect some variables.
    if [ -z ${KUBERNETES_SERVICE_HOST} ]; then
      error "KUBERNETES_SERVICE_HOST not set"; exit 1;
    fi
    if [ -z ${KUBERNETES_SERVICE_PORT} ]; then
      error "KUBERNETES_SERVICE_PORT not set"; exit 1;
    fi

    if [ "$SKIP_TLS_VERIFY" == "true" ]; then
      TLS_CFG="insecure-skip-tls-verify: true"
    elif [ -f "$KUBE_CA_FILE" ]; then
      TLS_CFG="certificate-authority-data: $(cat $KUBE_CA_FILE | base64 | tr -d '\n')"
    fi

    # Get the contents of service account token.
    SERVICEACCOUNT_TOKEN=$(cat $SERVICE_ACCOUNT_TOKEN_PATH)

    SKIP_TLS_VERIFY=${SKIP_TLS_VERIFY:-false}

    # Write a kubeconfig file for the CNI plugin.  Do this
    # to skip TLS verification for now.  We should eventually support
    # writing more complete kubeconfig files. This is only used
    # if the provided CNI network config references it.
    touch $SURVEYOR_TEMP_KUBECONFIG
    chmod ${KUBECONFIG_MODE:-600} $SURVEYOR_TEMP_KUBECONFIG
    # Write the kubeconfig to a temp file first.
    timenow=$(date)
    cat > $SURVEYOR_TEMP_KUBECONFIG <<EOF
# Kubeconfig file for Multus CNI plugin.
# Generated at ${timenow}
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: ${KUBERNETES_SERVICE_PROTOCOL:-https}://[${KUBERNETES_SERVICE_HOST}]:${KUBERNETES_SERVICE_PORT}
    $TLS_CFG
users:
- name: surveyor-cni
  user:
    token: "${SERVICEACCOUNT_TOKEN}"
contexts:
- name: surveyor-cni-context
  context:
    cluster: local
    user: surveyor-cni
current-context: surveyor-cni-context
EOF

    # Atomically move the temp kubeconfig to its permanent home.
    mv -f $SURVEYOR_TEMP_KUBECONFIG $SURVEYOR_KUBECONFIG

    # Keep track of the md5sum
    LAST_SERVICEACCOUNT_MD5SUM=$(md5sum $SERVICE_ACCOUNT_TOKEN_PATH | awk '{print $1}')
    LAST_KUBE_CA_FILE_MD5SUM=$(md5sum $KUBE_CA_FILE | awk '{print $1}')

  else
    warn "Doesn't look like we're running in a kubernetes environment (no serviceaccount token)"
  fi

# ---------------------- end Generate a "kube-config".

}

generateKubeConfig

# Watch a socket file.
# You can test writing to this with: 
# echo "quux" | sudo socat - UNIX-CONNECT:/var/run/surveyor-cni/surveyor.sock
rm -f /host/var/run/surveyor-cni/surveyor.sock
echo "Listening on socket..."
nc -lkU /host/var/run/surveyor-cni/surveyor.sock