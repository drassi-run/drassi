// +groupName=sandboxer.drassi.run
// +k8s:deepcopy-gen=package

//go:generate deepcopy-gen --output-file zz_generated.deepcopy.go drassi.run/core/pkg/sandboxer/apis/v1alpha1
//go:generate register-gen --output-file zz_generated.register.go drassi.run/core/pkg/sandboxer/apis/v1alpha1
package v1alpha1
