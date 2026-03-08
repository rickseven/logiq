package domain

// LogStream is a generic interface for reading streaming logs line by line
type LogStream <-chan string
