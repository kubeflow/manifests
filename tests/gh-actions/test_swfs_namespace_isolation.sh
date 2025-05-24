#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}SeaweedFS Security Test - Unauthorized Access Check${NC}"
echo "Testing if one namespace can access files from another namespace"

# Check dependencies
for cmd in kubectl aws jq; do
    if ! command -v $cmd &> /dev/null; then
        echo -e "${RED}Error: $cmd is required but not installed${NC}"
        exit 1
    fi
done

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    [ -n "$PORT_FORWARD_PID" ] && kill $PORT_FORWARD_PID 2>/dev/null || true
    rm -f test-file.txt accessed-file.txt
    kubectl delete profile test-profile-1 test-profile-2 --ignore-not-found
}
trap cleanup EXIT

# Create test profiles
create_profiles() {
    echo -e "\n${YELLOW}Creating test profiles...${NC}"

    # Create both profiles
    kubectl apply -f - <<EOF
apiVersion: kubeflow.org/v1
kind: Profile
metadata:
  name: test-profile-1
  labels:
    pipelines.kubeflow.org/enabled: "true"
spec:
  owner:
    kind: User
    name: test-user-1@example.com
---
apiVersion: kubeflow.org/v1
kind: Profile
metadata:
  name: test-profile-2
  labels:
    pipelines.kubeflow.org/enabled: "true"
spec:
  owner:
    kind: User
    name: test-user-2@example.com
EOF

    # Wait for namespaces
    echo "Waiting for namespaces..."
    for i in {1..30}; do
        if kubectl get namespace test-profile-1 test-profile-2 >/dev/null 2>&1; then
            echo -e "${GREEN}Namespaces created${NC}"
            return 0
        fi
        echo "Waiting... ($i/30)"
        sleep 5
    done

    echo -e "${RED}Error: Namespaces not created${NC}"
    exit 1
}

# Wait for S3 credentials
wait_for_credentials() {
    local namespace=$1
    echo "Waiting for S3 credentials in $namespace..."

    for i in {1..60}; do
        if kubectl get secret -n $namespace mlpipeline-minio-artifact >/dev/null 2>&1; then
            echo -e "${GREEN}Credentials found${NC}"
            return 0
        fi
        echo "Waiting... ($i/60)"
        sleep 5
    done

    echo -e "${RED}Error: No credentials found${NC}"
    return 1
}

# Get credentials for namespace
get_credentials() {
    local namespace=$1
    local access_key=$(kubectl get secret -n $namespace mlpipeline-minio-artifact -o jsonpath='{.data.accesskey}' | base64 -d)
    local secret_key=$(kubectl get secret -n $namespace mlpipeline-minio-artifact -o jsonpath='{.data.secretkey}' | base64 -d)
    echo "$access_key:$secret_key"
}

# Setup port forward to SeaweedFS
setup_port_forward() {
    if [ -n "$PORT_FORWARD_PID" ]; then
        return 0  # Already running
    fi

    echo "Setting up port-forward..."
    local pod=$(kubectl get pod -n kubeflow -l app=seaweedfs -o jsonpath='{.items[0].metadata.name}')
    kubectl port-forward -n kubeflow pod/$pod 8333:8333 >/dev/null 2>&1 &
    PORT_FORWARD_PID=$!
    sleep 3
}

# Upload test file
upload_file() {
    local namespace=$1
    echo -e "\n${YELLOW}Uploading test file to $namespace...${NC}"

    local creds=$(get_credentials $namespace)
    local access_key=$(echo $creds | cut -d: -f1)
    local secret_key=$(echo $creds | cut -d: -f2)

    setup_port_forward

    echo "Test file for $namespace" > test-file.txt

    AWS_ACCESS_KEY_ID=$access_key \
    AWS_SECRET_ACCESS_KEY=$secret_key \
    AWS_ENDPOINT_URL=http://localhost:8333 \
    aws s3 cp test-file.txt s3://mlpipeline/private-artifacts/$namespace/test-file.txt

    rm -f test-file.txt
}

# Test unauthorized access
test_unauthorized_access() {
    local from_namespace=$1
    local target_namespace=$2

    echo -e "\n${YELLOW}Testing unauthorized access from $from_namespace to $target_namespace...${NC}"

    local creds=$(get_credentials $from_namespace)
    local access_key=$(echo $creds | cut -d: -f1)
    local secret_key=$(echo $creds | cut -d: -f2)

    setup_port_forward

    # Try to access the other namespace's file
    if AWS_ACCESS_KEY_ID=$access_key \
       AWS_SECRET_ACCESS_KEY=$secret_key \
       AWS_ENDPOINT_URL=http://localhost:8333 \
       aws s3 cp s3://mlpipeline/private-artifacts/$target_namespace/test-file.txt ./accessed-file.txt 2>/dev/null; then

        echo -e "${RED}SECURITY ISSUE: Unauthorized access successful!${NC}"
        echo "File contents:"
        cat ./accessed-file.txt
        rm -f ./accessed-file.txt
        return 1
    else
        echo -e "${GREEN}Security OK: Access denied as expected${NC}"
        return 0
    fi
}

# Main test function
main() {
    echo -e "\n${YELLOW}Starting security test...${NC}"

    # Create test profiles
    create_profiles

    # Wait for credentials to be created
    echo "Waiting for profile controller to create credentials..."
    sleep 30

    wait_for_credentials "test-profile-1" || {
        echo -e "${RED}Failed to get credentials for test-profile-1${NC}"
        exit 1
    }

    wait_for_credentials "test-profile-2" || {
        echo -e "${RED}Failed to get credentials for test-profile-2${NC}"
        exit 1
    }

    # Upload file to first namespace
    upload_file "test-profile-1" || {
        echo -e "${RED}Failed to upload file${NC}"
        exit 1
    }

    # Test unauthorized access
    if test_unauthorized_access "test-profile-2" "test-profile-1"; then
        echo -e "\n${GREEN}SECURITY TEST PASSED: No unauthorized access detected${NC}"
    else
        echo -e "\n${RED}SECURITY TEST FAILED: Unauthorized access detected${NC}"
        echo "This indicates a security vulnerability in the SeaweedFS setup"
        exit 1
    fi
}

main
