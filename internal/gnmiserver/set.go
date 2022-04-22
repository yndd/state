/*
Copyright 2021 NDD.

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

package gnmiserver

import (
	"context"
	"strings"
	"time"

	gnmictypes "github.com/karimra/gnmic/types"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/pkg/errors"
	ndddvrv1 "github.com/yndd/ndd-core/apis/dvr/v1"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/ndd-runtime/pkg/utils"
	"github.com/yndd/ndd-yang/pkg/yparser"
	"github.com/yndd/nddp-state/internal/validator"
	"github.com/yndd/nddp-state/pkg/ygotnddpstate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultTimeout = 5 * time.Second
)

func (s *gnmiServerImpl) Set(ctx context.Context, req *gnmi.SetRequest) (*gnmi.SetResponse, error) {

	ok := s.unaryRPCsem.TryAcquire(1)
	if !ok {
		return nil, status.Errorf(codes.ResourceExhausted, errMaxNbrOfUnaryRPCReached)
	}
	defer s.unaryRPCsem.Release(1)

	numUpdates := len(req.GetUpdate())
	numReplaces := len(req.GetReplace())
	numDeletes := len(req.GetDelete())
	if numUpdates+numReplaces+numDeletes == 0 {
		return nil, status.Errorf(codes.InvalidArgument, errMissingPathsInGNMISet)
	}

	log := s.log.WithValues("numUpdates", numUpdates, "numReplaces", numReplaces, "numDeletes", numDeletes)
	prefix := req.GetPrefix()
	target := prefix.GetTarget()

	if numReplaces > 0 {
		log.Debug("Set Replace", "target", prefix.Target, "Path", yparser.GnmiPath2XPath(req.GetReplace()[0].GetPath(), true))

		// check if the target is initialized, if not initialize the target + goStruct as running Config
		if !s.config.HasTarget(target) {
			s.initializeConfig(target)
		}
		ce, err := s.config.Get(target)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, errTargetNotFoundInCache)
		}
		// s.cache.Add(target)

		if err := validator.ValidateUpdate(ce, req.GetReplace(), true, false, validator.Origin_GnmiServer); err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	}

	if numUpdates > 0 {
		log.Debug("Set Update", "target", prefix.Target, "Path", yparser.GnmiPath2XPath(req.GetUpdate()[0].GetPath(), true))
		return nil, status.Errorf(codes.Unimplemented, "not implemented")
	}

	if numDeletes > 0 {
		log.Debug("Set Delete", "target", prefix.Target, "Path", yparser.GnmiPath2XPath(req.GetDelete()[0], true))
		if !s.config.HasTarget(target) {
			return nil, nil
		}
		ce, err := s.config.Get(target)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, errTargetNotFoundInCache)
		}

		if err := validator.ValidateDelete(ce, req.GetDelete(), validator.Origin_GnmiServer); err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	}

	s.handleSet(ctx, target)

	return &gnmi.SetResponse{
		Response: []*gnmi.UpdateResult{
			{
				Timestamp: time.Now().UnixNano(),
			},
		}}, nil
}

func (s *gnmiServerImpl) initializeConfig(target string) error {
	s.config.Add(target)

	// initialize the go struct
	d := &ygotnddpstate.Device{}
	j, err := ygot.EmitJSON(d, &ygot.EmitJSONConfig{})
	if err != nil {
		return err
	}

	ce, err := s.config.Get(target)
	if err != nil {
		return errors.Wrap(err, errTargetNotFoundInCache)
	}
	m := ce.GetModel()

	goStruct, err := m.NewConfigStruct([]byte(j), true)
	if err != nil {
		return err
	}

	ce.SetRunningConfig(goStruct)
	return nil
}

func (s *gnmiServerImpl) handleSet(ctx context.Context, target string) error {
	log := s.log.WithValues("target", target)
	if !s.config.HasTarget(target) {
		return errors.New("target expected")
	}
	ce, err := s.config.Get(target)
	if err != nil {
		return err
	}
	runningConfig, ok := ce.GetRunningConfig().(*ygotnddpstate.Device)
	if !ok {
		return errors.New("unexpected Object")
	}
	// validate if there are still state entries in the config, if not we should delete the target
	if len(runningConfig.StateEntry) == 0 {
		if s.collector.IsActive(target) {
			s.collector.Stop(target)
		}

		if err := s.config.Delete(target); err != nil {
			return err
		}
		// s.cache.Remove(target)

		return nil
	}
	// states exists, we need to restart the collector with the new subscription
	tc, err := s.getTarget(ctx, target)
	if err != nil {
		return err
	}
	if s.collector.IsActive(target) {
		log.Debug("handleSet", "Active", true)
		if err := s.collector.Stop(target); err != nil {
			log.Debug("handleSet", "Stop success", false)
			return err
		}
		s.log.Debug("handleSet", "Stop success", true)

	}
	if err := s.collector.Start(tc, runningConfig); err != nil {
		log.Debug("handleSet", "Start success", false)
		return err
	}
	log.Debug("handleSet", "Start success", true)
	return nil
}

func (s *gnmiServerImpl) getTarget(ctx context.Context, target string) (*gnmictypes.TargetConfig, error) {
	nsn, err := getNamespaceNameFromTarget(target)
	if err != nil {
		return nil, err
	}
	nn := s.newNetworkNode()
	if err := s.client.Get(ctx, types.NamespacedName{
		Namespace: nsn[0],
		Name:      nsn[1],
	}, nn); err != nil {
		return nil, errors.Wrap(resource.IgnoreNotFound(err), errGetNetworkNode)
	}

	// Retrieve the Login details from the network node spec and validate
	// the network node details and build the credentials for communicating
	// to the network node.
	creds, err := s.validateCredentials(ctx, nn)
	if err != nil || creds == nil {
		return nil, errors.Wrap(err, "credentials error")
	}
	tc := getTargetConfig(target, nn, creds)
	return tc, nil

}

type Credentials struct {
	Username string
	Password string
}

func (s *gnmiServerImpl) validateCredentials(ctx context.Context, nn ndddvrv1.Nn) (creds *Credentials, err error) {
	//log := r.log.WithValues("namespace", nn.GetNamespace(), "credentialsName", nn.GetTargetCredentialsName(), "targetAddress", nn.GetTargetAddress())
	//log.Debug("Credentials Validation")
	// Retrieve the secret from Kubernetes for this network node

	credsSecret, err := s.getSecret(ctx, nn)
	if err != nil {
		return nil, err
	}

	// Check if address is defined on the network node
	if nn.GetTargetAddress() == "" {
		return nil, errors.New(errEmptyTargetAddress)
	}

	creds = &Credentials{
		Username: strings.TrimSuffix(string(credsSecret.Data["username"]), "\n"),
		Password: strings.TrimSuffix(string(credsSecret.Data["password"]), "\n"),
	}

	//log.Debug("Credentials", "creds", creds)

	if creds.Username == "" {
		return nil, errors.New(errMissingUsername)
	}
	if creds.Password == "" {
		return nil, errors.New(errMissingPassword)
	}

	return creds, nil
}

// Retrieve the secret containing the credentials for talking to the Network Node.
func (r *gnmiServerImpl) getSecret(ctx context.Context, nn ndddvrv1.Nn) (credsSecret *corev1.Secret, err error) {
	// if namespace
	// check if credentialName is specified
	if nn.GetTargetCredentialsName() == "" {
		return nil, errors.New(errEmptyTargetSecretReference)
	}
	namespace := nn.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	// check if credential secret exists
	secretKey := types.NamespacedName{
		Name:      nn.GetTargetCredentialsName(),
		Namespace: namespace,
	}
	credsSecret = &corev1.Secret{}
	if err := r.client.Get(ctx, secretKey, credsSecret); resource.IgnoreNotFound(err) != nil {
		return nil, errors.Wrap(err, errCredentialSecretDoesNotExist)
	}
	return credsSecret, nil
}

func getNamespaceNameFromTarget(target string) ([]string, error) {
	split := strings.Split(target, ".")
	if len(split) != 2 {
		return nil, errors.New("wrong target input")
	}
	return split, nil
}

func getTargetConfig(target string, nn ndddvrv1.Nn, creds *Credentials) *gnmictypes.TargetConfig {
	return &gnmictypes.TargetConfig{
		Name:       target,
		Address:    nn.GetTargetAddress(),
		Username:   &creds.Username,
		Password:   &creds.Password,
		Timeout:    defaultTimeout,
		Insecure:   utils.BoolPtr(nn.GetTargetInsecure()),
		SkipVerify: utils.BoolPtr(nn.GetTargetSkipVerify()),
		TLSCA:      utils.StringPtr(""), //TODO TLS
		TLSCert:    utils.StringPtr(""), //TODO TLS
		TLSKey:     utils.StringPtr(""), //TODO TLS
		Gzip:       utils.BoolPtr(false),
	}
}
