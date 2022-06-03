//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022 NDD.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by ndd-gen. DO NOT EDIT.

package v1alpha1

import nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"

// GetActive of this State.
func (x *State) GetDeploymentPolicy() nddv1.DeploymentPolicy {
	return x.Spec.Lifecycle.DeploymentPolicy
}

// GetCondition of this State.
func (x *State) GetCondition(ck nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ck)
}

// GetDeletionPolicy of this State.
func (x *State) GetDeletionPolicy() nddv1.DeletionPolicy {
	return x.Spec.Lifecycle.DeletionPolicy
}

// GetTargetReference of this State.
func (x *State) GetTargetReference() *nddv1.Reference {
	return x.Spec.TargetReference
}

// SetRootPaths of this State.
func (x *State) GetRootPaths() []string {
	return x.Status.RootPaths
}

// SetActive of this State.
func (x *State) SetDeploymentPolicy(b nddv1.DeploymentPolicy) {
	x.Spec.Lifecycle.DeploymentPolicy = b
}

// SetConditions of this State.
func (x *State) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

// SetDeletionPolicy of this State.
func (x *State) SetDeletionPolicy(r nddv1.DeletionPolicy) {
	x.Spec.Lifecycle.DeletionPolicy = r
}

// SetTargetReference of this State.
func (x *State) SetTargetReference(r *nddv1.Reference) {
	x.Spec.TargetReference = r
}

// SetRootPaths of this State.
func (x *State) SetRootPaths(n []string) {
	x.Status.RootPaths = n
}

func (x *State) SetHealthConditions(c nddv1.HealthConditionedStatus) {
	x.Status.Health = c
}
