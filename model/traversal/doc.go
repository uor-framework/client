/*
Copyright 2022 UOR-Framework Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package traversal

// This package defines methods for traversing through a tree of
// model.Node types. If a node type implements the model.Iterator interface,

// The Tracker types stores node relationship information for traversal.
// It also tracker which nodes have been visited and a node budget (if set).

// Walk performs a simple iterative DFS traversal over the tree. WalkNested will traverse a node and any
// and sub-nodes if iterable.
