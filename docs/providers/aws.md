# AWS UniK Provider

UniK supports running rumprun unikernels on AWS.
The AWS stub of your `daemon-config.yaml` file should look something like the following:
```yaml
providers:
  #...
  aws:
    - name: aws-1
      region: us-west-1
      zone: us-west-1a
```
UniK requires that your AWS credentials are set via [default AWS environment variables](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-environment) or your [AWS config file](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files).

UniK stores AWS data in the following paths:
* JSON representation of the state: `$HOME/.unik/aws/state.json`

* UniK boot volumes are stored as AMIs
* UniK data volumes are stored as EBS Backed Volumes
* UniK instances are `m1.small` EC2 Instances

If UniK gets into a bad state (i.e. you manually remove a file or AWS VM), you should manually edit the `$HOME/.unik/aws/state.json` file to remove the instance that no longer exists. UniK will eventually become self-correcting to deal with disruptions in the state.
