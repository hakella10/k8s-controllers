# k8s-controllers

<b>Challenges:</b>

 It is always a typical need to enhance the POD spec before deploying to K8S cluster. 
 
 Consider the below scenarios:
 
 1) APP Logging: Capture logs written to ephemeral file system within the POD. Need to push the logs to a central logging system. for e.g, ELK or Splunk
 
 2) APP Setup: Pull config files and/or certs from remote storage specific to our need and make them available for the application container to start via Init containers. 
 
 3) APM Monitoring: Add APM monitoring agent running a daemon to collect the APP runtime metrics and push to a central monitoring system. for e.g: Dynatrace
 
 4) Annotations: Add metadata like versions, tracking info etc. Especially, when using CICD tools to build and deploy the PODs.
 
 5) Under-the-Hood: Add respective annotations to leverage underlying cloud platform. For eg: Adding ALB,NLB specific annotations to services when deploying to AWS cloud. 

 
Normally, a Developer or a DevOps engineer modifies the POD spec to add these common utilities before deploying to K8S. 
What if, these utilities have their own lifecycle management? 
It is cumbersome to keep track of these changes and latest versions to be added to the spec and prone to manual errors.

To manage these automatically and with minimal overhead, we can implement a bit of automation by extending the K8S APIs with custom logic. With additional features like versioning the utilities, targetting specific applications - grouped by namespaces, tages by labels, tracked using annotations. 

<u>How to Solve:</u>

This is where K8S provide us with extension hooks and custom resources. 

![image](https://user-images.githubusercontent.com/72021023/129070246-407f16fa-e5cf-47ed-8621-533770fd402d.png)

