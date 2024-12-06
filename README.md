`drassi`
=======
The drassi `/δράση/` (Greek for `action`) aim to bring [GitHub Actions](https://github.com/features/actions) to
everyone and everywhere.

## 1. Use-cases

### 1.1 Alternative to [`actions/runner`](https://github.com/actions/runner) for self-hosted runners

`actions/runner` executes all jobs within the host environment, making it unsuitable for projects requiring strong
isolation and security. Additionally, `actions/runner` can only process one job at a time, which means implementing
autoscaling with self-hosted runners requires additional tools like `actions-runner-controller` (ARC).

Drassi's `gha-runner` offers a drop-in replacement for `actions/runner`, with enhanced security and isolation. It
executes each job within a sandboxed environment, providing flexibility by supporting various sandboxing implementations
such as Docker containers, Incus, or microVMs. These isolated environments enables multiple jobs can be able to run in
parallel, optimizing both resource utilization and time efficiency.

With Drassi's `gha-runner`, you gain improved job isolation, security, and parallel job execution, significantly
enhancing your CI/CD pipeline's performance and scalability.

### 1.2 Run your GitHub Actions locally

Rather than having to commit/push every time you want to test out the changes you are making to your
`.github/workflows/` files (or for any changes to embedded GitHub actions), you can use `drassi` CLI to run the actions
locally.

### 1.3 Bring GitHub Actions to other git-hosting like `gitea`, `gitlab`,...

If you’re using a self-hosted Git server, such as Gitea or GitLab, but prefer the flexibility, power, and extensive
ecosystem of GitHub Actions, Drassi provides solutions like `gitea-runner` and `gitlab-runner` to integrate seamlessly
with your Git server. Drassi is also a customizable framework, allowing you to tailor or bring your own runner to suit
your specific needs.

## 2. Architecture

![Arch](docs/architecture.svg)

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
