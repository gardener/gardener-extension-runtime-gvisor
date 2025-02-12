#!/usr/bin/env python3

import git
import os.path
import sys
import logging
import json

import urllib3

import ci.util
import ccc.github

import gitutil

GARDENER_EXTENSION_GVISOR_OWNER      = "gardener"
GARDENER_EXTENSION_GVISOR_REPOSITORY = "gardener-extension-runtime-gvisor"
GARDENER_EXTENSION_GVISOR_REPO_URL   = "https://github.com/gardener/gardener-extension-runtime-gvisor"

GVISOR_OWNER      = "google"
GVISOR_REPOSITORY = "gvisor"

logger = logging.getLogger(__name__)


def get_repo_root() -> str:
    file_in_repo = sys.argv[0]
    repo = git.Repo(file_in_repo, search_parent_directories=True)
    repo_root = repo.git.rev_parse("--show-toplevel")
    return repo_root


def get_current_version(repo_root: str) -> str:
    version_file = os.path.sep.join([repo_root, "GVISOR_VERSION"])
    with open(version_file, "r") as f:
        while line := f.readline():
            if line.startswith('#'):
                continue
            return line


def set_current_version(repo_root: str, version: str):
    version_file = os.path.sep.join([repo_root, "GVISOR_VERSION"])
    with open(version_file, "w") as f:
        f.writelines([f"{version}\n"])


def get_upstream_version_from_tags(organization: str, repository: str) -> str:
    api_root = 'api.github.com'
    url = f"https://{api_root}/repos/{organization}/{repository}/tags"

    resp=urllib3.request("GET", url)
    if resp.status != 200:
        raise RuntimeError(f"GitHub API did not return 200 status")
    
    raw_upstream_releases = json.loads(resp.data)

    releases = []
    for r in raw_upstream_releases:
        releases.append(r['name'].removeprefix('release-'))

    if len(releases) == 0:
        raise RuntimeError(f"failed to obtain a list of releases from {url}")
    
    releases.sort(reverse=True)

    return releases[0]


def get_git_helper(owner: str, repository:str, github_repo_url: str, repo_dir: str) -> gitutil.GitHelper:
    try:
        github_repo_path = ci.util.check_env('SOURCE_GITHUB_REPO_OWNER_AND_NAME')
    except ci.util.Failure:
        github_repo_path = f"{owner}/{repository}"

    github_cfg = ccc.github.github_cfg_for_repo_url(github_repo_url)

    git_helper=gitutil.GitHelper(
        repo=repo_dir,
        github_cfg=github_cfg,
        github_repo_path=github_repo_path,
    )

    try:
        git_helper.fetch_head('+refs/heads/*:refs/remotes/origin/*')
    except git.GitCommandError as e:
        logger.error("Failed to fetch from remote %s: %s", GARDENER_EXTENSION_GVISOR_REPO_URL, e)
        raise e
    
    return git_helper


def branch_exists(git_helper, branch: str) -> bool:
    """
    Check if a branch already exists in the remote repository.
    """
    existing_branches = [ref.remote_head for ref in git_helper.repo.remotes.origin.refs]
    return branch in existing_branches


def commit_and_push_to_branch(git_helper:gitutil.GitHelper, branch: str, commit_msg: str) -> str:
    """
    Commit and push changes to a specific branch.
    """
    try:
        # Check if branch already exists in remote.
        logger.info("Checking if remote branch %s exists...", branch)
        if branch_exists(git_helper, branch):
            logger.warning("Remote branch with name %s already exists, skipping push.", branch)
            git_helper.repo.git.reset('--hard')
            return "BranchExists"

        commit = git_helper.index_to_commit(message=commit_msg)
        head_sha = git_helper.repo.head.object.hexsha
        git_helper.repo.create_head(f'refs/heads/{branch}', head_sha)
        git_helper.push(from_ref=commit.hexsha, to_ref=f'refs/heads/{branch}')
        return "CommitPushed"
    except Exception as e:
        logger.error("Failed to commit or push: %s", str(e))
        raise e


def create_pull_request(git_helper: gitutil.GitHelper, org: str, name: str, pr_title: str, pr_body: str, branch: str):
    """
    Creates a pull request on a GitHub repository if no open pull request with the same title exists.
    """

    try:
        github = ccc.github.github_api(github_cfg=git_helper.github_cfg)
        repo = github.repository(org, name)
        open_prs = repo.pull_requests(state='open')
        for pr in open_prs:
            if pr.title == pr_title:
                logger.error("There is already an open PR with the same title -> %s", pr_title)
                raise RuntimeError(f"another PR with {pr_title=} already exists")
        repo.create_pull(title=pr_title, base='master', head=branch, body=pr_body)
        logger.info("Successfully created a pull request: %s", pr_title)
    except Exception as e:
        logger.error("Failed to create a pull request: %s with error: %s", pr_title, str(e))
        raise e


if __name__ == "__main__":
    repo_root = get_repo_root()
    current_version = get_current_version(repo_root)
    upstream_version = get_upstream_version_from_tags(organization=GVISOR_OWNER, repository=GVISOR_REPOSITORY)

    if upstream_version > current_version:
        logger.info(f"Upstream gVisor version {upstream_version} is newer that current version in repo {current_version}")
        set_current_version(repo_root=repo_root, version=upstream_version)

        git_helper = get_git_helper(
            owner=GARDENER_EXTENSION_GVISOR_OWNER,
            repository=GARDENER_EXTENSION_GVISOR_REPOSITORY,
            github_repo_url=GARDENER_EXTENSION_GVISOR_REPO_URL,
            repo_dir=repo_root
        )

        branch_name = f"bump-gvisor/{upstream_version}"
        commit_message = f"bump gVisor binaries to {upstream_version}"
        resp = commit_and_push_to_branch(git_helper=git_helper, branch=branch_name, commit_msg=commit_message)
        if resp == "BranchExists":
            logger.warning("exiting with rc=0")
            sys.exit(0)

        pr_title = f"Bump gVisor binaries to {upstream_version}"
        pr_body = f"""**What this PR does / why we need it**:

Bumps the version of the included gVisor binaries to {upstream_version}.

**Which issue(s) this PR fixes**:

None

**Release note**:

```other user
The gVisor binaries were updated to release `{upstream_version}`.
```
"""

        create_pull_request(
            git_helper=git_helper,
            org=GARDENER_EXTENSION_GVISOR_OWNER,
            name=GARDENER_EXTENSION_GVISOR_REPOSITORY,
            pr_title=pr_title,
            pr_body=pr_body,
            branch=branch_name
        )

    else:
        logger.info(f"There is no more recent upstream version of gVisor than {current_version}")
