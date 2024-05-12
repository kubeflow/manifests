# Federated-Learning-on-kubeflow-with-nodered
A simple example of federated learning on kubeflow with node-red which can control number of clients

## Implementation

```
git clone https://github.com/sefgsefg/Federated-Learning-on-kubeflow-with-nodered.git
```

```
cd Federated-Learning-on-kubeflow-with-nodered/FL_kubeflow_with_node-red/example
```

```
./run.sh main
```

Problem solve: -bash: ./run.sh: Permission denied
```
chmod +x run.sh
```

```
cd scripts
```

```
chmod +x entrypoint.sh
```

```
cd ..
```
Run ./run.sh main again
```
./run.sh main
```



1.Build the Federated-Learning flow on node-red.

![](https://github.com/sefgsefg/Federated-Learning-on-kubeflow-with-nodered/blob/main/FL_kubeflow_with_node-red/build_flow.png)

2.Double click the FL node and edit the number of clients.

![](https://github.com/sefgsefg/Federated-Learning-on-kubeflow-with-nodered/blob/main/FL_kubeflow_with_node-red/edit_node.png)

3.Deploy and run the flow. It will run on the kubeflow pipeline.

![](https://github.com/sefgsefg/Federated-Learning-on-kubeflow-with-nodered/blob/main/FL_kubeflow_with_node-red/FL_pipeline.png)

## Reference
https://github.com/justin0322/Node-RED-Kubeflow-Pipeline-Extension

