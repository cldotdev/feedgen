package feedgen

import (
	"fmt"
)

// ItemFetchError shows that the item can't be fetched from the source URL.
type ItemFetchError struct {
	SourceURL string
}

// Error returns error message
func (e ItemFetchError) Error() string {
	return fmt.Sprintf("items can't be fetched from the source URL: %s", e.SourceURL)
}

// PageContentNotFoundError shows that the page content is not available from the source URL.
type PageContentNotFoundError struct {
	SourceURL string
}

// Error returns error message
func (e PageContentNotFoundError) Error() string {
	return fmt.Sprintf("page content is not found from the source URL: %s.", e.SourceURL)
}

// ParameterValueInvalidError shows that the query parameter Parameter's value is not available yet.
type ParameterValueInvalidError struct {
	Parameter string
}

// Error returns error message
func (e ParameterValueInvalidError) Error() string {
	return fmt.Sprintf("parameter %s has invalid value", e.Parameter)
}

// ParameterNotFoundError shows that the parameter Parameter is required.
type ParameterNotFoundError struct {
	Parameter string
}

// Error returns error message
func (e ParameterNotFoundError) Error() string {
	return fmt.Sprintf("parameter %s is required", e.Parameter)
}
