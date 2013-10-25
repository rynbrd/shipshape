Shipshape
=========
Shipshape is a set of tools for integrating Docker, etcd, and Supervisor. It is currently under development. Until the first release this document will serve as a placeholder for proposed features and design decisions.

Proposed Features
-----------------
The following features are proposed:

- Service discovery.
- Service monitoring.
- Configuration templating.

Components
----------
Shipshape will consist of three major components: the bosun, the oiler, and the greaser.

The bosun runs on the host and is responsible for interation with Docker, etcd, and the greaser. The bosun announces services on behalf of the greaser instance running in a container and forwards service changes to those containers.

The greaser runs in each container and is responsible for generating configuration, monitoring services running in Supervisor via the oiler, and announcing services and handling events to/from the bosun.

The oiler is a Supervisor event listener which provides service status change events to the greaser.

License
-------
Shipshape is licensed under GPLv2. See LICENSE for a copy of the licensea copy of the license.
