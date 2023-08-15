# Releases

GoToSocial can be installed in a number of different ways. We publish official binary releases as well as container images. A number of third-party packages are maintained by different distributions and some people have created additional deployment tooling to make it easy to deploy GoToSocial yourself.

## Binary releases

We publish binary builds for Linux to [our GitHub project](https://github.com/superseriousbusiness/gotosocial/releases):

* 32-bit Intel/AMD (i386/x86)
* 64-bit Intel/AMD (amd64/x86_64)
* 32-bit ARM (v6 and v7)
* 64-bit ARM64

For FreeBSD we publish:

* 64-bit Intel/AMD (amd64/x86_64)

## Containers

We also publish container images [on the Docker Hub](https://hub.docker.com/r/superseriousbusiness/gotosocial).

Containers are released for the same Linux platforms as our binary releases, with the exception of 32-bit Intel/AMD.

## Snapshots

We publish snapshot binary builds and Docker images of whatever is currently on main.

We always recommend using a stable release instead, but if you want to live on the edge (at your own risk!) then see the [snapshots](https://github.com/superseriousbusiness/gotosocial#snapshots) section on our GitHub repo for more information.

## Third-party

Some folks have created distribution packages for GoToSocial or additional tooling to aid in installing GoToSocial.

### Distribution packages

These packages are not maintained by GoToSocial, so please direct questions and issues to the repository maintainers (and donate to them!).

[![Packaging status](https://repology.org/badge/vertical-allrepos/gotosocial.svg)](https://repology.org/project/gotosocial/versions)

### Deployment tools

You can deploy your own instance of GoToSocial with the help of:

- [YunoHost GoToSocial Packaging](https://github.com/YunoHost-Apps/gotosocial_ynh) by [OniriCorpe](https://github.com/OniriCorpe).
- [Ansible Playbook (MASH)](https://github.com/mother-of-all-self-hosting/mash-playbook): The playbook supports a many services, including GoToSocial. [Documentation](https://github.com/mother-of-all-self-hosting/mash-playbook/blob/main/docs/services/gotosocial.md)
- GoToSocial Helm Charts:
  - [GoToSocial Helm Chart](https://github.com/fSocietySocial/charts/tree/main/charts/gotosocial) by [0hlov3](https://github.com/0hlov3).
