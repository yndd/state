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

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	//"strings"
	"time"

	"github.com/karimra/gnmic/target"
	gnmitypes "github.com/karimra/gnmic/types"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/reconciler/managedgeneric"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/ndd-runtime/pkg/utils"
	targetv1 "github.com/yndd/ndd-target-runtime/apis/dvr/v1"
	"github.com/yndd/ndd-yang/pkg/yparser"
	statev1alpha1 "github.com/yndd/nddp-state/apis/state/v1alpha1"
	"github.com/yndd/nddp-state/internal/model"
	"github.com/yndd/nddp-state/internal/shared"
	"github.com/yndd/nddp-state/pkg/ygotnddpstate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// SetupDevice adds a controller that reconciles Devices.
func Setup(mgr ctrl.Manager, o controller.Options, nddcopts *shared.NddControllerOptions) error {
	//func SetupDevice(mgr ctrl.Manager, o controller.Options, nddcopts *shared.NddControllerOptions) error {

	name := managedgeneric.ControllerName(statev1alpha1.StateGroupKind)

	fm := &model.Model{
		StructRootType:  reflect.TypeOf((*ygotnddpstate.Device)(nil)),
		SchemaTreeRoot:  ygotnddpstate.SchemaTree["Device"],
		JsonUnmarshaler: ygotnddpstate.Unmarshal,
		EnumData:        ygotnddpstate.ΛEnum,
	}

	m := &model.Model{
		StructRootType:  reflect.TypeOf((*ygotnddpstate.NddpState_StateEntry)(nil)),
		SchemaTreeRoot:  ygotnddpstate.SchemaTree["NddpState_StateEntry"],
		JsonUnmarshaler: ygotnddpstate.Unmarshal,
		EnumData:        ygotnddpstate.ΛEnum,
	}

	r := managedgeneric.NewReconciler(mgr,
		resource.ManagedKind(statev1alpha1.StateGroupVersionKind),
		managedgeneric.WithPollInterval(nddcopts.Poll),
		managedgeneric.WithExternalConnecter(&connectorDevice{
			log:         nddcopts.Logger,
			kube:        mgr.GetClient(),
			usage:       resource.NewTargetUsageTracker(mgr.GetClient(), &targetv1.TargetUsage{}),
			m:           m,
			fm:          fm,
			newClientFn: target.NewTarget,
			gnmiAddress: nddcopts.GnmiAddress},
		),
		managedgeneric.WithLogger(nddcopts.Logger.WithValues("State", name)),
		managedgeneric.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	StateHandler := &EnqueueRequestForAllState{
		client: mgr.GetClient(),
		log:    nddcopts.Logger,
		ctx:    context.Background(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&statev1alpha1.State{}).
		Owns(&statev1alpha1.State{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Watches(&source.Kind{Type: &statev1alpha1.State{}}, StateHandler).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connectorDevice struct {
	log         logging.Logger
	kube        client.Client
	usage       resource.Tracker
	m           *model.Model
	fm          *model.Model
	newClientFn func(c *gnmitypes.TargetConfig) *target.Target
	gnmiAddress string
}

// Connect produces an ExternalClient by:
// 1. Tracking that the managed resource is using a NetworkNode.
// 2. Getting the managed resource's NetworkNode with connection details
// A resource is mapped to a single target
func (c *connectorDevice) Connect(ctx context.Context, mg resource.Managed) (managedgeneric.ExternalClient, error) {
	log := c.log.WithValues("resource", mg.GetName())
	//log.Debug("Connect")

	cr, ok := mg.(*statev1alpha1.State)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackTCUsage)
	}

	// find network node that is configured status
	nn := &targetv1.Target{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetTargetReference().Name}, nn); err != nil {
		return nil, errors.Wrap(err, errGetNetworkNode)
	}

	// if nn.GetCondition(ndrv1.ConditionKindDeviceDriverConfigured).Status != corev1.ConditionTrue {
	// 	return nil, errors.New(targetNotConfigured)
	// }

	cfg := &gnmitypes.TargetConfig{
		Name:       cr.GetTargetReference().Name,
		Address:    c.gnmiAddress,
		Username:   utils.StringPtr("admin"),
		Password:   utils.StringPtr("admin"),
		Timeout:    10 * time.Second,
		SkipVerify: utils.BoolPtr(true),
		Insecure:   utils.BoolPtr(true),
		TLSCA:      utils.StringPtr(""), //TODO TLS
		TLSCert:    utils.StringPtr(""), //TODO TLS
		TLSKey:     utils.StringPtr(""),
		Gzip:       utils.BoolPtr(false),
	}

	cl := target.NewTarget(cfg)
	if err := cl.CreateGNMIClient(ctx); err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	tns := []string{nn.GetName()}

	//return &externalDevice{client: cl, targets: tns, log: log, deviceSchema: c.deviceSchema, nddpSchema: c.nddpSchema, deviceModel: c.deviceModel, systemModel: c.systemModel}, nil
	return &externalDevice{client: cl, targets: tns, log: log, m: c.m, fm: c.fm}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type externalDevice struct {
	client  *target.Target
	targets []string
	log     logging.Logger
	m       *model.Model
	fm      *model.Model
}

func (e *externalDevice) Close() {
	e.client.Close()
}

func (e *externalDevice) Observe(ctx context.Context, mg resource.Managed) (managedgeneric.ExternalObservation, error) {
	log := e.log.WithValues("Resource", mg.GetName())
	log.Debug("Observing ...")

	stateEntry, err := e.getSpec(mg)
	if err != nil {
		return managedgeneric.ExternalObservation{}, err
	}

	crTarget := strings.Join([]string{mg.GetNamespace(), mg.GetTargetReference().Name}, ".")

	req := &gnmi.GetRequest{
		Prefix:   &gnmi.Path{Target: crTarget},
		Path:     []*gnmi.Path{{}},
		Encoding: gnmi.Encoding_JSON,
	}

	// gnmi get response
	resp, err := e.client.Get(ctx, req)
	if err != nil {
		log.Debug("Observing ...", "error", err)
		if er, ok := status.FromError(err); ok {
			switch er.Code() {
			case codes.Unavailable:
				// we use this to signal not ready
				return managedgeneric.ExternalObservation{}, nil
			case codes.NotFound:
				return managedgeneric.ExternalObservation{}, nil
			}
		}
	}

	var cacheState interface{}
	if len(resp.GetNotification()) == 0 {
		return managedgeneric.ExternalObservation{}, nil
	}
	if len(resp.GetNotification()) != 0 && len(resp.GetNotification()[0].GetUpdate()) != 0 {
		// get value from gnmi get response
		cacheState, err = yparser.GetValue(resp.GetNotification()[0].GetUpdate()[0].Val)
		if err != nil {
			return managedgeneric.ExternalObservation{}, errors.Wrap(err, errJSONMarshal)
		}

		switch cacheState.(type) {
		case nil:
			// resource has no data
			return managedgeneric.ExternalObservation{}, nil
		}
	}

	cacheStateData, err := json.Marshal(cacheState)
	if err != nil {
		return managedgeneric.ExternalObservation{}, err
	}
	log.Debug("Observing ...", "cacheStateData", string(cacheStateData))

	// validate the state cache as a validtedGoStruct
	validatedGoStruct, err := e.fm.NewConfigStruct(cacheStateData, true)
	if err != nil {
		return managedgeneric.ExternalObservation{}, err
	}
	// type casting
	cacheNddpStateDevice, ok := validatedGoStruct.(*ygotnddpstate.Device)
	if !ok {
		return managedgeneric.ExternalObservation{}, errors.New("wrong nddp state object")
	}

	log.Debug("Observing ...", "cacheNddpStateDevice", cacheNddpStateDevice)

	// check if the entry exists
	cacheStateEntry, ok := cacheNddpStateDevice.StateEntry[*stateEntry.Name]
	if !ok {
		return managedgeneric.ExternalObservation{}, nil
	}

	log.Debug("Observing ...", "cacheStateEntry", cacheStateEntry)

	// check if the cacheData is aligned with the crSpecData
	deletes, updates, err := e.diff(mg, cacheStateEntry)
	if err != nil {
		return managedgeneric.ExternalObservation{}, err
	}

	return managedgeneric.ExternalObservation{
		Exists:     true,
		IsUpToDate: len(deletes) == 0 && len(updates) == 0,
	}, nil
}

func (e *externalDevice) Create(ctx context.Context, mg resource.Managed) error {
	log := e.log.WithValues("Resource", mg.GetName())
	log.Debug("Creating ...")

	updates, err := e.getUpate(mg)
	if err != nil {
		return errors.Wrap(err, errCreateResource)
	}

	crTarget := strings.Join([]string{mg.GetNamespace(), mg.GetTargetReference().Name}, ".")

	req := &gnmi.SetRequest{
		Prefix:  &gnmi.Path{Target: crTarget},
		Replace: updates,
	}

	_, err = e.client.Set(ctx, req)
	if err != nil {
		return errors.Wrap(err, errCreateResource)
	}

	return nil
}

func (e *externalDevice) Delete(ctx context.Context, mg resource.Managed) error {
	log := e.log.WithValues("Resource", mg.GetName())
	log.Debug("Deleting ...")

	paths, err := e.getPath(mg)
	if err != nil {
		return errors.Wrap(err, errDeleteResource)
	}

	crTarget := strings.Join([]string{mg.GetNamespace(), mg.GetTargetReference().Name}, ".")

	req := &gnmi.SetRequest{
		Prefix: &gnmi.Path{Target: crTarget},
		Delete: paths,
	}

	_, err = e.client.Set(ctx, req)
	if err != nil {
		return errors.Wrap(err, errDeleteResource)
	}

	return nil
}

// getUpate returns an update to the cache
func (e *externalDevice) getUpate(mg resource.Managed) ([]*gnmi.Update, error) {
	e.log.Debug("getUpate")

	stateEntry, err := e.getSpec(mg)
	if err != nil {
		return nil, err
	}

	stateEntryJson, err := ygot.EmitJSON(stateEntry, &ygot.EmitJSONConfig{
		Format: ygot.RFC7951,
	})
	if err != nil {
		return nil, err
	}

	//return update
	return []*gnmi.Update{
		{
			Path: &gnmi.Path{
				Elem: []*gnmi.PathElem{
					{Name: "state-entry", Key: map[string]string{"name": *stateEntry.Name}},
				},
			},
			Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte(stateEntryJson)}},
		},
	}, nil
}

// getUpate returns an update to the cache
func (e *externalDevice) getPath(mg resource.Managed) ([]*gnmi.Path, error) {
	e.log.Debug("getUpate")

	stateEntry, err := e.getSpec(mg)
	if err != nil {
		return nil, err
	}

	//return path
	return []*gnmi.Path{
		{
			Elem: []*gnmi.PathElem{
				{Name: "state-entry", Key: map[string]string{"name": *stateEntry.Name}},
			},
		},
	}, nil
}

func (e *externalDevice) getValidatedGoStructFromCr(mg resource.Managed) (ygot.ValidatedGoStruct, error) {
	cr, ok := mg.(*statev1alpha1.State)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	return e.m.NewConfigStruct(cr.Spec.Properties.Raw, true)
}

// getSpec return the spec as a stateEntry
func (e *externalDevice) getSpec(mg resource.Managed) (*ygotnddpstate.NddpState_StateEntry, error) {
	validatedGoStruct, err := e.getValidatedGoStructFromCr(mg)
	if err != nil {
		return nil, err
	}
	stateEntry, ok := validatedGoStruct.(*ygotnddpstate.NddpState_StateEntry)
	if !ok {
		return nil, errors.New("wrong object nddp state entry")
	}

	return stateEntry, nil
}

func (e *externalDevice) diff(mg resource.Managed, cacheStateEntry interface{}) ([]*gnmi.Path, []*gnmi.Update, error) {
	// check if the cacheData is aligned with the crSpecData
	specConfig, err := e.getValidatedGoStructFromCr(mg)
	if err != nil {
		return nil, nil, err
	}
	e.log.Debug("observe diff", "specConfig", specConfig)

	cacheConfig, ok := cacheStateEntry.(ygot.ValidatedGoStruct)
	if !ok {
		return nil, nil, errors.New("invalid Object")
	}
	e.log.Debug("observe diff", "cacheConfig", cacheConfig)

	// create a diff of the actual compared to the to-become-new config
	actualVsSpecDiff, err := ygot.Diff(specConfig, cacheConfig, &ygot.DiffPathOpt{MapToSinglePath: true})
	if err != nil {
		return nil, nil, err
	}

	deletes, updates := validateNotification(actualVsSpecDiff)
	return deletes, updates, nil
}

func validateNotification(n *gnmi.Notification) ([]*gnmi.Path, []*gnmi.Update) {
	updates := make([]*gnmi.Update, 0)
	for _, u := range n.GetUpdate() {
		fmt.Printf("validateNotification diff update old path: %s, value: %v\n", yparser.GnmiPath2XPath(u.GetPath(), true), u.GetVal())
		// workaround since the diff can return double pathElem
		var changed bool
		changed, u.Path = validatePath(u.GetPath())
		if changed {
			u.Val = &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte("{}")}}
		}
		fmt.Printf("validateNotification diff update new path: %s, value: %v\n", yparser.GnmiPath2XPath(u.GetPath(), true), u.GetVal())
		updates = append(updates, u)
	}

	deletes := make([]*gnmi.Path, 0)
	for _, p := range n.GetDelete() {
		fmt.Printf("validateNotification diff delete old path: %s\n", yparser.GnmiPath2XPath(p, true))
		// workaround since the diff can return double pathElem
		_, p = validatePath(p)
		fmt.Printf("validateNotification diff delete new path: %s\n", yparser.GnmiPath2XPath(p, true))
		deletes = append(deletes, p)
	}
	return deletes, updates
}

// workaround for the diff handling
func validatePath(p *gnmi.Path) (bool, *gnmi.Path) {
	if len(p.GetElem()) <= 1 {
		return false, p
	}
	// when the 2nd last pathElem has a key and the last PathElem is an entry in the Key we should trim the last entry from the path
	// e.g. /interface[name=ethernet-1/49]/subinterface[index=1]/ipv4/address[ip-prefix=100.64.0.0/31]/ip-prefix, value: string_val:"100.64.0.0/31"
	// e.g. /interface[name=ethernet-1/49]/subinterface[index=1]/ipv4/address[ip-prefix=100.64.0.0/31]/ip-prefix, value: string_val:"100.64.0.0/31"
	if len(p.GetElem()[len(p.GetElem())-2].GetKey()) > 0 {
		if _, ok := p.GetElem()[len(p.GetElem())-2].GetKey()[p.GetElem()[len(p.GetElem())-1].GetName()]; ok {
			p.Elem = p.Elem[:len(p.GetElem())-1]
			return true, p
		}
	}
	return false, p
}
