Shipshape
=========
Shipshape is a set of tools for integrating Docker and etcd. It is currently under development.

Proposed Features
-----------------
The following features are proposed:

- Service discovery.
- Service monitoring.
- Configuration templating.

Components
----------
Shipshape will consist of two major components: Deckhand which runs in each Docker instance and the as-yet unnamed host daemon.

The host daemon is responsible for integration with Docker, etcd, and the Deckhand. It announces services to etcd on behalf of Deckhand and forwards service changes back to Deckhand.

Deckhand runs in each container and is responsible for configuring, running, and monitoring services inside of the container.

License
-------
Shipshape is licensed under GPLv3. See LICENSE for a copy of the license.
