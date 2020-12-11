package vcsim

// Option allows for customization of the vCenter simulator event provider
// TODO: return errors
type Option func(*EventStream)
