package reqctx

import "context"

// OperationType represents GraphQL operation types.
type OperationType string

const (
	OperationQuery        OperationType = "query"
	OperationMutation     OperationType = "mutation"
	OperationSubscription OperationType = "subscription"
)

// GraphQLOperation holds the current GraphQL operation info.
type GraphQLOperation struct {
	// Name is the operation name from the GraphQL query.
	// May be empty for anonymous operations.
	Name string

	// Type is the operation type: query, mutation, or subscription.
	Type OperationType
}

// WithGraphQLOperation stores GraphQL operation info in the context.
func WithGraphQLOperation(ctx context.Context, op *GraphQLOperation) context.Context {
	return context.WithValue(ctx, keyGraphQL, op)
}

// GraphQLOperationFromContext retrieves GraphQL operation info from the context.
// Returns nil, false if not set (not a GraphQL request).
func GraphQLOperationFromContext(ctx context.Context) (*GraphQLOperation, bool) {
	v := ctx.Value(keyGraphQL)
	if v == nil {
		return nil, false
	}
	op, ok := v.(*GraphQLOperation)
	return op, ok
}

// MustGraphQLOperation retrieves GraphQL operation info from the context.
// Panics if not set.
func MustGraphQLOperation(ctx context.Context) *GraphQLOperation {
	op, ok := GraphQLOperationFromContext(ctx)
	if !ok || op == nil {
		panic("reqctx: GraphQLOperation not found in context")
	}
	return op
}

// OperationNameFromContext returns the GraphQL operation name.
// Returns empty string if not a GraphQL request or anonymous operation.
func OperationNameFromContext(ctx context.Context) string {
	op, ok := GraphQLOperationFromContext(ctx)
	if !ok || op == nil {
		return ""
	}
	return op.Name
}

// OperationTypeFromContext returns the GraphQL operation type.
// Returns empty string if not a GraphQL request.
func OperationTypeFromContext(ctx context.Context) OperationType {
	op, ok := GraphQLOperationFromContext(ctx)
	if !ok || op == nil {
		return ""
	}
	return op.Type
}
