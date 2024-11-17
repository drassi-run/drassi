`drassi`
=======
The drassi `/δράση/` (Greek for `action`) aim to bring GitHub action to everyone & everywhere.

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
`.github/workflows/` files (or for any changes to embedded GitHub actions), you can use `drassi` to run the actions
locally.

### 1.3 Bring GitHub Actions to other git-hosting like `gitea`, `gitlab`,...

If you’re using a self-hosted Git server, such as Gitea or GitLab, but prefer the flexibility, power, and extensive
ecosystem of GitHub Actions, Drassi provides solutions like `gitea-runner` and `gitlab-runner` to integrate seamlessly
with your Git server. Drassi is also a customizable framework, allowing you to tailor or bring your own runner to suit
your specific needs.

## 2. Architecture

![Arch](docs/architecture.svg)
