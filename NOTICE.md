# NexaCard Secondary Development Notice

NexaCard API is an independently maintained secondary-development branch of the open-source Dujiao-Next project.

This repository is branded, released, and updated by the NexaCard project. Runtime update checks, release packages, Docker images, and deployment documentation must use NexaCard-owned repositories and release channels only:

- API releases: https://github.com/NexaCard/API/releases
- Admin frontend: https://github.com/NexaCard/admin
- User frontend: https://github.com/NexaCard/user
- Documentation: https://github.com/NexaCard/docs

Compatibility note: some API integration protocol names and request headers may still contain `Dujiao-Next-*` for wire-level compatibility with existing integrations. These names are protocol identifiers, not update sources.
