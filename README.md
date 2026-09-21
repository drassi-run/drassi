`drassi`
=======
Drassi `/δράση/` (Greek for `action`) aims to bring [GitHub Actions](https://github.com/features/actions) to
everyone and everywhere.

## Target users

### 1. GitHub users who want faster jobs

GitHub-hosted Actions runners
are [slow](https://medium.com/@akhilesh-mishra/your-github-actions-runners-are-slow-and-you-are-paying-too-much-for-them-5406577314fe).
Self-hosted runners can be faster, but each runner is limited to **one job at a time**.

### 2. Self-hosted runner users who need dynamic scaling

To run multiple jobs in parallel, users often need solutions
like [Actions Runner Controller (ARC)](https://github.com/actions/actions-runner-controller).
Many jobs also need Docker access, for example:

* [building images in CI jobs](https://github.com/marketplace/actions/build-and-push-docker-images)
* [using Docker container actions](https://docs.github.com/en/actions/tutorials/use-containerized-services/create-a-docker-container-action)
* [running jobs inside containers](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#jobsjob_idcontainer)

In these use cases, you usually need Docker-in-Docker (`--privileged`) or a mounted Docker socket.
Both options raise **security concerns** when a runner is shared across multiple teams or tenants.

### 3. GitLab, Gitea, and other users who want GitHub Actions

Many teams want the GitHub Actions model because it provides:

* composable steps
* multiple workflows
* reusable jobs
* a large actions ecosystem

However, `actions/runner` is designed to work with GitHub servers. `act`-based runners can support GitHub-like servers,
such as Gitea, but they are not fully compatible with the GitHub runner and miss features like problem matchers and
issue reporting.

### 4. Users who want to run jobs locally

Running jobs locally reduces the feedback loop before pushing to a Git server or debugging CI failures.

## Architecture

![Arch](docs/architecture.svg)

Drassi supports multiple runners, including GitHub, Gitea, and GitLab, from a shared core.
Each job runs in a sandboxed environment, with multiple sandboxers available to match different isolation, security, and
performance needs.

## 3. Usage

### 3.1 `drassi` CLI

// TODO

### 3.2 `gha-runner`

// TODO

### 3.3 `gitea-runner`

// TODO

## 3. Sandboxer

Drassi offers multiple sandboxers, each with different characteristics in terms of speed, security and isolation.

### 3.1 Host sandboxer

The host sandboxer is the most basic sandboxing environment that executes each job in a temporary directory on the host
system. As a result, the host sandboxer lacks isolation and security, making it unsuitable for production use. It is
primarily used for local CI testing where simplicity and speed are prioritized over security.

### 3.2 Container sandboxer

Container sandboxer utilizes Docker containers to isolate job executions. By running each job in its own container, it
offers a more secure and controlled environment. However, the default behavior of mounting the Docker socket from the
host to the container can introduce security risks. This makes container sandboxes more suitable for local development
and testing rather than production deployments.

### 3.3 microVM sandboxer

MicroVM sandboxer executes jobs within lightweight virtual machines (VMs). This approach offers a high level of security
and isolation, but it requires more complex setup and can be slower to start compared to container sandboxes.

### 3.4 incus sandboxer

Incus sandboxer is an alternative solution for environments that don't support nested virtualization, such as AWS EC2.
It provides a secure and isolated execution environment without the overhead of full virtualization.

### Summary

|                   | Isolation | Security | Performance | Recommendation                             |
|-------------------|:---------:|:--------:|:-----------:|--------------------------------------------|
| Host sandboxer    |    2/5    |   2/5    |     5/5     | local use                                  |
| Container sandbox |    4/5    |   4/5    |     5/5     | local use                                  |
| microVM sandbox   |    5/5    |   5/5    |     4/5     | production use                             |
| incus sandbox     |    5/5    |  4.5/5   |     4/5     | production use, where microVM's NOT worked |
