# Metadata Kustomization for Azure

This directory contains configurations and guidelines on setting up metadata service to connect to an [Azure MySQL](https://docs.microsoft.com/en-us/azure/mysql/) database.

#### 1. Create an Azure MySQL database
Create an Azure MySQL data base following the [guidance](https://docs.microsoft.com/en-us/azure/mysql/quickstart-create-mysql-server-database-using-azure-portal) using Azure Portal. Alternatively, you could also use Azure CLI by following [steps](https://docs.microsoft.com/en-us/azure/mysql/quickstart-create-mysql-server-database-using-azure-cli) here. Take notes for ```Server Name```, ```Admin username```, and ```Password```. 

By default the server created is protected with a firewall and is not accessible publicly. Follow the [guidance](https://docs.microsoft.com/en-us/azure/mysql/quickstart-create-mysql-server-database-using-azure-portal#configure-a-server-level-firewall-rule) to allow database to be accessible from external IP addresses. Based on your configuration, you might also enable all IP addresses, and disable ```Enforce SSL connection```.

#### 2. Deploy Kubeflow to use Azure metadata overlay
Follow the [installation document for Azure AKS](https://www.kubeflow.org/docs/azure/deploy/install-kubeflow/) until the step to build and apply the ```CONFIG_URI```. Download your configuration file, so that you can customize the configuration before deploying Kubeflow by running ```wget -O kfctl_azure.yaml ${CONFIG_URI}```, where the ```${CONFIG_URL}``` should be the one you specified in the previous steps. Run
```kfctl build -V -f kfctl_azure.yaml```.

Edit the Azure stack at ```/stacks/azure``` and make change under ```resources``` from ```- ../../metadata/v3``` to ```metadata``` to use Azure MySQL.

Edit ```params.env``` to provide parameters to config map as follows (change the ```[db_name]``` to the server name you used):
```
MYSQL_HOST=[db_name].mysql.database.azure.com
MYSQL_DATABASE=mlmetadata
MYSQL_PORT=3306
MYSQL_ALLOW_EMPTY_PASSWORD=true
```

Edit ```secrets.env``` to create a secret based on your database configuration (make sure the user name follows the pattern with an "@", like the one showed below):
```
MYSQL_USERNAME=[admin_user_name]@[db_name]
MYSQL_PASSWORD=[admin_password]
```

#### 3. Run Kubeflow Installation
```kfctl apply -V -f kfctl_azure.yaml```
