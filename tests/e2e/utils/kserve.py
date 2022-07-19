from kfp import components


def create_serving_task(model_name, model_namespace, tfjob_op, model_volume_op):

    api_version = 'serving.kserve.io/v1beta1'
    serving_component_url = 'https://raw.githubusercontent.com/kubeflow/pipelines/master/components/kserve/component.yaml'

    # Uncomment the following two lines if you are using KFServing v0.6.x or v0.5.x
    # api_version = 'serving.kubeflow.org/v1beta1'
    # serving_component_url = 'https://raw.githubusercontent.com/kubeflow/pipelines/master/components/kubeflow/kfserving/component.yaml'

    inference_service = '''
apiVersion: "{}"
kind: "InferenceService"
metadata:
  name: {}
  namespace: {}
  annotations:
    "sidecar.istio.io/inject": "false"
spec:
  predictor:
    tensorflow:
      storageUri: "pvc://{}/"
'''.format(api_version, model_name, model_namespace, str(model_volume_op.outputs["name"]))

    serving_launcher_op = components.load_component_from_url(serving_component_url)
    serving_launcher_op(action="apply", inferenceservice_yaml=inference_service).after(tfjob_op)