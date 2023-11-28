#!/bin/bash
set -ex
branch_name=tmp_release_history
git checkout --orphan $branch_name
git reset --hard
git add $0
git commit -m "start of history branch $branch_name"
git tag|while read tag; do
    git merge --allow-unrelated-histories --no-ff $tag -m "$branch_name: $tag"
done
