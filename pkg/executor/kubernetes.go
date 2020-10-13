package executor

import (
	"context"
	"fmt"
	"strings"

	"arhat.dev/pkg/envhelper"
	"arhat.dev/pkg/hashhelper"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientbatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"arhat.dev/renovate-server/pkg/conf"
	"arhat.dev/renovate-server/pkg/constant"
	"arhat.dev/renovate-server/pkg/types"
)

func NewKubernetesExecutor(config *conf.KubernetesExecutorConfig) (types.Executor, error) {
	client, _, err := config.KubeClient.NewKubeClient(nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	var pullPolicy corev1.PullPolicy
	switch strings.ToLower(config.RenovateImagePullPolicy) {
	case "always":
	case "never", "":
		pullPolicy = corev1.PullNever
	case "ifnotpresent", "if_not_present":
		pullPolicy = corev1.PullIfNotPresent
	default:
		return nil, fmt.Errorf("unsupported image pull policy: %s", config.RenovateImagePullPolicy)
	}
	image := config.RenovateImage
	if image == "" {
		image = constant.DefaultRenovateImage
	}

	return &KubernetesExecutor{
		image:           image,
		imagePullPolicy: pullPolicy,

		secretClient: client.CoreV1().Secrets(envhelper.ThisPodNS()),
		jobClient:    client.BatchV1().Jobs(envhelper.ThisPodNS()),
	}, nil
}

type KubernetesExecutor struct {
	ctx context.Context

	image           string
	imagePullPolicy corev1.PullPolicy

	secretClient clientcorev1.SecretInterface
	jobClient    clientbatchv1.JobInterface
}

func (k *KubernetesExecutor) Execute(args types.ExecutionArgs) error {
	trueP := true
	falseP := false
	zeroP := int64(0)
	oneP := int32(1)

	apiTokenBytes := []byte(args.APIToken)
	secretName := fmt.Sprintf("renovate-%s", hashhelper.MD5SumHex(apiTokenBytes))

	_, err := k.secretClient.Get(k.ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !kubeerrors.IsNotFound(err) {
			return fmt.Errorf("failed to check required secret: %w", err)
		}

		_, err = k.secretClient.Create(k.ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: envhelper.ThisPodNS(),
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"RENOVATE_TOKEN": apiTokenBytes,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create required secret: %w", err)
		}
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "renovate-",
			Namespace:    envhelper.ThisPodNS(),
		},
		Spec: batchv1.JobSpec{
			Parallelism:             &oneP,
			Completions:             &oneP,
			ActiveDeadlineSeconds:   nil,
			BackoffLimit:            &oneP,
			TTLSecondsAfterFinished: nil,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						constant.LabelRenovateRepo: strings.ToLower(strings.ReplaceAll(args.Repo, "/", "-")),
					},
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: nil,
					Containers: []corev1.Container{{
						Name:            "renovate",
						TTY:             true,
						Image:           k.image,
						ImagePullPolicy: k.imagePullPolicy,
						Command:         []string{},
						Args:            []string{args.Repo},
						EnvFrom: []corev1.EnvFromSource{{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secretName,
								},
							},
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "LOG_LEVEL",
								Value: "debug",
							},
							{
								Name:  "LOG_FORMAT",
								Value: "json",
							},
							{
								Name:  "RENOVATE_PLATFORM",
								Value: strings.ToLower(args.Platform),
							},
							{
								Name:  "RENOVATE_GIT_AUTHOR",
								Value: fmt.Sprintf("%s <%s>", args.GitUser, args.GitEmail),
							},
							{
								Name:  "RENOVATE_ONBOARDING",
								Value: "false",
							},
							{
								Name:  "RENOVATE_ONBOARDING",
								Value: "false",
							},
							{
								Name:  "RENOVATE_TRUST_LEVEL",
								Value: "low",
							},
							{
								Name:  "RENOVATE_BASE_DIR",
								Value: "/tmp/renovate",
							},
							{
								Name:  "RENOVATE_AUTODISCOVER",
								Value: "false",
							},
							{
								Name:  "RENOVATE_ENDPOINT",
								Value: args.APIURL,
							},
							{
								Name:  "RENOVATE_BINARY_SOURCE",
								Value: "global",
							},
						},
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add:  nil,
								Drop: []corev1.Capability{"all"},
							},
							Privileged:               nil,
							SELinuxOptions:           nil,
							WindowsOptions:           nil,
							RunAsUser:                nil,
							RunAsGroup:               nil,
							RunAsNonRoot:             &trueP,
							ReadOnlyRootFilesystem:   &falseP,
							AllowPrivilegeEscalation: &falseP,
							ProcMount:                nil,
						},
					}},
					RestartPolicy:                 corev1.RestartPolicyNever,
					TerminationGracePeriodSeconds: &zeroP,
					ActiveDeadlineSeconds:         nil,
					DNSPolicy:                     corev1.DNSClusterFirst,
					SecurityContext: &corev1.PodSecurityContext{
						SELinuxOptions:      nil,
						WindowsOptions:      nil,
						RunAsUser:           nil,
						RunAsGroup:          nil,
						RunAsNonRoot:        &trueP,
						SupplementalGroups:  nil,
						FSGroup:             nil,
						Sysctls:             nil,
						FSGroupChangePolicy: nil,
					},
				},
			},
		},
	}

	_, err = k.jobClient.Create(k.ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create kubernetes job: %w", err)
	}

	return nil
}
