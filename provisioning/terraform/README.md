# Creating a Nomad cluster on AWS to run OpenFaaS
This folder contains Terraform configuration for creating a simple Nomad and Consul cluster on AWS which you can use to run OpenFaaS.  By default this will create a single server and three worker nodes with the size `t2.medium`.  This can be changed by editing the file `terraform.tf`.

## Set your environment variables
```
export AWS_ACCESS_KEY_ID="xxxxxxxxxxxxxxxxxxxxxx"
export AWS_SECRET_ACCESS_KEY="xxxxxxxxxxxxxxxxxxxxxx"
export AWS_DEFAULT_REGION="eu-west-1"
```

## Run terraform init to fetch modules
```
terraform init
```

## Run terraform plan to inspect changes
```
terraform plan
```

You
should see some output something like the following
```
# ...
      enable_monitoring:                     "true"
      iam_instance_profile:                  "openfaas-server-nomad-consul-join"
      image_id:                              "ami-add175d4"
      instance_type:                         "t2.medium"
      key_name:                              "${var.key_name}"
      name:                                  "openfaas-server.server"
      root_block_device.#:                   <computed>
      security_groups.#:                     <computed>
      user_data:                             "12cd96318ce75c053ebc74eec32f2d671e096a66"


Plan: 30 to add, 0 to change, 0 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.
```

## Run terraform apply to create infrastructure
```
terraform apply
```

The output should look something like the following
```
Apply complete! Resources: 30 added, 0 changed, 0 destroyed.

Outputs:

grafana_endpoint = http://openfaas-openfaas-393791592.eu-west-1.elb.amazonaws.com:3000/
nomad_endpoint = http://openfaas-nomad-1620139897.eu-west-1.elb.amazonaws.com:4646/
openfaas_endpoint = http://openfaas-openfaas-393791592.eu-west-1.elb.amazonaws.com:8080/
prometheus_endpoint = http://openfaas-openfaas-393791592.eu-west-1.elb.amazonaws.com:9090/
```

## Now run the Nomad config to run the base OpenFaaS applications
```
$ NOMAD_ADDR=$(terraform output nomad_endpoint) nomad run ../nomad_job_files/faas.hcl

==> Monitoring evaluation "a3e54faa"
    Evaluation triggered by job "faas-nomadd"
    Allocation "28f60a54" created: node "867c6baa", group "faas-nomadd"
    Allocation "7223b65d" created: node "d196a533", group "faas-nomadd"
    Allocation "a4dbae6c" created: node "123e18c0", group "faas-nomadd"
    Evaluation status changed: "pending" -> "complete"
==> Evaluation "a3e54faa" finished with status "complete"
```

## To add monitoring with Prometheus and Grafana run the monitoring job
```
$ NOMAD_ADDR=$(terraform output nomad_endpoint) nomad run ../nomad_job_files/monitoring.hcl

==> Monitoring evaluation "7d9c46df"
    Evaluation triggered by job "faas-monitoring"
    Allocation "e20ace08" created: node "123e18c0", group "faas-monitoring"
    Evaluation status changed: "pending" -> "complete"
==> Evaluation "7d9c46df" finished with status "complete"
Apply complete! Resources: 30 added, 0 changed, 0 destroyed.
```

You can now interact with OpenFaaS using the ALB which has been created, to see those endpoints at any time you can run the following command from this directory.
```
terraform output
```

Please see the main README in the root of this repository for information on running and creating jobs with OpenFaaS.
