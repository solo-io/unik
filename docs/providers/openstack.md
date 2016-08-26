# OpenStack UniK Provider

**DISCLAIMER: OpenStack provider is under active development, please don't use it in production yet. **

UniK supports running OSv unikernels on OpenStack.
The OpenStack stub of your `daemon-config.yaml` file should look something like the following:
```yaml
providers:
  #...
  openstack:
    - name: openstack-1
      username: myusername
      password: mypassword
      auth_url: http://12.23.34.45:5000/v2.0
      tenant_id: 3dfe7bf545ff4885a3912a92a4a5f8e0
      tenant_name: admin
      project_name: admin
      region: RegionOne
```
You can override any of the settings above using environment variables, e.g.
```bash
$ export OS_USERNAME=realusername
$ export OS_PASSWORD=realpassword
```
UniK suggests that your OpenStack credentials are set via [default OpenStack environment variables](http://docs.openstack.org/user-guide/common/cli-set-environment-variables-using-openstack-rc.html), however, this is not asserted.

## About Flavors
UniK picks the most suitable flavor to run instance with, where most suitable means:
- as small as possible to fit logical image to it
- as little RAM as possible but not less than specified

## Misc
UniK stores OpenStack data in the following paths:
* JSON representation of the state: `$HOME/.unik/openstack/state.json`

If UniK gets into a bad state (i.e. you manually remove a state file or OpenStack VM), you should manually edit the `$HOME/.unik/openstack/state.json` file to remove the instance that no longer exists. UniK will eventually become self-correcting to deal with disruptions in the state.
