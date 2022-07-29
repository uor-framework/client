Design: Model
===
- [Design: Model](#design-model)
- [Interface types defined in the model package.](#interface-types-defined-in-the-model-package)
- [Node](#node)
- [NodeBuilder](#nodebuilder)
- [Iterator](#iterator)
- [Matcher](#matcher)
- [Attributes](#attributes)


The model package and the sub-package contains all types and methods that can be used to define and work with UOR data.

Package and sub-package layout 
```
model/
├── doc.go
├── nodes
│   ├── basic
│   │   └── basic.go
│   ├── collection
│   │   ├── assembly.go
│   │   ├── builder
│   │   │   └── builder.go
│   │   ├── collection.go
│   │   ├── collection_test.go
│   │   ├── doc.go
│   │   ├── edge.go
│   │   ├── iterator.go
│   │   └── iterator_test.go
│   ├── descriptor
│   │   ├── descriptor.go
│   │   └── descriptor_test.go
│   └── doc.go
└── types.go
```

> For more information on the concrete node types, see [nodes](nodes.md).

# Interface types defined in the model package.
# Node
Node is an interface that is used to represent different types of self-describing addressable data (typically stored in files locally or remotely). The methods defined in this interface are intended to be read-only. For methods that manipulate or assemble nodes, see NodeBuilder.

# NodeBuilder
NodeBuilder is an interface that defines methods for building immutable Node types. 

# Iterator
Nodes can be iterable (e.g. a UOR Collection). Using the iterator interface allows these structures to be iterated over during traversal.

# Matcher
Matcher is an interface that defines methods for node searching/matching that can guide Node graph traversal.

# Attributes

Attributes is an interface that represents the methods that would be used by a structure containing a set of Node attributes.


