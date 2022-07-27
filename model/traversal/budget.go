package traversal

import (
	"fmt"

	"github.com/uor-framework/uor-client-go/model"
)

// Budget tracks budgeted operations during graph traversal.
type Budget struct {
	// Maximum numbers of nodes to visit in a single traversal operator before stopping.
	NodeBudget int64
}

// ErrBudgetExceeded is an error that described the event where
// the maximum amount of nodes have been visited with no match.
type ErrBudgetExceeded struct {
	Node model.Node
}

func (e *ErrBudgetExceeded) Error() string {
	return fmt.Sprintf("traversal budget exceeded: node budget for reached zero while on node %v", e.Node.Address())
}
