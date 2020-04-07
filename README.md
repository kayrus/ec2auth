# ec2auth

Authenticates an OpenStack user using EC2 credentials and returns a generated token ID.

## Example

Generate an EC2 credential using a OpenStack CLI:

```sh
$ openstack ec2 credentials create
```

Then authenticate against OpenStack Keystone using `ec2auth` CLI and credentials generated above:

```sh
$ ec2auth --access 7522162ced8f4e3eb9502168ef199584 --secret c558d9401a6943bbbb77a83ce910e5a5
```

The output is an OpenStack token ready to be used with an OpenStack CLI or `curl`.
