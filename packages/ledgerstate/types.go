package ledgerstate

// empty structs consume 0 memory and are therefore better in map based set implementations that i.e. bool
type empty struct{}

// void represents an empty value that consumed 0 memory
var void empty
