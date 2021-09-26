package utils

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"

	"github.com/soer3n/yaho/pkg/client"
	"sigs.k8s.io/descheduler/pkg/descheduler/evictions"
)

const rescheduleStrategy = ""

// RescheduleQuarantineController represents descheduling of quarantine controller if needed due to validation
func RescheduleQuarantineController(excludedNodes []string) error {

	var err error
	var success bool
	var excludedNodesObj []*corev1.Node

	pod := &corev1.Pod{}
	node := &corev1.Node{}
	ev := evictions.NewPodEvictor(client.New().TypedClient, rescheduleStrategy, false, 1, excludedNodesObj, false)

	if success, err = ev.EvictPod(context.TODO(), pod, node); err != nil {
		return err
	}

	if !success {
		return errors.New("no success on rescheduling quarantine controller")
	}

	return nil
}
