package workflows

type Container struct {
	// The Docker image to use as the container to run the action.
	// The value can be the Docker Hub image name or a registry name.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontainerimage
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `vars`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Image string `json:"image,omitempty" yaml:"image,omitempty" mapstructure:"image,omitempty" validate:"required"`

	// If the image's container registry requires authentication to pull the image,
	// you can use credentials to set a map of the username and password.
	// The credentials are the same values that you would provide to the `docker login` command.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontainercredentials
	Credentials *ContainerCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty" mapstructure:"credentials,omitempty"`

	// Sets an array of environment variables in the container.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontainerenv
	Env Env `json:"env,omitempty" yaml:"env,omitempty" mapstructure:"env,omitempty"`

	// Sets an array of ports to expose on the container.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontainerports
	Ports []string `json:"ports,omitempty" yaml:"ports,omitempty" mapstructure:"ports,omitempty"`

	// Sets an array of volumes for the container to use. You can use volumes to share data between services or other steps in a job.
	// You can specify named Docker volumes, anonymous Docker volumes, or bind mounts on the host.
	// To specify a volume, you specify the source and destination path: <source>:<destinationPath>
	// The <source> is a volume name or an absolute path on the host machine, and <destinationPath> is an absolute path in the container.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontainervolumes
	Volumes []string `json:"volumes,omitempty" yaml:"volumes,omitempty" mapstructure:"volumes,omitempty"`

	// Additional Docker container resource options.
	// For a list of options, see https://docs.docker.com/engine/reference/commandline/create/#options.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontaineroptions
	Options string `json:"options,omitempty" yaml:"options,omitempty" mapstructure:"options,omitempty"`
}

// Context available: `github`, `needs`, `strategy`, `matrix`, `env`, `vars`, `secrets`, `inputs`
// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
type ContainerCredentials struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password,omitempty"`
}
