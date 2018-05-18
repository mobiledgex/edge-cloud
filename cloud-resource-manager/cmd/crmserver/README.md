# Cloud Resource Manager server

CRM server connects to the Controller and registers Cloudlets.

It also collects Cloudlet resources and information.

Its own API endpoint allows adding resources from Operator's point of view. And
also allows Controller and others to submit application deployments.

It will also interact with CRM agents to control Cloudlets and collect data.

CRM will support Kubernetes and Openstack initially and other platforms as needed.
