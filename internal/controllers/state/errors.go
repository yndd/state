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

package state

const (
	errTrackTCUsage     = "cannot track TargetConfig usage"
	errGetTarget        = "cannot get Target"
	targetNotConfigured = "target is not configured to proceed"
	errNewClient        = "cannot create new client"
	errJSONMarshal      = "cannot marshal JSON object"
	errUnexpectedObject = "the managed resource is not a state managed resource"
	errObserveResource  = "cannot observe State"
	errCreateResource   = "cannot create State"
	errDeleteResource   = "cannot delete State"
)
