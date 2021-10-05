package webhook

// Option allows for customization of the webhook event provider
// TODO: return errors
type Option func(s *Server)
