#!/bin/bash

set -eux

kind export logs ./logs
tar -czvf kind-logs-${JOB_INDEX}.tar.gz ./logs