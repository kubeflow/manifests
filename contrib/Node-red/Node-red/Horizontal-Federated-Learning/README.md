# Federated-Learning-on-kubeflow-with-nodered
A simple example of federated learning on kubeflow with node-red which can control number of clients

## Table of Contents
<!-- toc -->
- [Federated Learning Overview](#Federated)
- [Self-defined Node](#Self-defined-Node)
  * [Prerequisites](#Prerequisites)
  * [snippet.js](#snippet.js)
  * [example.js](#example.js)
  * [example.html](#example.html)
- [Reference](#Reference)

<!-- tocstop -->

# Federated Learning Overview
Federated Learning is a decentralized machine learning technique where model training happens locally on devices holding data, preserving privacy. Instead of sending data to a central server, model updates are sent, allowing collaborative learning without exposing sensitive information. This approach is efficient for edge devices, reduces data transfer, and is ideal for privacy-sensitive applications like healthcare and finance, revolutionizing how machine learning is done in connected environments.

## Installation
.docker

.Wsl(for windows), make sure it has connected to docker.

## Implementation
Open terminal(wsl) and type.
```
git clone https://github.com/sefgsefg/Federated-Learning-on-kubeflow-with-nodered.git
```


open the file FL_kubeflow_with_node-red/example/main/node_modules/snippets.js and change some data
1. Training data change, you can use example data we provide.

![](https://github.com/sefgsefg/manifests/blob/master/contrib/Node-red/Node-red/Horizontal-Federated-Learning/FL_kubeflow_with_node-red/data_select.png)


2. Type your Kubeflow's url and account.

![](https://github.com/sefgsefg/manifests/blob/master/contrib/Node-red/Node-red/Horizontal-Federated-Learning/FL_kubeflow_with_node-red/account.png)


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

