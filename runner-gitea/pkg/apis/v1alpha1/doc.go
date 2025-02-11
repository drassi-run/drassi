// +groupName=runner.drassi.run
// +k8s:deepcopy-gen=package

//go:generate deepcopy-gen --output-file zz_generated.deepcopy.go drassi.run/gitea-runner/pkg/apis/v1alpha1
//go:generate register-gen --output-file zz_generated.register.go drassi.run/gitea-runner/pkg/apis/v1alpha1
package v1alpha1
