Design: Node Types
===
- [Design: Node Types](#design-node-types)
- [Type Requirements](#type-requirements)
- [Basic Node](#basic-node)
- [Descriptor Node](#descriptor-node)
- [Collection Node](#collection-node)
  - [Linked Collection](#linked-collection)
      - [Why not use the manifest referrers-api?](#why-not-use-the-manifest-referrers-api)

# Type Requirements

The model.Node types must implement the following methods:

- ID: This is a string ID that **MUST** be unique with the implemented graph are sub-graph.
- Address: This represent the location of the node whether it be a relative or absolute path on disk or the registry path for a remote location.
- Attributes: This is a collection of key, values pair that represent information about the node. Example for a file could be name, size, and file type.

# Basic Node

A basic node represent the smallest unit of information in a UOR dataset. This can be part of a structure alongside Collection node
types or references within a Collection node.

# Descriptor Node

A descriptor node represent an OCI descriptor in a UOR dataset. This can be part of a structure alongside Collection node
types or references within a Collection node. When using a UOR Collection to describe an OCI DAG, descriptor nodes are useful within 
collection nodes.

# Collection Node

A Collection node implements a node of nodes pattern. A collection node implements the model.Node interface itself so its relationship to other nodes can be described.

Quick information about the Collection node:
- A Collection arranges nodes in a directed acyclic graph (may or may not be rooted).
- A Collection can be arranged in a tree structure if a Collection stores Collection nodes.
- When the Attributes method is called on a Collection, the root node (if applicable) attributes are returned.

The reference implementation use case for the Collection node is the following:

A Collection node represents an OCI artifact DAG and a structure of linked collection.

A Collection node can represent a Collection as an OCI artifact and descriptor nodes that refer to the artifact.

The Collection node represent a workspace with files that contained links to other files within the workspace. In order to render these
files in a way that would be consumable in a registry, the relationship between the files must be tracked. This is the way workspaces are created
in the builder.Compatibility mode.

## Linked Collection

The reference implementation has the concept of linked collections. This is a way to use the UOR model to arrange collections of OCI artifacts and any referring artifacts into a structure that can be traversed for various tooling and used by other parts of the model (e.g. Matcher). Linked collection information is stored within OCI manifest top-level annotation for collection building and retrieval.

**Important Annotations**

```
# Address of the default collection schema
uor.schema
```
```
# Address of default schema in linked collections.
# Can be used for validation or cross-link traversal.
uor.schema.linked
```

```
# Address of linked collections
uor.collections.linked
```

#### Why not use the manifest referrers-api?
[Info here](https://github.com/oras-project/artifacts-spec/blob/main/manifest-referrers-api.md)

Collections can refer to other collections, but this linkage does not fit into the scope of the referrers API because these
references must be mutable. They can be coupled with existing collection, but there is not a one-to-many relationship
between collections.