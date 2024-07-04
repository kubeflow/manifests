
# CVE Scanning Process Documentation for Kubeflow 1.9

## Overview

This document provides an overview of the security practices implemented by the security working group for Kubeflow 1.9. It details the recurring CVE scanning process, the use of Trivy, the scheduling and storage of scan results, and policies regarding result retention and disclosure.

## Setup Process

### Prerequisites
- Ensure you have access to the Kubeflow repository on GitHub.
- Install [Trivy](https://github.com/aquasecurity/trivy), a comprehensive and easy-to-use vulnerability scanner for container images.

### GitHub Actions Workflow
The CVE scanning process is automated using GitHub Actions. The action is triggered on each merge to the master branch. The workflow file that sets up this process can be found at [.github/workflows/trivy.yaml](https://github.com/kubeflow/manifests/blob/master/.github/workflows/trivy.yaml).

### Trivy Scan Script
The Trivy scan is performed by a Python script located at [hack/trivy_scan.py](https://github.com/kubeflow/manifests/blob/master/hack/trivy_scan.py). This script is invoked by the GitHub action.

### Running the Process
1. The GitHub Action is triggered on each merge to the master branch.
2. The action runs the Trivy scan script.
3. Trivy scans the container images for known vulnerabilities.
4. The results are compiled into a report and stored.

## Using Trivy

Trivy is used to scan container images for Common Vulnerabilities and Exposures (CVEs). It identifies vulnerabilities in operating system packages (Alpine, RHEL, CentOS, etc.) and application dependencies (Bundler, Composer, npm, yarn, etc.).

### Scan Execution
The scan is executed by the GitHub Action defined in the workflow file. The Trivy script ([hack/trivy_scan.py](https://github.com/kubeflow/manifests/blob/master/hack/trivy_scan.py)) is responsible for initiating the scan and processing the results.

## Scan Frequency and Results Storage

### Frequency
The CVE scanning process runs on each merge to the master branch of the Kubeflow repository. This ensures that any new code integrated into the master branch is immediately scanned for vulnerabilities.

### Results Storage
The results of each scan are stored as part of the GitHub Actions job. The results can be viewed at the bottom of the job run output, such as in this example: [GitHub Actions Run](https://github.com/kubeflow/manifests/actions/runs/9744982451/job/26891834585).

## Retention Policy

### Results Retention
The scan results are retained in the GitHub Actions logs. The logs are retained according to GitHub's standard retention policy, which is typically 90 days for public repositories. 

### Policy for Disclosing CVE Results
The Kubeflow security working group follows a responsible disclosure policy for CVE results:

- **Internal Review**: All CVE findings are initially reviewed internally by the security working group.
- **Severity Assessment**: Each CVE is assessed for severity and potential impact on the Kubeflow project.
- **Disclosure**: For high and critical severity CVEs, the security working group will:
  - Notify the maintainers and contributors.
  - Provide a fix or mitigation strategy.
  - Publicly disclose the CVE details once a fix is available, ensuring users can take necessary actions to secure their deployments.
