package typed

import "k8s.io/client-go/kubernetes/typed/policy/v1beta1"

// Nodes represents mock func for similar runtime client func
func (pe *PolicyV1Beta1) Evictions() v1beta1.EvictionInterface {
	args := pe.Called()
	v := args.Get(0)
	return v.(v1beta1.EvictionInterface)
}
