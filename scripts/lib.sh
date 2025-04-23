#!/usr/bin/env bash
# Common functions for Kubeflow manifest synchronization scripts

setup_error_handling() {
  set -euxo pipefail
  IFS=$'\n\t'
}

# Check if the git repository has uncommitted changes
check_uncommitted_changes() {
  if [ -n "$(git status --porcelain)" ]; then
    echo "WARNING: You have uncommitted changes"
  fi
}

# Create a new git branch if it doesn't exist
create_branch() {
  local branch="$1"
  
  check_uncommitted_changes
  
  if [ $(git branch --list "$branch") ]; then
    echo "WARNING: Branch $branch already exists."
  fi
  
  if ! git show-ref --verify --quiet refs/heads/$branch; then
    git checkout -b "$branch"
  else
    echo "Branch $branch already exists."
  fi
}

clone_and_checkout() {
  local src_dir="$1"
  local repo_url="$2"
  local repo_dir="$3"
  local commit="$4"
  
  echo "Checking out in $src_dir to $commit..."
  
  mkdir -p "$src_dir"
  cd "$src_dir"
  
  # Clone repository if it doesn't exist
  if [ ! -d "$repo_dir/.git" ]; then
    git clone "$repo_url" "$repo_dir"
  fi
  
  # Checkout to specific commit
  cd "$src_dir/$repo_dir"
  if ! git rev-parse --verify --quiet "$commit"; then
    git checkout -b "$commit"
  else
    git checkout "$commit"
  fi
  
  check_uncommitted_changes
}

# Copy manifests from source to destination
copy_manifests() {
  local src="$1"
  local dst="$2"
  
  echo "Copying manifests..."
  
  if [ -d "$dst" ]; then
    rm -r "$dst"
  fi
  
  cp "$src" "$dst" -r
  echo "Successfully copied all manifests."
}

# Update README with new commit reference
update_readme() {
  local manifests_dir="$1"
  local src_txt="$2"
  local dst_txt="$3"
  
  sed -i "s|$src_txt|$dst_txt|g" "${manifests_dir}/README.md"
}

# Commit changes to git repository
commit_changes() {
  local manifests_dir="$1"
  local commit_msg="$2"
  local paths_to_add=("${@:3}")
  
  cd "$manifests_dir"
  
  for path in "${paths_to_add[@]}"; do
    git add "$path"
  done
  
  git commit -s -m "$commit_msg"
} 