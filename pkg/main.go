package pkg

import (
	elasticsearchkubernetesv1 "buf.build/gen/go/plantoncloud/project-planton/protocolbuffers/go/project/planton/apis/provider/kubernetes/elasticsearchkubernetes/v1"
	"github.com/pkg/errors"
	"github.com/plantoncloud/elasticsearch-kubernetes-pulumi-module/pkg/outputs"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *elasticsearchkubernetesv1.ElasticsearchKubernetesStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesClusterCredential(ctx,
		stackInput.KubernetesCluster, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	createdNamespace, err := kubernetescorev1.NewNamespace(ctx, locals.ElasticsearchKubernetes.Metadata.Id,
		&kubernetescorev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaPtrInput(
				&metav1.ObjectMetaArgs{
					Name:   pulumi.String(locals.ElasticsearchKubernetes.Metadata.Id),
					Labels: pulumi.ToStringMap(locals.Labels),
				}),
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create namespace")
	}

	//export name of the namespace
	ctx.Export(outputs.Namespace, createdNamespace.Metadata.Name())

	if err := elasticsearch(ctx, locals, createdNamespace); err != nil {
		return errors.Wrap(err, "failed to create elastic search resources")
	}

	if locals.ElasticsearchKubernetes.Spec.Ingress.IsEnabled {
		if err := ingress(ctx, locals, createdNamespace, kubernetesProvider); err != nil {
			return errors.Wrap(err, "failed to create ingress resources")
		}
	}

	return nil
}
