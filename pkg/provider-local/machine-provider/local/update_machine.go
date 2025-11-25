package local

import (
	"context"
	"fmt"

	apiv1alpha1 "github.com/gardener/gardener/pkg/provider-local/machine-provider/api/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (d *localDriver) UpdateMachine(ctx context.Context, req *driver.UpdateMachineRequest) (*driver.UpdateMachineResponse, error) {
	klog.V(2).Info("UpdateMachine: This is the new code written in gardener hackathon 25")

	if isEmptyUpdateRequest(req) {
		return nil, status.Error(codes.InvalidArgument, "received empty request")
	}

	if req.MachineClass.Provider != apiv1alpha1.Provider {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("requested provider '%s' is not supported by the driver '%s'", req.MachineClass.Provider, apiv1alpha1.Provider))
	}

	klog.V(2).Infof("Machine update request has been received for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine update request has been processed for %q", req.Machine.Name)

	providerSpec, err := validateProviderSpecAndSecret(req.MachineClass, req.Secret)
	if err != nil {
		klog.Error("failed to validate pod")
		return nil, err
	}
	klog.V(2).Info("validateProviderSpecAndSecret Successful!!!")

	klog.V(2).Info("the provider spec looks like", "image", providerSpec.Image)
	klog.V(2).Info("machine name and machine class name", "machine", req.Machine.Name, "machineclass", req.MachineClass.Name)

	podToGet := &corev1.Pod{}
	if err := d.client.Get(ctx, types.NamespacedName{Name: podName(req.Machine.Name), Namespace: getNamespaceForMachine(req.Machine, req.MachineClass)}, podToGet); err != nil {
		klog.Error("failed to get pod", err)
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}
	klog.V(2).Info("d.client.Get Pod Successful!!!")

	patch := client.MergeFrom(podToGet.DeepCopy())

	annotations := podToGet.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["gardener.cloud/update-machine/image"] = providerSpec.Image
	podToGet.SetAnnotations(annotations)

	if err := d.client.Patch(ctx, podToGet, patch); err != nil {
		klog.Error("failed to patch the pod annotation", err)
		return nil, fmt.Errorf("failed annotating pod %s: %w", podToGet.Name, err)
	}

	klog.V(2).Info("Patching Pod is Successful!!!")

	return &driver.UpdateMachineResponse{}, nil
}

func isEmptyUpdateRequest(req *driver.UpdateMachineRequest) bool {
	return req == nil || req.MachineClass == nil || req.Machine == nil || req.Secret == nil
}
