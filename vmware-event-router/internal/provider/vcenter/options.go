package vcenter

// Option allows for customization of the vCenter event provider
// TODO: change signature to return errors
type Option func(*EventStream)
